package metric

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/gosnmp/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/filter"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

var (
	confDir string              // Needed to get File Filters data
	dbc     *config.DatabaseCfg // Needed to get Custom Filter  data
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
	// NeverReport  default value for metric resport flag
	NeverReport = 0
	// AlwaysReport  metric will send data allways
	AlwaysReport = 1
	// OnNonZeroReport metric will send data only if computed value different than 0
	OnNonZeroReport = 2
	// OnChangedReport metric will send data only if data has a change from last changed value ( not implemente yet)
	OnChangedReport = 3
)

// SnmpMetric type to metric runtime
type SnmpMetric struct {
	cfg         *config.SnmpMetricCfg
	Valid       bool // indicate if has been updated in the last gathered process
	CookedValue interface{}
	CurValue    interface{}
	LastValue   interface{}
	CurTime     time.Time
	LastTime    time.Time
	ElapsedTime float64
	Compute     func(arg ...interface{})                `json:"-"`
	Scale       func()                                  `json:"-"`
	Convert     func()                                  `json:"-"`
	SetRawData  func(pdu gosnmp.SnmpPDU, now time.Time) `json:"-"`
	RealOID     string
	Report      int // if false this metric won't be sent to the output buffer (is just taken as a coomputed input for other metrics)
	// for STRINGPARSER/MULTISTRINGPARSER
	re   *regexp.Regexp
	mm   []*config.MetricMultiMap
	expr *govaluate.EvaluableExpression
	// for CONDITIONEVAL
	condflt filter.Filter
	// Logger
	log utils.Logger
}

// GetDataSrcType get needed data
func (s *SnmpMetric) GetDataSrcType() string {
	return s.cfg.DataSrcType
}

// IsTag needed to generate Influx measurements
func (s *SnmpMetric) IsTag() bool {
	return s.cfg.IsTag
}

// GetFieldName  needed to generate Influx measurements
func (s *SnmpMetric) GetFieldName() string {
	return s.cfg.FieldName
}

// GetID  needed to generate Influx measurements
func (s *SnmpMetric) GetID() string {
	return s.cfg.ID
}

// New create a new snmpmetric with a specific logger
func New(c *config.SnmpMetricCfg, l utils.Logger) (*SnmpMetric, error) {
	metric := &SnmpMetric{log: l}
	err := metric.Init(c)
	return metric, err
}

// SetLogger attach logger to the current snmpmetric object
func (s *SnmpMetric) SetLogger(l utils.Logger) {
	s.log = l
}

// Conversion functions

