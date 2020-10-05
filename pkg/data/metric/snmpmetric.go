package metric

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/raven-go"
	"reflect"

	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"snmpcollector/pkg/config"
	"snmpcollector/pkg/data/filter"
	"snmpcollector/pkg/data/snmp"
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
	// NeverReport  default value for metric resport flag
	NeverReport = 0
	// AlwaysReport  metric will send data allways
	AlwaysReport = 1
	// OnNonZeroReport metric will send data only if computed value different than 0
	OnNonZeroReport = 2
	// OnChangedReport metric will send data only if data has a change from last changed value ( not implemente yet)
	OnChangedReport = 3
)

//SnmpMetric type to metric runtime
type SnmpMetric struct {
	cfg         *config.SnmpMetricCfg
	Valid       bool //indicate if has been updated in the last gathered process
	CookedValue map[string]interface{}
	CurValue    map[string]interface{}
	LastValue   map[string]interface{}
	CurTime     time.Time
	LastTime    time.Time
	ElapsedTime float64
	Compute     func(arg ...interface{})                `json:"-"`
	Scale       func()                                  `json:"-"`
	Convert     func()                                  `json:"-"`
	SetRawData  func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) `json:"-"`
	RealOID     string
	Report      int //if false this metric won't be sent to the output buffer (is just taken as a coomputed input for other metrics)
	//for STRINGPARSER/MULTISTRINGPARSER
	re   *regexp.Regexp
	mm   []*config.MetricMultiMap
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
func New(c *config.SnmpMetricCfg, l *logrus.Logger) (*SnmpMetric, error) {
	metric := &SnmpMetric{log: l}
	metric.CookedValue = make(map[string]interface{})
	metric.CurValue = make(map[string]interface{})
	metric.LastValue = make(map[string]interface{})
	err := metric.Init(c)
	return metric, err
}

// SetLogger attach logger to the current snmpmetric object
func (s *SnmpMetric) SetLogger(l *logrus.Logger) {
	s.log = l
}

// Conversion functions

