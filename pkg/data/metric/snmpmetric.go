package metric

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/measurement/filter"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"math"
	"regexp"
	"strconv"
	"time"
)

var (
	confDir string              //Needed to get File Filters data
	dbc     *config.DatabaseCfg //Needed to get Custom Filter  data
)

// SetConfDir  enable load File Filters from anywhere in the our FS.
func SetConfDir(dir string) {
	confDir = dir
}

// SetDB load database config to load data if needed (used in filters)
func SetDB(db *config.DatabaseCfg) {
	dbc = db
}

const (
	NeverReport     = 0
	AlwaysReport    = 1
	OnNonZeroReport = 2
	OnChangedReport = 3
)

//SnmpMetric type to metric runtime
type SnmpMetric struct {
	cfg         *config.SnmpMetricCfg
	ID          string
	CookedValue interface{}
	CurValue    interface{}
	LastValue   interface{}
	CurTime     time.Time
	LastTime    time.Time
	ElapsedTime float64
	Compute     func(arg ...interface{})                `json:"-"`
	Scale       func()                                  `json:"-"`
	SetRawData  func(pdu gosnmp.SnmpPDU, now time.Time) `json:"-"`
	RealOID     string
	Report      int //if false this metric won't be sent to the ouput buffer (is just taken as a coomputed input for other metrics)
	//for STRINGPARSER
	re   *regexp.Regexp
	expr *govaluate.EvaluableExpression
	//for CONDITIONEVAL
	condflt filter.Filter
	// Logger
	log *logrus.Logger
}

// GetDataSrcType get needed data
func (s *SnmpMetric) GetDataSrcType() string {
	return s.cfg.DataSrcType
}

// PrintDebugCfg helps users get data about metric configuration
func (s *SnmpMetric) PrintDebugCfg() {
	s.log.Debugf("DEBUG METRIC  CONFIG %+v", s.cfg)
}

// IsTag needed to generate Influx measurements
func (s *SnmpMetric) IsTag() bool {
	return s.cfg.IsTag
}

// GetFieldName  needed to generate Influx measurements
func (s *SnmpMetric) GetFieldName() string {
	return s.cfg.FieldName
}

// New constructor
func New(c *config.SnmpMetricCfg) (*SnmpMetric, error) {
	metric := &SnmpMetric{}
	err := metric.Init(c)
	return metric, err
}

func NewWithLog(c *config.SnmpMetricCfg, l *logrus.Logger) (*SnmpMetric, error) {
	metric := &SnmpMetric{log: l}
	err := metric.Init(c)
	return metric, err
}

func (s *SnmpMetric) SetLogger(l *logrus.Logger) {
	s.log = l
}