func (s *SnmpMetric) convertFromUInteger() {
	// check first the rigth
	switch vt := s.CookedValue.(type) {
	case uint64, uint:
		// everything ok
		break
	default:
		s.log.Errorf("ERROR: expected value on metric %s type UINT64 and got %T  type ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
		return
	}
	// the only acceptable conversions
	// signed integer 64 -> float64
	// signet integer 64 -> boolean ( true if value != 0 )
	switch s.cfg.Conversion {
	case config.INTEGER:
		s.CookedValue = int64(s.CookedValue.(uint64))
		return
	case config.FLOAT:
		s.CookedValue = float64(s.CookedValue.(uint64))
		return
	case config.BOOLEAN:
		if s.CookedValue.(uint64) != 0 {
			s.CookedValue = true
		} else {
			s.CookedValue = false
		}
		return
	case config.STRING:
		s.CookedValue = strconv.FormatUint(s.CookedValue.(uint64), 10)
		return
	default:
		s.log.Errorf("Bad conversion: requested %s from %T type", s.cfg.Conversion.GetString(), s.CookedValue)
	}
}

func (s *SnmpMetric) convertFromInteger() {
	switch vt := s.CookedValue.(type) {
	case int64, int:
		// everything ok
		break
	default:
		s.log.Errorf("ERROR: expected value on metric %s type INT64 and got %T ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
		return
	}
	// the only acceptable conversions
	// signed integer 64 -> float64
	// signet integer 64 -> boolean ( true if value != 0 )
	switch s.cfg.Conversion {
	case config.INTEGER:
		return
	case config.FLOAT:
		s.CookedValue = float64(s.CookedValue.(int64))
		return
	case config.BOOLEAN:
		if s.CookedValue.(int64) != 0 {
			s.CookedValue = true
		} else {
			s.CookedValue = false
		}
		return
	case config.STRING:
		s.CookedValue = strconv.FormatInt(s.CookedValue.(int64), 10)
		return
	default:
	}
}

func (s *SnmpMetric) convertFromFloat() {
	switch vt := s.CookedValue.(type) {
	case float64:
		// everything ok
		break
	default:
		s.log.Errorf("ERROR: expected value on metric %s type Float64 and got %T type ( %+v) \n", s.cfg.ID, vt, s.CookedValue)
		return
	}
	// the only acceptable conversions
	// signed float -> int64 (will do rounded value)
	switch s.cfg.Conversion {
	case config.INTEGER:
		s.CookedValue = int64(math.Round(s.CookedValue.(float64)))
		return
	case config.FLOAT:
		return
	case config.BOOLEAN:
		if s.CookedValue.(float64) != 0.0 {
			s.CookedValue = true
		} else {
			s.CookedValue = false
		}
		return
	case config.STRING:
		s.CookedValue = strconv.FormatFloat(s.CookedValue.(float64), 'f', -1, 64)
		return
	default:
		s.log.Errorf("Bad conversion: requested on metric %s: to type  %s from %T type", s.cfg.ID, s.cfg.Conversion.GetString(), s.CookedValue)
	}
}

func (s *SnmpMetric) convertFromString() {
	switch vt := s.CookedValue.(type) {
	case string:
		// everything ok
		break
	default:
		s.log.Errorf("ERROR: expected value on metric %s type STRING and got %T type ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
		return
	}
	// the only acceptable conversions
	// string -> int64
	// string -> float (the default)
	// string -> boolean
	// string -> string
	switch s.cfg.Conversion {
	case config.STRING:
		return
	case config.INTEGER:
		value, err := strconv.ParseInt(s.CookedValue.(string), 10, 64)
		if err != nil {
			s.log.Warnf("Error parsing Integer from String  %s metric %s : error: %s", s.CookedValue.(string), s.cfg.ID, err)
			return
		}
		s.CookedValue = value
		return
	case config.FLOAT:
		value, err := strconv.ParseFloat(s.CookedValue.(string), 64)
		if err != nil {
			s.log.Warnf("Error parsing float from String  %s metric %s : error: %s", s.CookedValue.(string), s.cfg.ID, err)
			return
		}
		s.CookedValue = value
		return
	case config.BOOLEAN:
		value, err := strconv.ParseBool(s.CookedValue.(string))
		if err != nil {
			s.log.Warnf("Error parsing Boolean from String  %s metric %s : error: %s", s.CookedValue.(string), s.cfg.ID, err)
			return
		}
		s.CookedValue = value
		return
	default:
		s.log.Errorf("Bad conversion: requested on metric %s: to type  %s from %T type", s.cfg.ID, s.cfg.Conversion.GetString(), s.CookedValue)
	}
}

func (s *SnmpMetric) convertFromAny() {
	switch s.CookedValue.(type) {
	case float32, float64:
		s.convertFromFloat()
		return
	case uint64, uint32:
		s.convertFromUInteger()
		return
	case int64, int32:
		s.convertFromInteger()
		return
	case string:
		s.convertFromString()
		return
	case bool:
		s.log.Errorf("Bad conversion: requested on metric %s: to type  %s from %T type", s.cfg.ID, s.cfg.Conversion.GetString(), s.CookedValue)
	default:
	}
}

// Init Initialice a new snmpmetric object with the specific configuration
func (s *SnmpMetric) Init(c *config.SnmpMetricCfg) error {
	if c == nil {
		return fmt.Errorf("Error on initialice device, configuration struct is nil")
	}
	s.cfg = c
	s.RealOID = c.BaseOID
	// set default conversion funcion
	s.Convert = s.convertFromAny
	if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
		s.Scale = func() {
			// always Scale shoud return float (this avoids precission lost)
			switch v := s.CookedValue.(type) {
			case uint64:
				s.CookedValue = (s.cfg.Scale * float64(s.CookedValue.(uint64))) + s.cfg.Shift
				// here change uint64 to float , with the apropiate type conversion at the end
				// this temporal format change avoids precission lost
			case int64:
				s.CookedValue = (s.cfg.Scale * float64(s.CookedValue.(int64))) + s.cfg.Shift
				// here change uint64 to float , with the apropiate type conversion at the end
				// this temporal format change avoids precission lost
			case float64:
				// should return float
				s.CookedValue = float64((s.cfg.Scale * float64(s.CookedValue.(float64))) + s.cfg.Shift)
			case string:
				s.log.Errorf("Error Trying to  Scale Function from non numbered STRING type value : %s ", s.CookedValue)
			default:
				s.log.Errorf("Error Trying to  Scale Function from unknown type %T value: %#+v", v, s.CookedValue)
			}
		}
	} else {
		s.Scale = func() {
		}
	}
	switch s.cfg.DataSrcType {
	case "CONDITIONEVAL":
		// select
		cond, err := dbc.GetOidConditionCfgByID(s.cfg.ExtraData)
		if err != nil {
			s.log.Errorf("Error getting CONDITIONEVAL [id: %s ] data : %s", s.cfg.ExtraData, err)
		}
		// get Regexp
		if cond.IsMultiple == true {
			s.condflt = filter.NewOidMultipleFilter(cond.OIDCond, s.log)
		} else {
			s.condflt = filter.NewOidFilter(cond.OIDCond, cond.CondType, cond.CondValue, s.log)
		}
		s.Compute = func(arg ...interface{}) {
			s.condflt.Init(arg...)
			s.condflt.Update()
			s.CookedValue = s.condflt.Count()
			s.CurTime = time.Now()
			s.Valid = true
		}
		// Sign
		// set Process Data
	case "TIMETICKS": // Cooked TimeTicks
		s.Convert = s.convertFromInteger

		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Convert = s.convertFromFloat
		}

		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := snmp.PduVal2Int64(pdu)
			s.CookedValue = val / 100 // now data in secoonds
			s.CurTime = now
			s.Scale()
			s.Convert()
			s.Valid = true
		}

		// Signed Integers
	case "INTEGER", "Integer32":
		s.Convert = s.convertFromInteger

		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Convert = s.convertFromFloat
		}

		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue = snmp.PduVal2Int64(pdu)
			s.CurTime = now
			s.Scale()
			s.Convert()
			s.Valid = true
		}
		// Unsigned Integers
	case "Counter32", "Gauge32", "Counter64", "TimeTicks", "UInteger32", "Unsigned32":
		s.Convert = s.convertFromUInteger

		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Convert = s.convertFromFloat
		}

		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue = snmp.PduVal2UInt64(pdu)
			s.CurTime = now
			s.Scale()
			s.Convert()
			s.Valid = true
		}
	case "COUNTER32": // Increment computed
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			// first time only set values and reassign itself to the complete method this will avoi to send invalid data
			val := snmp.PduVal2UInt64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.Valid = true
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				val := snmp.PduVal2UInt64(pdu)
				s.LastTime = s.CurTime
				s.LastValue = s.CurValue
				s.CurValue = val
				s.CurTime = now
				s.Compute()
				s.Scale()
				s.Convert()
				s.Valid = true
			}
		}
		if s.cfg.GetRate == true {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = float64(math.MaxUint32-s.LastValue.(uint64)+s.CurValue.(uint64)) / s.ElapsedTime
				} else {
					s.CookedValue = float64(s.CurValue.(uint64)-s.LastValue.(uint64)) / s.ElapsedTime
				}
			}
			s.Convert = s.convertFromFloat
		} else {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = math.MaxUint32 - s.LastValue.(uint64) + s.CurValue.(uint64)
				} else {
					s.CookedValue = s.CurValue.(uint64) - s.LastValue.(uint64)
				}
			}
			s.Convert = s.convertFromUInteger
		}

		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Convert = s.convertFromFloat
		}

	case "COUNTER64": // Increment computed
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			// log.Debugf("========================================>COUNTER64: first time :%s ", s.RealOID)
			// first time only set values and reassign itself to the complete method
			val := snmp.PduVal2UInt64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.Valid = true
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				// log.Debugf("========================================>COUNTER64: the other time:%s", s.RealOID)
				val := snmp.PduVal2UInt64(pdu)
				s.LastTime = s.CurTime
				s.LastValue = s.CurValue
				s.CurValue = val
				s.CurTime = now
				s.Compute()
				s.Scale()
				s.Convert()
				s.Valid = true
			}
		}
		if s.cfg.GetRate == true {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = float64(math.MaxUint64-s.LastValue.(uint64)+s.CurValue.(uint64)) / s.ElapsedTime
				} else {
					s.CookedValue = float64(s.CurValue.(uint64)-s.LastValue.(uint64)) / s.ElapsedTime
				}
			}
			s.Convert = s.convertFromFloat
		} else {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) < s.LastValue.(uint64) {
					s.CookedValue = math.MaxUint64 - s.LastValue.(uint64) + s.CurValue.(uint64)
				} else {
					s.CookedValue = s.CurValue.(uint64) - s.LastValue.(uint64)
				}
			}
			s.Convert = s.convertFromUInteger
		}

		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Convert = s.convertFromFloat
		}

	case "COUNTERXX": // Generic Counter With Unknown range or buggy counters that  Like Non negative derivative
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			// first time only set values and reassign itself to the complete method this will avoi to send invalid data
			val := snmp.PduVal2UInt64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.Valid = true
			s.log.Debugf("FIRST RAW(post-compute): %T - %#+v", s.CookedValue, s.CookedValue)
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				val := snmp.PduVal2UInt64(pdu)
				s.LastTime = s.CurTime
				s.LastValue = s.CurValue
				s.CurValue = val
				s.CurTime = now
				s.Compute()
				s.Valid = true
			}
		}
		if s.cfg.GetRate == true {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) >= s.LastValue.(uint64) {
					s.CookedValue = float64(s.CurValue.(uint64)-s.LastValue.(uint64)) / s.ElapsedTime
					s.Scale()
					s.Convert()
				} else {
					// Else => nothing to do last value will be sent
					s.log.Warnf("Warning Negative COUNTER increment [current: %d | last: %d ] last value will be sent %f", s.CurValue, s.LastValue, s.CookedValue)
				}
			}
			s.Convert = s.convertFromFloat
		} else {
			s.Compute = func(arg ...interface{}) {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue.(uint64) >= s.LastValue.(uint64) {
					s.CookedValue = s.CurValue.(uint64) - s.LastValue.(uint64)
					s.Scale()
					s.Convert()
				} else {
					// Else => nothing to do last value will be sent
					s.log.Warnf("Warning Negative COUNTER increment [current: %d | last: %d ] last value will be sent %f", s.CurValue, s.LastValue, s.CookedValue)
				}
			}
			s.Convert = s.convertFromUInteger
		}

		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Convert = s.convertFromFloat
		}

	case "BITS":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			barray := snmp.PduVal2BoolArray(pdu)
			names := []string{}
			for i, b := range barray {
				if b {
					names = append(names, s.cfg.Names[i])
				}
			}
			s.CookedValue = strings.Join(names, ",")
			s.CurTime = now
			s.Valid = true
			s.log.Debugf("SETRAW BITS %+v, RESULT %s", s.cfg.Names, s.CookedValue)
		}
	case "BITSCHK":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			barray := snmp.PduVal2BoolArray(pdu)
			index, _ := strconv.Atoi(s.cfg.ExtraData)
			b := barray[index]
			if b {
				s.CookedValue = 1.0
			} else {
				s.CookedValue = 0.0
			}
			s.Convert()
			s.CurTime = now
			s.Valid = true
			s.log.Debugf("BITS CHECK bit %+v, Position %d , RESULT %t", barray, index, s.CookedValue)
		}
	case "ENUM":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			idx := snmp.PduVal2Int64(pdu)
			if val, ok := s.cfg.Names[int(idx)]; ok {
				s.CookedValue = val
			} else {
				s.CookedValue = strconv.Itoa(int(idx))
			}
			s.Valid = true
			s.log.Debugf("SETRAW ENUM %+v, RESULT %s", s.cfg.Names, s.CookedValue)
		}
	case "OCTETSTRING":
		switch s.cfg.Conversion {

		case config.INTEGER:
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				val, err := snmp.PduValHexString2Uint(pdu)
				s.CookedValue = val
				s.CurTime = now
				if err != nil {
					s.log.Warnf("Error on HexString to UINT conversion: %s", err)
					return
				}
				s.Valid = true
			}
		// For compatibility purposes with previous versions
		case config.FLOAT:
			s.log.Errorf("WARNING ON SNMPMETRIC ( %s ): You are using version >=0.8 version without database upgrade: you should upgrade the DB by executing this SQL on your database \"update snmp_metric_cfg set Conversion=3 where datasrctype='OCTETSTRING';\", to avoid this message ", s.cfg.ID)
			fallthrough
		case config.STRING:
			s.Compute = func(arg ...interface{}) {
			}
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				s.CookedValue = snmp.PduVal2str(pdu)
				s.CurTime = now
				s.Compute()
				s.Valid = true
			}
			// check if trimming needed
			if len(s.cfg.ExtraData) > 0 {
				if s.cfg.ExtraData == "trimspace" {
					s.Compute = func(arg ...interface{}) {
						s.CookedValue = strings.TrimSpace(s.CookedValue.(string))
					}
					break
				}
				// https://regex101.com/r/55riOI/1/
				re := regexp.MustCompile(`([^\(]*)\(([^\)]*)\)$`)
				if !re.MatchString(s.cfg.ExtraData) {
					s.log.Errorf("Error on get Trim Config for OctecString with config %s", s.cfg.ExtraData)
					break
				}

				res := re.FindStringSubmatch(s.cfg.ExtraData)
				// s.log.Debugf("REGEXP: %+v", res)
				if len(res) < 3 {
					s.log.Errorf("Error on get Trim Config for OctecString got %v", res)
					break
				}
				function := res[1]
				args := strings.Trim(res[2], "'\"") // removed unneded quotes if arguments enclosed on "" or ''
				s.log.Debugf("Trim type [%s] with args [%s]", function, args)
				switch function {
				case "trimspace":
					s.Compute = func(arg ...interface{}) {
						s.CookedValue = strings.TrimSpace(s.CookedValue.(string))
					}
				case "trim":
					s.Compute = func(arg ...interface{}) {
						s.CookedValue = strings.Trim(s.CookedValue.(string), args)
					}
				case "trimleft":
					s.Compute = func(arg ...interface{}) {
						s.CookedValue = strings.TrimLeft(s.CookedValue.(string), args)
					}
				case "trimright":
					s.Compute = func(arg ...interface{}) {
						s.CookedValue = strings.TrimRight(s.CookedValue.(string), args)
					}
				}
			}

		default:
			s.log.Errorf("WARNING ON SNMPMETRIC ( %s ): Invalid conversion mode from OCTETSTRING to %s", s.cfg.ID, s.cfg.Conversion.GetString())
			s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				s.CookedValue = snmp.PduVal2str(pdu)
				s.CurTime = now
				s.Valid = true
			}
		}

	case "OID":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue = snmp.PduVal2OID(pdu)
			s.CurTime = now
			s.Valid = true
		}
	case "IpAddress":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = snmp.PduVal2IPaddr(pdu)
			s.CurTime = now
			s.Valid = true
		}
	case "HWADDR":
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = snmp.PduVal2Hwaddr(pdu)
			s.CurTime = now
			s.Valid = true
		}
	case "STRINGPARSER":
		// get Regexp
		re, err := regexp.Compile(s.cfg.ExtraData)
		if err != nil {
			return fmt.Errorf("Error on initialice STRINGPARSER, invalind Regular Expression : %s", s.cfg.ExtraData)
		}
		s.re = re
		// set Process Data
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			str := snmp.PduVal2str(pdu)
			retarray := s.re.FindStringSubmatch(str)
			if len(retarray) < 2 {
				s.log.Warnf("Error for metric [%s] parsing REGEXG [%s] on string [%s] without capturing group", s.cfg.ID, s.cfg.ExtraData, str)
				return
			}
			// retarray[0] contains full string
			if len(retarray[1]) == 0 {
				s.log.Warnf("Error for metric [%s] parsing REGEXG [%s] on string [%s] cause  void capturing group", s.cfg.ID, s.cfg.ExtraData, str)
				return
			}
			s.CookedValue = retarray[1]
			s.CurTime = now
			s.convertFromString()
			// s.Scale() <-only valid if Integer or Float
			s.Valid = true
		}
	case "MULTISTRINGPARSER":
		// get Regexp
		re, err := regexp.Compile(s.cfg.ExtraData)
		if err != nil {
			return fmt.Errorf("Error on initialice MULTISTRINGPARSER, invalind Regular Expression : %s", s.cfg.ExtraData)
		}
		s.re = re

		mm, err := s.cfg.GetMultiStringTagFieldMap()
		if err != nil {
			return fmt.Errorf("Error on initialice MULTISTRINGPARSER, invalind Field/Tag definition Format : %s", err)
		}
		s.mm = mm
		// set Process Data
		s.SetRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			str := snmp.PduVal2str(pdu)
			s.CookedValue = str
			s.CurTime = now
			s.Valid = true
		}
	case "STRINGEVAL":

		expression, err := govaluate.NewEvaluableExpression(s.cfg.ExtraData)
		if err != nil {
			s.log.Errorf("Error on initialice STRINGEVAL, evaluation : %s : ERROR : %s", s.cfg.ExtraData, err)
			return err
		}
		s.expr = expression
		// set Process Data
		s.Compute = func(arg ...interface{}) {
			parameters := arg[0].(map[string]interface{})
			s.log.Debugf("Evaluating Metric %s with eval expresion [%s] with parameters %+v", s.cfg.ID, s.cfg.ExtraData, parameters)
			result, err := s.expr.Evaluate(parameters)
			if err != nil {
				s.log.Errorf("Error in metric %s On EVAL string: %s : ERROR : %s", s.cfg.ID, s.cfg.ExtraData, err)
				return
			}
			// Influxdb has not support for NaN,Inf values
			// https://github.com/influxdata/influxdb/issues/4089
			switch v := result.(type) {
			case float64:
				if math.IsNaN(v) || math.IsInf(v, 0) {
					s.log.Warnf("Warning in metric %s On EVAL string: %s : Value is not a valid Floating Pint (NaN/Inf) : %f", s.cfg.ID, s.cfg.ExtraData, v)
					return
				}
			}
			s.CookedValue = result
			// conversion depends onthe type of the evaluted data.
			s.CurTime = time.Now()
			s.Scale()
			s.Convert() // default
			s.Valid = true
		}
	}
	return nil
}