func (s *SnmpMetric) convertFromUInteger() {
	//check first the rigth
	for cookedValueKey, cookedValue := range s.CookedValue {

		if cookedValue == nil {
			delete(s.CookedValue, cookedValueKey)
			continue
		}

		var reflectType = reflect.TypeOf(cookedValue)
		s.log.Info("cookedValue type: ", reflectType)

		switch vt := cookedValue.(type) {
		case int64:
			//everything ok
			break
		case uint64:
			s.CookedValue[cookedValueKey] = int64(cookedValue.(uint64))
			cookedValue = s.CookedValue[cookedValueKey]
			break
		case float64:
			s.CookedValue[cookedValueKey] = int64(cookedValue.(float64))
			cookedValue = s.CookedValue[cookedValueKey]
			break
		default:
			s.log.Errorf("ERROR: expected value on metric %s type UINT64 and got %T  type ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
			// var msg = fmt.Sprintf("ERROR: expected value on metric %s type UINT64 and got %T  type ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
			// raven.CaptureMessage(msg, map[string]string{"category": "typecast"})
			continue
		}
		//the only acceptable conversions
		// signed integer 64 -> float64
		// signet integer 64 -> boolean ( true if value != 0 )
		switch s.cfg.Conversion {
		case config.INTEGER:
			continue
		case config.FLOAT:
			s.CookedValue[cookedValueKey] = float64(cookedValue.(int64))
			continue
		case config.BOOLEAN:
			if cookedValue.(int64) != 0 {
				s.CookedValue[cookedValueKey] = true
			} else {
				s.CookedValue[cookedValueKey]= false
			}
			continue
		case config.STRING:
			s.CookedValue[cookedValueKey] = strconv.FormatInt(cookedValue.(int64), 10)
			continue
		default:
			s.log.Errorf("Bad conversion: requested %s from %T type", s.cfg.Conversion.GetString(), s.CookedValue)
		}

	}
}

func (s *SnmpMetric) convertFromInteger() {
	for cookedValueKey, cookedValue := range s.CookedValue {


		if cookedValue == nil {
			delete(s.CookedValue, cookedValueKey)
			continue
		}

		var reflectType = reflect.TypeOf(cookedValue)
		s.log.Info("cookedValue type: ", reflectType)

		switch vt := cookedValue.(type) {
		case int64:
			//everything ok
			break
		case uint64:
			s.CookedValue[cookedValueKey] = int64(cookedValue.(uint64))
			cookedValue = s.CookedValue[cookedValueKey]
			break
		case float64:
			s.CookedValue[cookedValueKey] = int64(cookedValue.(float64))
			cookedValue = s.CookedValue[cookedValueKey]
			break
		default:
			s.log.Errorf("ERROR: expected value on metric %s type INT64 and got %T ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
			// var msg = fmt.Sprintf("ERROR: expected value on metric %s type INT64 and got %T ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
			// raven.CaptureMessage(msg, map[string]string{"category": "typecast"})
			continue
		}
		//the only acceptable conversions
		// signed integer 64 -> float64
		// signet integer 64 -> boolean ( true if value != 0 )
		switch s.cfg.Conversion {
		case config.INTEGER:
			continue
		case config.FLOAT:
			s.CookedValue[cookedValueKey] = float64(cookedValue.(int64))
			continue
		case config.BOOLEAN:
			if cookedValue.(int64) != 0 {
				s.CookedValue[cookedValueKey] = true
			} else {
				s.CookedValue[cookedValueKey] = false
			}
			continue
		case config.STRING:
			s.CookedValue[cookedValueKey]= strconv.FormatInt(cookedValue.(int64), 10)
			continue
		default:
			s.log.Errorf("Bad conversion: requested %s from %T type", s.cfg.Conversion.GetString(), s.CookedValue)
		}

	}
}

func (s *SnmpMetric) convertFromFloat() {
	for cookedValueKey, cookedValue := range s.CookedValue {

		if cookedValue == nil {
			delete(s.CookedValue, cookedValueKey)
			continue
		}

		switch vt := cookedValue.(type) {
		case float64:
			//everything ok
			break
		default:
			s.log.Errorf("ERROR: expected value on metric %s type Float64 and got %T type ( %+v) \n", s.cfg.ID, vt, s.CookedValue)
			var msg = fmt.Sprintf("ERROR: expected value on metric %s type Float64 and got %T type ( %+v) \n", s.cfg.ID, vt, s.CookedValue)
			raven.CaptureMessage(msg, map[string]string{"category": "typecast"})
			continue
		}
		//the only acceptable conversions
		// signed float -> int64 (will do rounded value)
		switch s.cfg.Conversion {
		case config.INTEGER:
			s.CookedValue[cookedValueKey] = int64(math.Round(cookedValue.(float64)))
			continue
		case config.FLOAT:
			continue
		case config.BOOLEAN:
			if cookedValue.(float64) != 0.0 {
				s.CookedValue[cookedValueKey] = true
			} else {
				s.CookedValue[cookedValueKey] = false
			}
			continue
		case config.STRING:
			s.CookedValue[cookedValueKey] = strconv.FormatFloat(cookedValue.(float64), 'f', -1, 64)
			continue
		default:
			s.log.Errorf("Bad conversion: requested on metric %s: to type  %s from %T type", s.cfg.ID, s.cfg.Conversion.GetString(), s.CookedValue)
		}
	}

}

func (s *SnmpMetric) convertFromString() {
	for cookedValueKey, cookedValue := range s.CookedValue {

		switch vt := cookedValue.(type) {
		case string:
			//everything ok
			break
		default:
			s.log.Errorf("ERROR: expected value on metric %s type STRING and got %T type ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
			// var msg = fmt.Sprintf("ERROR: expected value on metric %s type STRING and got %T type ( %+v) type \n", s.cfg.ID, vt, s.CookedValue)
			// raven.CaptureMessage(msg, map[string]string{"category": "typecast"})
			continue
		}
		//the only acceptable conversions
		// string -> int64
		// string -> float (the default)
		// string -> boolean
		// string -> string
		switch s.cfg.Conversion {
		case config.STRING:
			continue
		case config.INTEGER:
			value, err := strconv.ParseInt(cookedValue.(string), 10, 64)
			if err != nil {
				s.log.Warnf("Error parsing Integer from String  %s metric %s : error: %s", cookedValue.(string), s.cfg.ID, err)
				continue
			}
			s.CookedValue[cookedValueKey] = value
			continue
		case config.FLOAT:
			value, err := strconv.ParseFloat(cookedValue.(string), 64)
			if err != nil {
				s.log.Warnf("Error parsing float from String  %s metric %s : error: %s", cookedValue.(string), s.cfg.ID, err)
				continue
			}
			s.CookedValue[cookedValueKey] = value
			continue
		case config.BOOLEAN:
			value, err := strconv.ParseBool(cookedValue.(string))
			if err != nil {
				s.log.Warnf("Error parsing Boolean from String  %s metric %s : error: %s", cookedValue.(string), s.cfg.ID, err)
				continue
			}
			s.CookedValue[cookedValueKey] = value
			continue
		default:
			s.log.Errorf("Bad conversion: requested on metric %s: to type  %s from %T type", s.cfg.ID, s.cfg.Conversion.GetString(), s.CookedValue)
		}
	}

}

func (s *SnmpMetric) convertFromAny() {

	for _, cookedValue := range s.CookedValue {

		switch cookedValue.(type) {
		case float32, float64:
			s.convertFromFloat()
			continue
		case uint64, uint32:
			s.convertFromUInteger()
			continue
		case int64, int32:
			s.convertFromInteger()
			continue
		case string:
			s.convertFromString()
			continue
		case bool:
			s.log.Errorf("Bad conversion: requested on metric %s: to type  %s from %T type", s.cfg.ID, s.cfg.Conversion.GetString(), s.CookedValue)
		default:
		}
	}


}

// Init Initialice a new snmpmetric object with the specific configuration
func (s *SnmpMetric) Init(c *config.SnmpMetricCfg) error {
	if c == nil {
		return fmt.Errorf("Error on initialice device, configuration struct is nil")
	}
	s.cfg = c
	s.RealOID = c.BaseOID
	//set default conversion funcion
	s.Convert = s.convertFromAny
	//Force conversion to STRING if metric is tag.
	if s.cfg.IsTag == true {
		s.cfg.Conversion = config.STRING
	}
	if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
		s.Scale = func() {

			for cookedValueKey, cookedValue := range s.CookedValue {

				switch v := cookedValue.(type) {
				case uint64:
					//should return Integer
					val := (s.cfg.Scale * float64(cookedValue.(uint64))) + s.cfg.Shift
					s.CookedValue[cookedValueKey] = uint64(math.Round(val))
				case float64:
					//should return float
					s.CookedValue[cookedValueKey] = float64((s.cfg.Scale * float64(cookedValue.(float64))) + s.cfg.Shift)
				case string:
					s.log.Errorf("Error Trying to  Scale Function from non numbered STRING type value : %s ", s.CookedValue)
				default:
					s.log.Errorf("Error Trying to  Scale Function from unknown type %T value: %#+v", v, s.CookedValue)
				}

			}

		}
	} else {
		s.Scale = func() {
		}
	}

	if len(s.CookedValue) == 0 {
		s.CookedValue[""] = nil // init values for setting raw data methods
	}

	for cookedValueKey := range s.CookedValue {

		switch s.cfg.DataSrcType {
		case "CONDITIONEVAL":
			//select
			cond, err := dbc.GetOidConditionCfgByID(s.cfg.ExtraData)
			if err != nil {
				s.log.Errorf("Error getting CONDITIONEVAL [id: %s ] data : %s", s.cfg.ExtraData, err)
			}
			//get Regexp
			if cond.IsMultiple == true {
				s.condflt = filter.NewOidMultipleFilter(cond.OIDCond, s.log)
			} else {
				s.condflt = filter.NewOidFilter(cond.OIDCond, cond.CondType, cond.CondValue, s.log, cond.Encoding)
			}
			s.Compute = func(arg ...interface{}) {
				s.condflt.Init(arg...)
				s.condflt.Update()
				s.CookedValue[cookedValueKey] = s.condflt.Count()
				s.CurTime = time.Now()
				s.Valid = true
			}
			//Sign
			//set Process Data
		case "TIMETICKS": //Cooked TimeTicks
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}
				val := snmp.PduVal2Int64(pdu)
				s.CookedValue[valKey] = val / 100 //now data in secoonds
				s.CurTime = now
				s.Scale()
				s.convertFromInteger()
				s.Valid = true
			}

			//Signed Integers
		case "INTEGER", "Integer32":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}
				s.CookedValue[valKey] = snmp.PduVal2Int64(pdu)
				s.CurTime = now
				s.Scale()
				s.convertFromInteger()
				s.Valid = true
			}
			//Unsigned Integers
		case "Counter32", "Gauge32", "Counter64", "TimeTicks", "UInteger32", "Unsigned32":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}
				s.CookedValue[valKey] = uint64(snmp.PduVal2UInt64(pdu))
				s.CurTime = now
				s.Scale()
				s.convertFromUInteger()
				s.Valid = true
			}
		case "COUNTER32": //Increment computed
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
				//first time only set values and reassign itself to the complete method this will avoi to send invalid data


				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				val := snmp.PduVal2UInt64(pdu)
				s.CurValue[cookedValueKey] = val
				s.CurTime = now
				s.Valid = true

				s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

					valKey = cookedValueKey
					if subIndexes != nil {
						valKey = subIndex
					}

					val := snmp.PduVal2UInt64(pdu)
					s.LastTime = s.CurTime
					s.LastValue[valKey] = s.CurValue[valKey]
					s.CurValue[valKey]= val
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
					if s.CurValue[cookedValueKey].(uint64) < s.LastValue[cookedValueKey].(uint64) {
						s.CookedValue[cookedValueKey] = float64(math.MaxUint32-s.LastValue[cookedValueKey].(uint64)+s.CurValue[cookedValueKey].(uint64)) / s.ElapsedTime
					} else {
						s.CookedValue[cookedValueKey] = float64(s.CurValue[cookedValueKey].(uint64)-s.LastValue[cookedValueKey].(uint64)) / s.ElapsedTime
					}
				}
				s.Convert = s.convertFromFloat
			} else {
				s.Compute = func(arg ...interface{}) {
					s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
					if s.CurValue[cookedValueKey].(uint64) < s.LastValue[cookedValueKey].(uint64) {
						s.CookedValue[cookedValueKey] = math.MaxUint32 - s.LastValue[cookedValueKey].(uint64) + s.CurValue[cookedValueKey].(uint64)
					} else {
						s.CookedValue[cookedValueKey] = s.CurValue[cookedValueKey].(uint64) - s.LastValue[cookedValueKey].(uint64)
					}
				}
				s.Convert = s.convertFromUInteger
			}
		case "COUNTER64": //Increment computed
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
				//log.Debugf("========================================>COUNTER64: first time :%s ", s.RealOID)
				//first time only set values and reassign itself to the complete method

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}


				val := snmp.PduVal2UInt64(pdu)
				s.CurValue[valKey] = val
				s.CurTime = now
				s.Valid = true
				s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
					//log.Debugf("========================================>COUNTER64: the other time:%s", s.RealOID)

					valKey = cookedValueKey
					if subIndexes != nil {
						valKey = subIndex
					}

					val := snmp.PduVal2UInt64(pdu)
					s.LastTime = s.CurTime
					s.LastValue[valKey] = s.CurValue[valKey]
					s.CurValue[valKey] = val
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
					if s.CurValue[cookedValueKey].(uint64) < s.LastValue[cookedValueKey].(uint64) {
						s.CookedValue[cookedValueKey] = float64(math.MaxUint64-s.LastValue[cookedValueKey].(uint64)+s.CurValue[cookedValueKey].(uint64)) / s.ElapsedTime
					} else {
						s.CookedValue[cookedValueKey] = float64(s.CurValue[cookedValueKey].(uint64)-s.LastValue[cookedValueKey].(uint64)) / s.ElapsedTime
					}
				}
				s.Convert = s.convertFromFloat
			} else {
				s.Compute = func(arg ...interface{}) {
					s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
					if s.CurValue[cookedValueKey].(uint64) < s.LastValue[cookedValueKey].(uint64) {
						s.CookedValue[cookedValueKey] = math.MaxUint64 - s.LastValue[cookedValueKey].(uint64) + s.CurValue[cookedValueKey].(uint64)
					} else {
						s.CookedValue[cookedValueKey] = s.CurValue[cookedValueKey].(uint64) - s.LastValue[cookedValueKey].(uint64)
					}
				}
				s.Convert = s.convertFromUInteger
			}
		case "COUNTERXX": //Generic Counter With Unknown range or buggy counters that  Like Non negative derivative
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
				//first time only set values and reassign itself to the complete method this will avoi to send invalid data

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				val := snmp.PduVal2UInt64(pdu)
				s.CurValue[cookedValueKey] = val
				s.CurTime = now
				s.Valid = true
				s.log.Debugf("FIRST RAW(post-compute): %T - %#+v", s.CookedValue, s.CookedValue)
				s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

					valKey = cookedValueKey
					if subIndexes != nil {
						valKey = subIndex
					}

					val := snmp.PduVal2UInt64(pdu)
					s.LastTime = s.CurTime
					s.LastValue[valKey] = s.CurValue[valKey]
					s.CurValue[valKey] = val
					s.CurTime = now
					s.Compute()
					s.Valid = true
				}
			}
			if s.cfg.GetRate == true {
				s.Compute = func(arg ...interface{}) {
					s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
					if s.CurValue[cookedValueKey].(uint64) >= s.LastValue[cookedValueKey].(uint64) {
						s.CookedValue[cookedValueKey] = float64(s.CurValue[cookedValueKey].(uint64)-s.LastValue[cookedValueKey].(uint64)) / s.ElapsedTime
						s.Scale()
						s.Convert()
					} else {
						// Else => nothing to do last value will be sent
						s.log.Warnf("Warning Negative COUNTER increment [current: %d | last: %d ] last value will be sent %f", s.CurValue[cookedValueKey], s.LastValue[cookedValueKey], s.CookedValue[cookedValueKey])
					}
				}
				s.Convert = s.convertFromFloat
			} else {
				s.Compute = func(arg ...interface{}) {
					s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
					if s.CurValue[cookedValueKey].(uint64) >= s.LastValue[cookedValueKey].(uint64) {
						s.CookedValue[cookedValueKey] = s.CurValue[cookedValueKey].(uint64) - s.LastValue[cookedValueKey].(uint64)
						s.Scale()
						s.Convert()
					} else {
						// Else => nothing to do last value will be sent
						s.log.Warnf("Warning Negative COUNTER increment [current: %d | last: %d ] last value will be sent %f", s.CurValue[cookedValueKey], s.LastValue[cookedValueKey], s.CookedValue[cookedValueKey])
					}
				}
				s.Convert = s.convertFromUInteger
			}
		case "BITS":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				barray := snmp.PduVal2BoolArray(pdu)
				names := []string{}
				for i, b := range barray {
					if b {
						names = append(names, s.cfg.Names[i])
					}
				}
				s.CookedValue[valKey] = strings.Join(names, ",")
				s.CurTime = now
				s.Valid = true
				s.log.Debugf("SETRAW BITS %+v, RESULT %s", s.cfg.Names, s.CookedValue)
			}
		case "BITSCHK":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {
				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				barray := snmp.PduVal2BoolArray(pdu)
				index, _ := strconv.Atoi(s.cfg.ExtraData)
				b := barray[index]
				if b {
					s.CookedValue[valKey] = 1.0
				} else {
					s.CookedValue[valKey] = 0.0
				}
				s.Convert()
				s.CurTime = now
				s.Valid = true
				s.log.Debugf("BITS CHECK bit %+v, Position %d , RESULT %t", barray, index, s.CookedValue)
			}
		case "ENUM":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				idx := snmp.PduVal2Int64(pdu)
				if val, ok := s.cfg.Names[int(idx)]; ok {
					s.CookedValue[valKey] = val
				} else {
					s.CookedValue[valKey] = strconv.Itoa(int(idx))
				}
				s.Valid = true
				s.log.Debugf("SETRAW ENUM %+v, RESULT %s", s.cfg.Names, s.CookedValue)
			}
		case "OCTETSTRING":
			switch s.cfg.Conversion {

			case config.INTEGER:
				s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

					var valKey = cookedValueKey
					if subIndexes != nil {
						valKey = subIndex
					}

					val, err := snmp.PduValHexString2Uint(pdu)
					s.CookedValue[valKey] = val
					s.CurTime = now
					if err != nil {
						s.log.Warnf("Error on HexString to UINT conversion: %s", err)
						return
					}
					s.Valid = true
				}
				//For compatibility purposes with previous versions
			case config.FLOAT:
				s.log.Errorf("WARNING ON SNMPMETRIC ( %s ): You are using version >=0.8 version without database upgrade: you should upgrade the DB by executing this SQL on your database \"update snmp_metric_cfg set Conversion=3 where datasrctype='OCTETSTRING';\", to avoid this message ", s.cfg.ID)
				fallthrough
			case config.STRING:
				s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

					var valKey = cookedValueKey
					if subIndexes != nil {
						valKey = subIndex
					}

					s.CookedValue[valKey] = snmp.PduVal2str(pdu)
					s.CurTime = now
					s.Valid = true
				}
			default:
				s.log.Errorf("WARNING ON SNMPMETRIC ( %s ): Invalid conversion mode from OCTETSTRING to %s", s.cfg.ID, s.cfg.Conversion.GetString())
				s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

					var valKey = cookedValueKey
					if subIndexes != nil {
						valKey = subIndex
					}

					s.CookedValue[valKey] = snmp.PduVal2str(pdu)
					s.CurTime = now
					s.Valid = true
				}
			}

		case "OID":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				s.CookedValue[valKey] = snmp.PduVal2OID(pdu)
				s.CurTime = now
				s.Valid = true
			}
		case "IpAddress":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				s.CookedValue[valKey], _ = snmp.PduVal2IPaddr(pdu)
				s.CurTime = now
				s.Valid = true
			}
		case "HWADDR":
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				s.CookedValue[valKey], _ = snmp.PduVal2Hwaddr(pdu)
				s.CurTime = now
				s.Valid = true
			}
		case "STRINGPARSER":
			//get Regexp
			re, err := regexp.Compile(s.cfg.ExtraData)
			if err != nil {
				return fmt.Errorf("Error on initialice STRINGPARSER, invalind Regular Expression : %s", s.cfg.ExtraData)
			}
			s.re = re
			//set Process Data
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

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
				s.CookedValue[valKey] = retarray[1]
				s.CurTime = now
				s.convertFromString()
				//s.Scale() <-only valid if Integer or Float
				s.Valid = true
			}
		case "MULTISTRINGPARSER":
			//get Regexp
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
			//set Process Data
			s.SetRawData = func(pdu gosnmp.SnmpPDU, subIndexes []string, subIndex string, now time.Time) {

				var valKey = cookedValueKey
				if subIndexes != nil {
					valKey = subIndex
				}

				str := snmp.PduVal2str(pdu)
				s.CookedValue[valKey] = str
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
			//set Process Data
			s.Compute = func(arg ...interface{}) {
				parameters := arg[0].(map[string]interface{})
				s.log.Debugf("Evaluating Metric %s with eval expresion [%s] with parameters %+v", s.cfg.ID, s.cfg.ExtraData, parameters)
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
				s.CookedValue[cookedValueKey] = result
				//conversion depends onthe type of the evaluted data.
				s.CurTime = time.Now()
				s.Scale()
				s.Convert() //default
				s.Valid = true
			}
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
		if s.Valid == true { //only valid for compute if it has been updated last
			params[s.cfg.FieldName] = s.CookedValue
		}
	}
}

