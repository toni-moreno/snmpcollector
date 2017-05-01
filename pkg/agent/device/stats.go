package device

import (
	"github.com/Sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"sync"
	"time"
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
	Requests int64
	Gets     int64
	Errors   int64

	//time statistics
	// t - Gathering all snmpdata
	GatherTime     time.Time
	GatherDuration time.Duration
	// t - Apply filters on measurements
	FltUpdateTime     time.Time
	FltUpdateDuration time.Duration
	// t - Send data over output backend
	SentDuration time.Duration

	//device state
	ReloadLoopsPending int
	DeviceActive       bool
	DeviceConnected    bool
	//extra measurement statistics
	NumMeasurements int
	NumMetrics      int
}

func (s *DevStat) Init(id string, tm map[string]string, l *logrus.Logger) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.id = id
	s.tagMap = tm
	s.log = l

}

func (s *DevStat) reset() {
	s.Gets = 0
	s.Errors = 0
	s.SentDuration = 0
}

func (s *DevStat) getMetricFields() map[string]interface{} {
	fields := map[string]interface{}{
		"process_t": s.GatherDuration.Seconds(),
		"getsent":   s.Gets,
		"geterror":  s.Errors,
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
	return &DevStat{
		Requests: s.Requests,
		Gets:     s.Gets,
		Errors:   s.Errors,
		//ReloadLoopsPending:    s.ReloadLoopsPending,
		GatherTime:        s.GatherTime,
		GatherDuration:    s.GatherDuration,
		FltUpdateTime:     s.FltUpdateTime,
		FltUpdateDuration: s.FltUpdateDuration,
		SentDuration:      s.SentDuration,
	}

}

// Send send data to the selfmon device
func (s *DevStat) Send() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.log.Infof("STATS: snmp pooling took [%s] SNMP: Gets [%d] Errors [%d]", s.GatherDuration, s.Gets, s.Errors)
	s.log.Infof("STATS: influx send took [%s]", s.SentDuration)
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

// AddRequests add num request
func (s *DevStat) AddRequests(n int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Requests += n
}

// AddGets update Gets
func (s *DevStat) AddGets(n int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Gets += n
}

// AddErrors update errors
func (s *DevStat) AddErrors(n int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Errors += n
}

// Update Gather Duration stats
func (s *DevStat) SetGatherDuration(start time.Time, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.GatherTime = start
	s.GatherDuration = duration
}

// Update Gather Duration stats
func (s *DevStat) AddSentDuration(duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.SentDuration += duration
}

func (s *DevStat) SetFltUpdateStats(start time.Time, duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.FltUpdateTime = start
	s.FltUpdateDuration = duration
}
