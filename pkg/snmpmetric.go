package main

import (
	"errors"
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
	ABSOLUTE //It is intended for counters which are reset upon reading. In effect, the type is very similar to GAUGE except that the value is an (unsigned) integer
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
	case "ABSOLUTE":
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
	cookedValue float64
	curValue    int64
	lastValue   int64
	curTime     time.Time
	lastTime    time.Time
	Compute     func()
	setRawData  func(val int64, now time.Time)
	realOID     string
}

func (s *SnmpMetric) Init() error {
	switch s.cfg.DataSrcType {
	case "GAUGE":
		s.setRawData = func(val int64, now time.Time) {
			s.cookedValue = float64(val)
			s.curTime = now
		}
		s.Compute = func() {
		}
	case "INTEGER":
		s.setRawData = func(val int64, now time.Time) {
			s.cookedValue = float64(val)
			s.curTime = now
		}
		s.Compute = func() {
		}
	case "COUNTER32":
		s.setRawData = func(val int64, now time.Time) {
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
		s.setRawData = func(val int64, now time.Time) {
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

	case "ABSOLUTE":
		//TODO
	}
	return nil
}
