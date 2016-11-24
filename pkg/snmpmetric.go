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
	CurValue    int64
	LastValue   int64
	CurTime     time.Time
	LastTime    time.Time
	ElapsedTime float64
	Compute     func() `json:"-"`
	Scale       func() `json:"-"`
	setRawData  func(pdu gosnmp.SnmpPDU, now time.Time)
	RealOID     string
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
	if s.cfg.Scale != 0.0 || s.cfg.Shift != 0.0 {
		s.Scale = func() {
			s.CookedValue = (s.cfg.Scale * float64(s.CookedValue.(float64))) + s.cfg.Shift
		}
	} else {
		s.Scale = func() {
		}
	}
	switch s.cfg.DataSrcType {
	case "GAUGE", "INTEGER":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			val := pduVal2Int64(pdu)
			s.CookedValue = float64(val)
			s.CurTime = now
			//s.Compute()
			s.Scale()
		}
	case "COUNTER32":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			//first time only set values and reassign itself to the complete method this will avoi to send invalid data
			val := pduVal2Int64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				val := pduVal2Int64(pdu)
				s.LastTime = s.CurTime
				s.LastValue = s.CurValue
				s.CurValue = val
				s.CurTime = now
				s.Compute()
				s.Scale()
			}
		}
		if s.cfg.GetRate == true {
			s.Compute = func() {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue < s.LastValue {
					s.CookedValue = float64(math.MaxInt32-s.LastValue+s.CurValue) / s.ElapsedTime
				} else {
					s.CookedValue = float64(s.CurValue-s.LastValue) / s.ElapsedTime
				}
			}
		} else {
			s.Compute = func() {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue < s.LastValue {
					s.CookedValue = float64(math.MaxInt32 - s.LastValue + s.CurValue)
				} else {
					s.CookedValue = float64(s.CurValue - s.LastValue)
				}
			}
		}
	case "COUNTER64":
		s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
			//log.Debugf("========================================>COUNTER64: first time :%s ", s.RealOID)
			//first time only set values and reassign itself to the complete method
			val := pduVal2Int64(pdu)
			s.CurValue = val
			s.CurTime = now
			s.setRawData = func(pdu gosnmp.SnmpPDU, now time.Time) {
				//log.Debugf("========================================>COUNTER64: the other time:%s", s.RealOID)
				val := pduVal2Int64(pdu)
				s.LastTime = s.CurTime
				s.LastValue = s.CurValue
				s.CurValue = val
				s.CurTime = now
				s.Compute()
				s.Scale()
			}
		}
		if s.cfg.GetRate == true {
			s.Compute = func() {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				//duration := s.CurTime.Sub(s.LastTime)
				if s.CurValue < s.LastValue {
					s.CookedValue = float64(math.MaxInt64-s.LastValue+s.CurValue) / s.ElapsedTime
				} else {
					s.CookedValue = float64(s.CurValue-s.LastValue) / s.ElapsedTime
				}
			}
		} else {
			s.Compute = func() {
				s.ElapsedTime = s.CurTime.Sub(s.LastTime).Seconds()
				if s.CurValue < s.LastValue {
					s.CookedValue = float64(math.MaxInt64 - s.LastValue + s.CurValue)
				} else {
					s.CookedValue = float64(s.CurValue - s.LastValue)
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