// GetEvaluableVariables get all posible values to add to the
func (s *SnmpMetric) GetEvaluableVariables(params map[string]interface{}) {
	s.log.Debugf("Get Evaluable parameters for Metric %s", s.cfg.ID)
	switch s.cfg.DataSrcType {
	case "MULTISTRINGPARSER":
		tags := make(map[string]string)
		_ = s.addMultiStringParserValues(tags, params)
		for k, v := range tags {
			params[k] = v
		}
	default:
		if s.Valid == true { // only valid for compute if it has been updated last
			params[s.cfg.FieldName] = s.CookedValue
		}
	}
}

func (s *SnmpMetric) addSingleField(mid string, fields map[string]interface{}) int64 {
	if s.Report == OnNonZeroReport {
		if s.CookedValue == 0.0 {
			s.log.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", s.cfg.ID, mid)
			return 0
		}
	}
	// assuming float Cooked Values
	s.log.Debugf("generating field for %s value %#v ", s.cfg.FieldName, s.CookedValue)
	s.log.Debugf("DEBUG METRIC %+v", s)
	fields[s.cfg.FieldName] = s.CookedValue
	return 0
}

func (s *SnmpMetric) addSingleTag(mid string, tags map[string]string) int64 {
	var tag string
	switch v := s.CookedValue.(type) {
	case string:
		tag = v
	case uint32:
		tag = strconv.FormatUint(uint64(v), 10)
	case uint64:
		tag = strconv.FormatUint(v, 10)
	case int32:
		tag = strconv.FormatInt(int64(v), 10)
	case int64:
		tag = strconv.FormatInt(v, 10)
	case bool:
		tag = strconv.FormatBool(v)
	case float32:
		tag = strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		tag = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		s.log.Debugf("ERROR wrong type %T for ID [%s] from MEASUREMENT[ %s ] when converting to TAG(STRING) won't be reported to the output backend", v, s.cfg.ID, mid)
		return 1
	}
	// I don't know if a OnNonZeroReport could have sense in any configuration.
	if s.Report == OnNonZeroReport {
		if tag == "0" {
			s.log.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", s.cfg.ID, mid)
			return 0
		}
	}
	s.log.Debugf("generating Tag for Metric: %s : tagname: %s", s.cfg.FieldName, tag)
	tags[s.cfg.FieldName] = tag
	return 0
}