func (s *SnmpMetric) addSingleField(mid string, fields map[string]interface{}) int64 {

		if s.Report == OnNonZeroReport {
			if s.CookedValue[""] == 0.0 {
				s.log.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", s.cfg.ID, mid)
				return 0
			}
		}
		//assuming float Cooked Values
		s.log.Debugf("generating field for %s value %#v ", s.cfg.FieldName, s.CookedValue)
		s.log.Debugf("DEBUG METRIC %+v", s)
		fields[s.cfg.FieldName] = s.CookedValue


	return 0
}

func (s *SnmpMetric) addSingleTag(mid string, tags map[string]string) int64 {

	for _, cookedValue := range s.CookedValue {
		var tag string
		switch v := cookedValue.(type) {
		case string:
			tag = v
		default:
			s.log.Debugf("ERROR wrong type %T for ID [%s] from MEASUREMENT[ %s ] when converting to TAG(STRING) won't be reported to the output backend", v, s.cfg.ID, mid)
			return 1
		}
		//I don't know if a OnNonZeroReport could have sense in any configuration.
		if s.Report == OnNonZeroReport {
			if tag == "0" {
				s.log.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", s.cfg.ID, mid)
				return 0
			}
		}
		s.log.Debugf("generating Tag for Metric: %s : tagname: %s", s.cfg.FieldName, tag)
		tags[s.cfg.FieldName] = tag
	}


	return 0
}

