package selfmon

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	log *logrus.Logger
)

// SetLogger set log output
func SetLogger(l *logrus.Logger) {
	log = l
}

//SelfMon configuration for self monitoring
type SelfMon struct {
	cfg                 *config.SelfMonConfig
	Influx              *output.InfluxDB
	runtimeStatsRunning bool
	TagMap              map[string]string
	Fields              map[string]interface{}
	bps                 *client.BatchPoints
	chExit              chan bool
	mutex               sync.Mutex
	rt_meas_name        string
	gvm_meas_name       string
	initialized         bool
	imutex              sync.Mutex
}

// NewNotInit create strut without initialization
func NewNotInit(c *config.SelfMonConfig) *SelfMon {
	return &SelfMon{cfg: c}
}

// Init Initialize the Object data and check for consistence
func (sm *SelfMon) Init() {
	if sm.CheckAndSetInitialized() == true {
		log.Info("Self monitoring thread  already Initialized (skipping Initialization)")
		return
	}

	//Init extra tags
	if len(sm.cfg.ExtraTags) > 0 {
		sm.TagMap = make(map[string]string)
		for _, tag := range sm.cfg.ExtraTags {
			s := strings.Split(tag, "=")
			if len(s) == 2 {
				key, value := s[0], s[1]
				sm.TagMap[key] = value
			} else {
				log.Errorf("Error on tag  definition TAG=VALUE [ %s ] for SelfMon", tag)
			}
		}
	}
	// Measurement Names
	sm.rt_meas_name = "selfmon_rt"
	sm.gvm_meas_name = "selfmon_gvm"
	if len(sm.cfg.Prefix) > 0 {
		sm.rt_meas_name = fmt.Sprintf("%sselfmon_rt", sm.cfg.Prefix)
		sm.gvm_meas_name = fmt.Sprintf("%sselfmon_gvm", sm.cfg.Prefix)
	}

	//Init Measurment Fields.
	sm.Fields = map[string]interface{}{
		"runtime_goroutines":    0.0,
		"mem.alloc":             0.0,
		"mem.mallocs":           0.0,
		"mem.frees":             0.0,
		"mem.heapAlloc":         0.0,
		"mem.stackInuse":        0.0,
		"gc.total_pause_ns":     0.0,
		"gc.pause_per_second":   0.0,
		"gc.pause_per_interval": 0.0,
		"gc.gc_per_second":      0.0,
		"gc.gc_per_interval":    0.0,
	}
	sm.chExit = make(chan bool)

}

// CheckAndSetInitialized set
func (sm *SelfMon) CheckAndSetInitialized() bool {
	sm.imutex.Lock()
	defer sm.imutex.Unlock()
	retval := sm.initialized
	sm.initialized = true
	return retval
}

// CheckAndUnSetInitialized unset
func (sm *SelfMon) CheckAndUnSetInitialized() bool {
	sm.imutex.Lock()
	defer sm.imutex.Unlock()
	retval := sm.initialized
	sm.initialized = false
	return retval
}

// IsInitialized check if this thread is already working
func (sm *SelfMon) IsInitialized() bool {
	sm.imutex.Lock()
	defer sm.imutex.Unlock()
	return sm.initialized
}

// SetOutput set out data
func (sm *SelfMon) SetOutput(val *output.InfluxDB) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.Influx = val
	//Creating a bachpoint to begin writing data
	sm.bps, _ = sm.Influx.BP()
}

func (sm *SelfMon) sendData() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.Influx.Send(sm.bps)
	//BatchPoint Init again
	sm.bps, _ = sm.Influx.BP()
}

func (sm *SelfMon) addDataPoint(pt *client.Point) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if sm.bps != nil {
		(*sm.bps).AddPoint(pt)
	}
}