func (s *SnmpMetric) computeMultiStringParserValues() {
	ni := len(s.mm)
	var str string
	switch v := s.CookedValue.(type) {
	case string:
		str = s.CookedValue.(string)
	default:
		s.log.Warnf("Error for metric [%s] Type value is not string is %T", s.cfg.ID, v)
		return
	}

	retarray := s.re.FindStringSubmatch(str)
	if len(retarray) < ni {
		s.log.Warnf("Error for metric [%s] parsing REGEXG [%s] on string [%s] without capturing group", s.cfg.ID, s.cfg.ExtraData, str)
		return
	}
	// retarray[0] contains full string
	if len(retarray[1]) == 0 {
		s.log.Warnf("Error for metric [%s] parsing REGEXG [%s] on string [%s] cause  void capturing group", s.cfg.ID, s.cfg.ExtraData, str)
		return
	}
	for k, i := range s.mm {
		s.log.Debugf("Parsing Metric %s MULTISTRING %d value %s (part %s)", s.cfg.ID, k, retarray[0], retarray[k+1])

		var err error
		bitstr := retarray[k+1]
		switch i.IConv {
		case "STR":
			i.Value = bitstr
		case "INT":
			i.Value, err = strconv.ParseInt(bitstr, 10, 64)
		case "BL":
			i.Value, err = strconv.ParseBool(bitstr)
		case "FP":
			i.Value, err = strconv.ParseFloat(bitstr, 64)
		}
		if err != nil {
			s.log.Warnf("Error for Metric %s MULTISTRINGPARSER  Field [%s|%s|%s] Coversion  from  [%s] error: %s", s.cfg.ID, i.IType, i.IName, i.IConv, bitstr, err)
			i.Value = nil
		}
	}
}