func (s *SnmpMetric) Init(c *config.SnmpMetricCfg) error {
	if c == nil {
		return fmt.Errorf("Error on initialice device, configuration struct is nil")
	}
	s.cfg = c
	s.RealOID = c.BaseOID
	s.ID = s.cfg.ID
	if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
		s.Scale = func() {
			s.CookedValue = (s.cfg.Scale * float64(s.CookedValue.(float64))) + s.cfg.Shift
		}
	} else {
		s.Scale = func() {
		}
	}
	switch s.cfg.DataSrcType {
	case "CONDITIONEVAL":
		//select
		cond, err := dbc.GetOidConditionCfgByID(s.cfg.ExtraData)
		if err != nil {
			s.log.Errorf("Error getting CONDITIONEVAL [id: %s ] data : %s", s.cfg.ExtraData, err)
		}
		//get Regexp
		s.condflt = filter.NewOidFilter(cond.OIDCond, cond.CondType, cond.CondValue, s.log)

		s.Compute = func(arg ...interface{}) {
			//walk := arg[0].(func(string, gosnmp.WalkFunc) error)
			//err := s.condflt.Init(walk)
			s.condflt.Init(arg...)
			s.condflt.Update()
			s.CookedValue = s.condflt.Count()
			s.CurTime = time.Now()
			s.Scale()
		}
		//Sign
		//set Process Data
	case "TIMETICKS": //Cooked TimeTicks
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := snmp.PduVal2Int64(pdu)
			s.CookedValue = float64(val / 100) //now data in secoonds
			s.CurTime = now
			s.Scale()
		}

		//Signed Integers
	case "INTEGER", "Integer32":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := snmp.PduVal2Int64(pdu)
			s.CookedValue = float64(val)
			s.CurTime = now
			s.Scale()
		}
		//Unsigned Integers
	case "Counter32", "Gauge32", "Counter64", "TimeTicks", "UInteger32", "Unsigned32":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := snmp.PduVal2UInt64(pdu)
			s.CookedValue = float64(val)
			s.CurTime = now
			s.Scale()
		}
	case "COUNTER32": //Increment computed
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			//first time only set values and reassign itself to the complete method this will avoi to send invalid data
			val := snmp.PduVal2UInt64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				val := snmp.PduVal2UInt64(pdu)
				s.LastTime = s.CurTime
				s.LastValue = s.CurValue
				s.CurValue = val
				s.CurTime = now
				s.Compute()
				s.Scale()
			}
		}
		if s.cfg.GetRate == true {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = float64(math.MaxInt32-s.LastValue.(uint64)+s.CurValue.(uint64)) / s.ElapsedTime
				} else {
					s.CookedValue = float64(s.CurValue.(uint64)-s.LastValue.(uint64)) / s.ElapsedTime
				}
			}
		} else {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = float64(math.MaxInt32 - s.LastValue.(uint64) + s.CurValue.(uint64))
				} else {
					s.CookedValue = float64(s.CurValue.(uint64) - s.LastValue.(uint64))
				}
			}
		}
	case "COUNTER64": //Increment computed
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			//log.Debugf("========================================>COUNTER64: first time :%s ", s.RealOID)
			//first time only set values and reassign itself to the complete method
			val := snmp.PduVal2UInt64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				//log.Debugf("========================================>COUNTER64: the other time:%s", s.RealOID)
				val := snmp.PduVal2UInt64(pdu)
				s.LastTime = s.CurTime
				s.LastValue = s.CurValue
				s.CurValue = val
				s.CurTime = now
				s.Compute()
				s.Scale()
			}
		}
		if s.cfg.GetRate == true {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				//duration := s.CurTime.Sub(s.LastTime)
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = float64(math.MaxInt64-s.LastValue.(uint64)+s.CurValue.(uint64)) / s.ElapsedTime
				} else {
					s.CookedValue = float64(s.CurValue.(uint64)-s.LastValue.(uint64)) / s.ElapsedTime
				}
			}
		} else {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = float64(math.MaxInt64 - s.LastValue.(uint64) + s.CurValue.(uint64))
				} else {
					s.CookedValue = float64(s.CurValue.(uint64) - s.LastValue.(uint64))
				}
			}

		}
	case "OCTETSTRING":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue = snmp.PduVal2str(pdu)
			s.CurTime = now
		}
	case "IpAddress":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = snmp.PduVal2IPaddr(pdu)
			s.CurTime = now
		}
	case "HWADDR":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = snmp.PduVal2Hwaddr(pdu)
			s.CurTime = now
		}
	case "STRINGPARSER":
		//get Regexp
		re, err := regexp.Compile(s.cfg.ExtraData)
		if err != nil {
			return fmt.Errorf("Error on initialice STRINGPARSER, invalind Regular Expression : %s", s.cfg.ExtraData)
		}
		s.re = re
		//set Process Data
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			str := snmp.PduVal2str(pdu)
			retarray := s.re.FindStringSubmatch(str)
			if len(retarray) < 2 {
				s.log.Warnf("Error for metric [%s] parsing REGEXG [%s] on string [%s] without capturing group", s.cfg.ID, s.cfg.ExtraData, str)
				return
			}
			//retarray[0] contains full string
			if len(retarray[1]) == 0 {
				s.log.Warnf("Error for metric [%s] parsing REGEXG [%s] on string [%s] cause  void capturing group", s.cfg.ID, s.cfg.ExtraData, str)
				return
			}
			value, err := strconv.ParseFloat(retarray[1], 64)
			if err != nil {
				s.log.Warnf("Error parsing float for metric %s : error: %s", s.cfg.ID, err)
				return
			}
			s.CookedValue = value
			s.CurTime = now
			s.Scale()
		}
	case "STRINGEVAL":

		expression, err := govaluate.NewEvaluableExpression(s.cfg.ExtraData)
		if err != nil {
			s.log.Errorf("Error on initialice STRINGEVAL, evaluation : %s : ERROR : %s", s.cfg.ExtraData, err)
			return err
		}
		s.expr = expression
		//set Process Data
		s.Compute = func(arg ...interface{}) {
			//parameters := make(map[string]interface{})
			parameters := arg[0].(map[string]interface{})
			result, err := s.expr.Evaluate(parameters)
			if err != nil {
				s.log.Errorf("Error in metric %s On EVAL string: %s : ERROR : %s", s.cfg.ID, s.cfg.ExtraData, err)
				return
			}
			//Influxdb has not support for NaN,Inf values
			//https://github.com/influxdata/influxdb/issues/4089
			switch v := result.(type) {
			case float64:
				if math.IsNaN(v) || math.IsInf(v, 0) {
					s.log.Warnf("Warning in metric %s On EVAL string: %s : Value is not a valid Floating Pint (NaN/Inf) : %f", s.cfg.ID, s.cfg.ExtraData, v)
					return
				}
			}
			s.CookedValue = result
			s.CurTime = time.Now()
			s.Scale()
		}
	}
	return nil
}