// AddDeviceMetrics add data from devices
func (sm *SelfMon) AddDeviceMetrics(deviceid string, fields map[string]interface{}) {
	if !sm.IsInitialized() {
		return
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	tagMap := make(map[string]string)
	for k, v := range sm.TagMap {
		tagMap[k] = v
	}
	tagMap["device"] = deviceid
	now := time.Now()
	pt, _ := client.NewPoint(
		sm.rt_meas_name,
		tagMap,
		fields,
		now)

	(*sm.bps).AddPoint(pt)
}

func (sm *SelfMon) End() {
	if sm.CheckAndUnSetInitialized() {
		close(sm.chExit)
	}
}

// StartGather for stopping selfmonitori goroutine
func (sm *SelfMon) StartGather(wg *sync.WaitGroup) {
	if !sm.cfg.Enabled {
		log.Info("SELFMON: disabled, skipping start gather")
		return
	}
	if sm.runtimeStatsRunning {
		log.Error("SELFMON:Runtime stats is already running")
		return
	}

	sm.runtimeStatsRunning = true

	go sm.reportRuntimeStats(wg)
}

// StopGather for stopping selfmonitori goroutine
func (sm *SelfMon) StopGather() {
	if sm.cfg.Enabled {
		sm.chExit <- true
	}
}

func (sm *SelfMon) reportRuntimeStats(wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	log.Info("SELFMON: Beginning  selfmonitor process for device")

	memStats := &runtime.MemStats{}
	lastSampleTime := time.Now()
	var lastPauseNs uint64 = 0
	var lastNumGc uint32 = 0

	nsInMs := float64(time.Millisecond)
	s := time.Tick(time.Duration(sm.cfg.Freq) * time.Second)
	for {

		runtime.ReadMemStats(memStats)

		now := time.Now()

		sm.Fields["runtime_goroutines"] = float64(runtime.NumGoroutine())
		sm.Fields["mem.alloc"] = float64(memStats.Alloc)
		sm.Fields["mem.mallocs"] = float64(memStats.Mallocs)
		sm.Fields["mem.frees"] = float64(memStats.Frees)
		sm.Fields["gc.total_pause_ns"] = float64(memStats.PauseTotalNs) / nsInMs
		sm.Fields["mem.heapAlloc"] = float64(memStats.HeapAlloc)
		sm.Fields["mem.stackInuse"] = float64(memStats.StackInuse)

		if lastPauseNs > 0 {
			pauseSinceLastSample := memStats.PauseTotalNs - lastPauseNs
			sm.Fields["gc.pause_per_second"] = float64(pauseSinceLastSample) / nsInMs / time.Duration(sm.cfg.Freq).Seconds()
			sm.Fields["gc.pause_per_interval"] = float64(pauseSinceLastSample) / nsInMs
		}
		lastPauseNs = memStats.PauseTotalNs

		countGc := int(memStats.NumGC - lastNumGc)
		if lastNumGc > 0 {
			diff := float64(countGc)
			diffTime := now.Sub(lastSampleTime).Seconds()
			sm.Fields["gc.gc_per_second"] = diff / diffTime
			sm.Fields["gc.gc_per_interval"] = diff
		}

		if countGc > 0 {
			if countGc > 256 {
				log.Warn("We're missing some gc pause times")
				countGc = 256
			}
			var totalPause float64 = 0
			for i := 0; i < countGc; i++ {
				idx := int((memStats.NumGC-uint32(i))+255) % 256
				pause := float64(memStats.PauseNs[idx])
				totalPause += pause
				//	sm.Report(fmt.Sprintf("%s.memory.gc.pause", prefix), pause/nsInMs, now)
			}
			//sm.Report(fmt.Sprintf("%s.memory.gc.pause_per_interval", prefix), totalPause/nsInMs, now)
			sm.Fields["gc.pause_per_interval"] = totalPause / nsInMs
		}

		lastNumGc = memStats.NumGC
		lastSampleTime = now

		pt, _ := client.NewPoint(
			sm.gvm_meas_name,
			sm.TagMap,
			sm.Fields,
			now,
		)

		//add data to the batchpoint
		sm.addDataPoint(pt)
		//BatchPoint Send
		sm.sendData()

	LOOP:
		for {
			select {
			case <-s:
				//log.Infof("SELFMON: breaking LOOP  ")
				break LOOP
			case <-sm.chExit:
				log.Infof("SELFMON: EXIT from SelfMonitoring Gather process ")
				sm.runtimeStatsRunning = false
				return
			}
		}
	}

}