func (s *SnmpMetric) addMultiStringParserValues(tags map[string]string, fields map[string]interface{}) int64 {
	var fErrors int64
	s.computeMultiStringParserValues()
	for _, i := range s.mm {
		switch i.IType {
		case "T":
			if i.Value == nil {
				continue
			}
			tags[i.IName] = i.Value.(string)
		case "F":
			if i.Value == nil {
				continue
			}
			fields[i.IName] = i.Value
		}
	}
	return fErrors
}

// ImportFieldsAndTags Add Fields and tags from the metric and returns number of metric sent and metric errors found
func (s *SnmpMetric) ImportFieldsAndTags(mid string, fields map[string]interface{}, tags map[string]string) int64 {
	s.log.Debugf("DEBUG METRIC  CONFIG %+v", s.cfg)
	if s.CookedValue == nil {
		s.log.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", s.cfg.ID, mid, tags, s)
		return 1
	}
	if s.Valid == false {
		s.log.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", s.cfg.ID, mid, tags, s)
		return 1
	}
	if s.Report == NeverReport {
		s.log.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", s.cfg.ID, mid)
		return 0
	}

	switch s.cfg.DataSrcType {
	case "MULTISTRINGPARSER":
		er := s.addMultiStringParserValues(tags, fields)
		return er
	default:
		if s.cfg.IsTag == true {
			er := s.addSingleTag(mid, tags)
			return er
		} else {
			er := s.addSingleField(mid, fields)
			return er
		}
	}
	return 0
}

