package device

import (
	"github.com/Sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"sync"
	"time"
)

type DevStatType uint

const (
	SnmpGetQueries        = 0 //Query stats
	SnmpWalkQueries       = 1
	SnmpGetErrors         = 2
	SnmpWalkErrors        = 3
	SnmpQueryTimeouts     = 4
	SnmpOIDGetAll         = 5 //Snmp metrics gathered stats
	SnmpOIDGetProcessed   = 6
	SnmpOIDGetErrors      = 7
	EvalMetricsAll        = 8 // Eval metrics
	EvalMetricsOk         = 9
	EvalMetricsErrors     = 10
	MetricSent            = 11 // Sent metrics stats
	MetricSentErrors      = 12
	MeasurementSent       = 13
	MeasurementSentErrors = 14
	CicleGatherStartTime  = 15 //Process Elapsed Time stats
	CicleGatherDuration   = 16
	FilterStartTime       = 17
	FilterDuration        = 18
	BackEndSentStartTime  = 19 // this counter en concurrent sent mode can be confused
	BackEndSentDuration   = 20
	DevStatTypeSize       = 21
)

// DevStat minimal info to show users
type DevStat struct {
	//ID
	id     string
	tagMap map[string]string
	//Control
	log     *logrus.Logger
	selfmon *selfmon.SelfMon
	mutex   sync.Mutex

	//Counter Statistics
	Counters []interface{}

	//device state
	ReloadLoopsPending int
	DeviceActive       bool
	DeviceConnected    bool
	//extra measurement statistics
	NumMeasurements int
	NumMetrics      int
}

// Init initializes the device stat object
func (s *DevStat) Init(id string, tm map[string]string, l *logrus.Logger) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.id = id
	s.tagMap = tm
	s.log = l
	s.Counters = make([]interface{}, DevStatTypeSize)
	s.Counters[SnmpGetQueries] = 0
	s.Counters[SnmpWalkQueries] = 0
	s.Counters[SnmpGetErrors] = 0
	s.Counters[SnmpWalkErrors] = 0
	s.Counters[SnmpQueryTimeouts] = 0
	s.Counters[SnmpOIDGetAll] = 0
	s.Counters[SnmpOIDGetProcessed] = 0
	s.Counters[SnmpOIDGetErrors] = 0
	s.Counters[EvalMetricsAll] = 0
	s.Counters[EvalMetricsOk] = 0
	s.Counters[EvalMetricsErrors] = 0
	s.Counters[MetricSent] = 0
	s.Counters[MeasurementSent] = 0
	s.Counters[MetricSentErrors] = 0
	s.Counters[MeasurementSentErrors] = 0
	s.Counters[CicleGatherStartTime] = 0
	s.Counters[CicleGatherDuration] = 0.0
	s.Counters[FilterStartTime] = 0
	s.Counters[FilterDuration] = 0.0
	s.Counters[BackEndSentStartTime] = 0
	s.Counters[BackEndSentDuration] = 0.0
}

func (s *DevStat) reset() {
	for k, val := range s.Counters {
		switch v := val.(type) {
		case string:
			s.Counters[k] = ""
		case int32, int64, int:
			s.Counters[k] = 0
		case float64, float32:
			s.Counters[k] = 0.0
		default:
			s.log.Warnf("unknown typpe for counter %#v", v)
		}
	}
}