func (s *SnmpMetric) computeMultiStringParserValues() {

	for _, cookedValue := range s.CookedValue {

		ni := len(s.mm)
		var str string
		switch v := cookedValue.(type) {
		case string:
			str = cookedValue.(string)
		default:
			s.log.Warnf("Error for metric [%s] Type value is not string is %T", s.cfg.ID, v)
			return
		}

		retarray := s.re.FindStringSubmatch(str)
		if len(retarray) < ni {
			s.log.Warnf("Error for metric [%s] parsing REGEXG [%s] on string [%s] without capturing group", s.cfg.ID, s.cfg.ExtraData, str)
			return
		}
		//retarray[0] contains full string
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
			fields[i.IName] = i.Value
		}
	}
	return fErrors
}

// ImportFieldsAndTags Add Fields and tags from the metric and returns number of metric sent and metric errors found
func (s *SnmpMetric) ImportFieldsAndTags(mid string, fields map[string]interface{}, tags map[string]string) (int64, int64) {
	var metError int64
	var metSent int64
	s.log.Debugf("DEBUG METRIC  CONFIG %+v", s.cfg)
	if s.CookedValue == nil {
		s.log.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", s.cfg.ID, mid, tags, s)
		metError++ //not sure if an tag error should be count as metric
		return metError, metSent
	}
	if s.Valid == false {
		s.log.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", s.cfg.ID, mid, tags, s)
		return 0, 0
	}
	if s.Report == NeverReport {
		s.log.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", s.cfg.ID, mid)
		return 0, 0
	}

	switch s.cfg.DataSrcType {
	case "MULTISTRINGPARSER":
		er := s.addMultiStringParserValues(tags, fields)
		metError += er
	default:
		if s.cfg.IsTag == true {
			er := s.addSingleTag(mid, tags)
			metError += er
		} else {
			er := s.addSingleField(mid, fields)
			metError += er
		}
	}

	return metSent, metError
}

// MarshalJSON return JSON formatted data
func (s *SnmpMetric) MarshalJSON() ([]byte, error) {
	//type Alias SnmpMetric
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
