package main

import (
	"errors"
	"github.com/soniah/gosnmp"
	"math"
	"strings"
	"time"
)

//https://collectd.org/wiki/index.php/Data_source

const (
	GAUGE = 0 << iota //value is simply stored as-is
	INTEGER
	COUNTER32
	COUNTER64
	STRING
	HWADDR
	IPADDR
)

/*
3.- Check minimal data is set  (pending)
name, BaseOID BaseOID begining with "."
fieldname != null
*/
// Init initialize metrics
func (m *SnmpMetricCfg) Init(name string) error {
	m.ID = name
	//valIDate config values
	if len(m.FieldName) == 0 {
		return errors.New("FieldName not set in metric Config " + m.ID)
	}
	if len(m.BaseOID) == 0 {
		return errors.New("BaseOid not set in metric Config " + m.ID)
	}
	switch m.DataSrcType {
	case "GAUGE":
	case "INTEGER":
	case "COUNTER32":
	case "COUNTER64":
	case "STRING":
	case "HWADDR":
	case "IPADDR":
	default:
		return errors.New("UnkNown DataSourceType:" + m.DataSrcType + " in metric Config " + m.ID)
	}
	if !strings.HasPrefix(m.BaseOID, ".") {
		return errors.New("Bad BaseOid format:" + m.BaseOID + " in metric Config " + m.ID)
	}

	return nil
}

//SnmpMetric type to metric runtime
type SnmpMetric struct {
	cfg         *SnmpMetricCfg
	cookedValue interface{}
	//cookedValue float64
	curValue   int64
	lastValue  int64
	curTime    time.Time
	lastTime   time.Time
	Compute    func()
	setRawData func(pdu gosnmp.SnmpPDU, now time.Time)
	realOID    string
}

func (s *SnmpMetric) Init() error {
	switch s.cfg.DataSrcType {
	case "GAUGE", "INTEGER":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.cookedValue = float64(val)
			s.curTime = now
			s.Compute()
		}
		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Compute = func() {
				s.cookedValue = (s.cfg.Scale * float64(s.cookedValue.(float64))) + s.cfg.Shift
			}
		} else {
			s.Compute = func() {
			}
		}
	case "COUNTER32":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.lastTime = s.curTime
			s.lastValue = s.curValue
			s.curValue = val
			s.curTime = now
			s.Compute()
		}
		if s.cfg.GetRate == true {
			s.Compute = func() {
				duration := s.curTime.Sub(s.lastTime)
				if s.curValue < s.lastValue {
					s.cookedValue = float64(math.MaxInt32-s.lastValue+s.curValue) / duration.Seconds()
				} else {
					s.cookedValue = float64(s.curValue-s.lastValue) / duration.Seconds()
				}
			}
		} else {
			s.Compute = func() {
				if s.curValue < s.lastValue {
					s.cookedValue = float64(math.MaxInt32 - s.lastValue + s.curValue)
				} else {
					s.cookedValue = float64(s.curValue - s.lastValue)
				}
			}

		}
	case "COUNTER64":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.lastTime = s.curTime
			s.lastValue = s.curValue
			s.curValue = val
			s.curTime = now
			s.Compute()
		}
		if s.cfg.GetRate == true {
			s.Compute = func() {
				duration := s.curTime.Sub(s.lastTime)
				if s.curValue < s.lastValue {
					s.cookedValue = float64(math.MaxInt64-s.lastValue+s.curValue) / duration.Seconds()
				} else {
					s.cookedValue = float64(s.curValue-s.lastValue) / duration.Seconds()
				}
			}
		} else {
			s.Compute = func() {
				if s.curValue < s.lastValue {
					s.cookedValue = float64(math.MaxInt64 - s.lastValue + s.curValue)
				} else {
					s.cookedValue = float64(s.curValue - s.lastValue)
				}
			}

		}
	case "STRING":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.cookedValue = pduVal2str(pdu)
			s.curTime = now
		}
	case "IPADDR":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.cookedValue, _ = pduVal2IPaddr(pdu)
			s.curTime = now
		}
	case "HWADDR":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.cookedValue, _ = pduVal2IPaddr(pdu)
			s.curTime = now
		}
	}
	return nil
}