func (s *DevStat) getMetricFields() map[string]interface{} {
	fields := map[string]interface{}{
		/*0*/ //"snmp_get_queries": s.Counters[SnmpGetQueries],
		/*1*/ //"snmp_walk_queries": s.Counters[SnmpWalkQueries],
		/*2*/ //"snmp_get_errors": s.Counters[SnmpGetErrors],
		/*3*/ //"snmp_walk_errors": s.Counters[SnmpWalkErrors],
		/*4*/ //"snmp_query_timeouts": s.Counters[SnmpQueryTimeouts],
		/*5*/ "snmp_oid_get_all": s.Counters[SnmpOIDGetAll],
		/*6*/ "snmp_oid_get_processed": s.Counters[SnmpOIDGetProcessed],
		/*7*/ "snmp_oid_get_errors": s.Counters[SnmpOIDGetErrors],
		/*8*/ //"eval_metric_all": s.Counters[EvalMetricsAll],
		/*9*/ //"eval_metric_ok": s.Counters[EvalMetricsOk],
		/*10*/ //"eval_metric_errors": s.Counters[EvalMetricsErrors],
		/*11*/ "metric_sent": s.Counters[MetricSent],
		/*12*/ "metric_sent_errors": s.Counters[MetricSentErrors],
		/*13*/ "measurement_sent": s.Counters[MeasurementSent],
		/*14*/ "measurement_sent_errors": s.Counters[MeasurementSentErrors],
		/*15*/ "cicle_gather_start_time": s.Counters[CicleGatherStartTime],
		/*16*/ "cicle_gather_duration": s.Counters[CicleGatherDuration],
		/*17*/ "filter_start_time": s.Counters[FilterStartTime],
		/*18*/ "filter_duration": s.Counters[FilterDuration],
		/*19*/ "backend_sent_start_time": s.Counters[BackEndSentStartTime],
		/*20*/ "backend_sent_duration": s.Counters[BackEndSentDuration],
	}
	return fields
}

// SetSelfMonitoring set the ouput device where send monitoring metrics
func (s *DevStat) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.selfmon = cfg
}

// ThSafeCopy get a new object with public data copied in thread safe way
func (s *DevStat) ThSafeCopy() *DevStat {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	st := &DevStat{}
	st.Init(s.id, s.tagMap, s.log)
	for k, v := range s.Counters {
		st.Counters[k] = v
	}
	return st
}

// Send send data to the selfmon device
func (s *DevStat) Send() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.log.Infof("STATS SNMP GET: snmp pooling took [%f seconds] SNMP: Gets [%d] , Processed [%d], Errors [%d]", s.Counters[CicleGatherDuration], s.Counters[SnmpOIDGetAll], s.Counters[SnmpOIDGetProcessed], s.Counters[SnmpOIDGetErrors])
	s.log.Infof("STATS SNMP FILTER: filter pooling took [%f seconds] ", s.Counters[FilterDuration])
	s.log.Infof("STATS INFLUX: influx send took [%f seconds]", s.Counters[BackEndSentDuration])
	if s.selfmon != nil {
		s.selfmon.AddDeviceMetrics(s.id, s.getMetricFields(), s.tagMap)
	}
}

// ResetCounters initialize metric counters
func (s *DevStat) ResetCounters() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.reset()
}

// CounterInc n values to the counter set by id
func (s *DevStat) CounterInc(id DevStatType, n int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[id] = s.Counters[id].(int) + int(n)
}

func (s *DevStat) AddMeasStats(mets int64, mete int64, meass int64, mease int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[MetricSent] = s.Counters[MetricSent].(int) + int(mets)
	s.Counters[MetricSentErrors] = s.Counters[MetricSentErrors].(int) + int(mete)
	s.Counters[MeasurementSent] = s.Counters[MeasurementSent].(int) + int(meass)
	s.Counters[MeasurementSentErrors] = s.Counters[MeasurementSentErrors].(int) + int(mease)
}

func (s *DevStat) UpdateSnmpGetStats(g int64, p int64, e int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[SnmpOIDGetAll] = s.Counters[SnmpOIDGetAll].(int) + int(g)
	s.Counters[SnmpOIDGetProcessed] = s.Counters[SnmpOIDGetProcessed].(int) + int(p)
	s.Counters[SnmpOIDGetErrors] = s.Counters[SnmpOIDGetErrors].(int) + int(e)
}

// Update Gather Duration stats
func (s *DevStat) SetGatherDuration(start time.Time, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[CicleGatherStartTime] = start.Unix()
	s.Counters[CicleGatherDuration] = duration.Seconds()
}

// Update Gather Duration stats
func (s *DevStat) AddSentDuration(start time.Time, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//only register the first start time on concurrent mode
	if s.Counters[BackEndSentStartTime] == 0 {
		s.Counters[BackEndSentStartTime] = start.Unix()
	}
	s.Counters[BackEndSentDuration] = s.Counters[BackEndSentDuration].(float64) + duration.Seconds()
}

func (s *DevStat) SetFltUpdateStats(start time.Time, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Counters[FilterStartTime] = start.Unix()
	s.Counters[FilterDuration] = duration.Seconds()
}
