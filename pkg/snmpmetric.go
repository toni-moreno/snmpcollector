package main

import (
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
3.- Check minimal data is set  (pending)
name, BaseOID BaseOID begining with "."
fieldname != null
*/
// Init initialize metrics
func (m *SnmpMetricCfg) Init() error {
	//valIDate config values
	if len(m.FieldName) == 0 {
		return errors.New("FieldName not set in metric Config " + m.ID)
	}
	if len(m.BaseOID) == 0 && m.DataSrcType != "STRINGEVAL" {
		return fmt.Errorf("BaseOid not set in metric Config %s type  %s"+m.ID, m.DataSrcType)
	}
	//https://tools.ietf.org/html/rfc2578 (SMIv2)
	//https://tools.ietf.org/html/rfc2579 (Textual Conventions for SMIv2)
	//https://tools.ietf.org/html/rfc2851 (Textual Conventions for Internet Network Address)
	switch m.DataSrcType {
	case "INTEGER", "Integer32":
	case "Gauge32":
	case "UInteger32", "Unsigned32":
	case "Counter32", "COUNTER32": //raw and cooked increment of Counter32
	case "Counter64", "COUNTER64": //raw and Cooked increment of Counter34
	case "TimeTicks", "TIMETICKS": //raw and cooked to second of timeticks
	case "OCTETSTRING":
	case "HWADDR":
	case "IpAddress":
	case "STRINGPARSER":
	case "STRINGEVAL":
	default:
		return errors.New("UnkNown DataSourceType:" + m.DataSrcType + " in metric Config " + m.ID)
	}
	if m.DataSrcType != "STRINGEVAL" && !strings.HasPrefix(m.BaseOID, ".") {
		return errors.New("Bad BaseOid format:" + m.BaseOID + " in metric Config " + m.ID)
	}
	if m.DataSrcType == "STRINGPARSER" && len(m.ExtraData) == 0 {
		return errors.New("STRINGPARSER type requires extradata to work " + m.ID)
	}
	if m.DataSrcType == "STRINGEVAL" && len(m.ExtraData) == 0 {
		return fmt.Errorf("ExtraData not set in metric Config %s type  %s"+m.ID, m.DataSrcType)
	}
	return nil
}

func (m SnmpMetricCfg) CheckEvalCfg(parameters map[string]interface{}) error {
	if m.DataSrcType != "STRINGEVAL" {
		return nil
	}
	expression, err := govaluate.NewEvaluableExpression(m.ExtraData)
	if err != nil {
		//log.Errorf("Error on initialice STRINGEVAL on metric %s evaluation : %s : ERROR : %s", m.ID, m.ExtraData, err)
		return err
	}
	_, err = expression.Evaluate(parameters)
	if err != nil {
		//log.Errorf("Error in metric %s On EVAL string: %s : ERROR : %s", m.ID, m.ExtraData, err)
		return err
	}
	return nil
}

//SnmpMetric type to metric runtime
type SnmpMetric struct {
	cfg         *SnmpMetricCfg
	ID          string
	CookedValue interface{}
	CurValue    interface{}
	LastValue   interface{}
	CurTime     time.Time
	LastTime    time.Time
	ElapsedTime float64
	Compute     func(arg ...interface{}) `json:"-"`
	Scale       func()                   `json:"-"`
	setRawData  func(pdu gosnmp.SnmpPDU, now time.Time)
	RealOID     string
	Report      bool //if false this metric won't be sent to the ouput buffer (is just taken as a coomputed input for other metrics)
	//for STRINGPARSER
	re   *regexp.Regexp
	expr *govaluate.EvaluableExpression
	log  *logrus.Logger
}

// NewSnmpMetric constructor
func NewSnmpMetric(c *SnmpMetricCfg) (*SnmpMetric, error) {
	metric := &SnmpMetric{}
	err := metric.Init(c)
	return metric, err
}

func (s *SnmpMetric) SetLogger(l *logrus.Logger) {
	s.log = l
}

func (s *SnmpMetric) Init(c *SnmpMetricCfg) error {
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
	case "TIMETICKS": //Cooked TimeTicks
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.CookedValue = float64(val / 100) //now data in secoonds
			s.CurTime = now
			s.Scale()
		}
		//Signed Integers
	case "INTEGER", "Integer32":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.CookedValue = float64(val)
			s.CurTime = now
			s.Scale()
		}
		//Unsigned Integers
	case "Counter32", "Gauge32", "Counter64", "TimeTicks", "UInteger32", "Unsigned32":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2UInt64(pdu)
			s.CookedValue = float64(val)
			s.CurTime = now
			s.Scale()
		}
	case "COUNTER32": //Increment computed
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			//first time only set values and reassign itself to the complete method this will avoi to send invalid data
			val := pduVal2UInt64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				val := pduVal2UInt64(pdu)
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
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			//log.Debugf("========================================>COUNTER64: first time :%s ", s.RealOID)
			//first time only set values and reassign itself to the complete method
			val := pduVal2UInt64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				//log.Debugf("========================================>COUNTER64: the other time:%s", s.RealOID)
				val := pduVal2UInt64(pdu)
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
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue = pduVal2str(pdu)
			s.CurTime = now
		}
	case "IpAddress":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = pduVal2IPaddr(pdu)
			s.CurTime = now
		}
	case "HWADDR":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = pduVal2Hwaddr(pdu)
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
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			str := pduVal2str(pdu)
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
