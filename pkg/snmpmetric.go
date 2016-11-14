package main

import (
	"errors"
	"fmt"
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
	ID          string
	CookedValue interface{}
	//CookedValue float64
	curValue   int64
	lastValue  int64
	CurTime    time.Time
	lastTime   time.Time
	Compute    func() `json:"-"`
	setRawData func(pdu gosnmp.SnmpPDU, now time.Time)
	RealOID    string
}

func NewSnmpMetric(c *SnmpMetricCfg) (*SnmpMetric, error) {
	metric := &SnmpMetric{}
	err := metric.Init(c)
	return metric, err
}

func (s *SnmpMetric) Init(c *SnmpMetricCfg) error {
	if c == nil {
		return fmt.Errorf("Error on initialice device, configuration struct is nil")
	}
	s.cfg = c
	s.RealOID = c.BaseOID
	s.ID = s.cfg.ID
	switch s.cfg.DataSrcType {
	case "GAUGE", "INTEGER":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.CookedValue = float64(val)
			s.CurTime = now
			s.Compute()
		}
		if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
			s.Compute = func() {
				s.CookedValue = (s.cfg.Scale * float64(s.CookedValue.(float64))) + s.cfg.Shift
			}
		} else {
			s.Compute = func() {
			}
		}
	case "COUNTER32":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.lastTime = s.CurTime
			s.lastValue = s.curValue
			s.curValue = val
			s.CurTime = now
			s.Compute()
		}
		if s.cfg.GetRate == true {
			s.Compute = func() {
				duration := s.CurTime.Sub(s.lastTime)
				if s.curValue < s.lastValue {
					s.CookedValue = float64(math.MaxInt32-s.lastValue+s.curValue) / duration.Seconds()
				} else {
					s.CookedValue = float64(s.curValue-s.lastValue) / duration.Seconds()
				}
			}
		} else {
			s.Compute = func() {
				if s.curValue < s.lastValue {
					s.CookedValue = float64(math.MaxInt32 - s.lastValue + s.curValue)
				} else {
					s.CookedValue = float64(s.curValue - s.lastValue)
				}
			}

		}
	case "COUNTER64":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.lastTime = s.CurTime
			s.lastValue = s.curValue
			s.curValue = val
			s.CurTime = now
			s.Compute()
		}
		if s.cfg.GetRate == true {
			s.Compute = func() {
				duration := s.CurTime.Sub(s.lastTime)
				if s.curValue < s.lastValue {
					s.CookedValue = float64(math.MaxInt64-s.lastValue+s.curValue) / duration.Seconds()
				} else {
					s.CookedValue = float64(s.curValue-s.lastValue) / duration.Seconds()
				}
			}
		} else {
			s.Compute = func() {
				if s.curValue < s.lastValue {
					s.CookedValue = float64(math.MaxInt64 - s.lastValue + s.curValue)
				} else {
					s.CookedValue = float64(s.curValue - s.lastValue)
				}
			}

		}
	case "STRING":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue = pduVal2str(pdu)
			s.CurTime = now
		}
	case "IPADDR":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = pduVal2IPaddr(pdu)
			s.CurTime = now
		}
	case "HWADDR":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			s.CookedValue, _ = pduVal2IPaddr(pdu)
			s.CurTime = now
		}
	}
	return nil
}