// MarshalJSON return JSON formatted data
func (s *SnmpMetric) MarshalJSON() ([]byte, error) {
	// type Alias SnmpMetric
	switch s.cfg.DataSrcType {
	case "COUNTER32", "COUNTER64", "COUNTERXX":
		return json.Marshal(&struct {
			FieldName   string
			CookedValue interface{}
			CurValue    interface{}
			LastValue   interface{}
			CurTime     time.Time
			LastTime    time.Time
			Type        string
			Valid       bool
		}{
			FieldName:   s.cfg.FieldName,
			CookedValue: s.CookedValue,
			CurValue:    s.CurValue,
			LastValue:   s.LastValue,
			CurTime:     s.CurTime,
			LastTime:    s.LastTime,
			Type:        s.cfg.DataSrcType,
			Valid:       s.Valid,
		})
	case "MULTISTRINGPARSER":
		return json.Marshal(&struct {
			FieldName   string
			CookedValue interface{}
			ValueMap    []*config.MetricMultiMap
			CurTime     time.Time
			Type        string
			Valid       bool
		}{
			FieldName:   s.cfg.FieldName,
			CookedValue: s.CookedValue,
			ValueMap:    s.mm,
			CurTime:     s.CurTime,
			Type:        s.cfg.DataSrcType,
			Valid:       s.Valid,
		})
	default:
		return json.Marshal(&struct {
			FieldName   string
			CookedValue interface{}
			CurTime     time.Time
			Type        string
			Valid       bool
		}{
			FieldName:   s.cfg.FieldName,
			CookedValue: s.CookedValue,
			CurTime:     s.CurTime,
			Type:        s.cfg.DataSrcType,
			Valid:       s.Valid,
		})
	}
}
