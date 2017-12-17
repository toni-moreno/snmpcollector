package selfmon

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/config"
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
	OutDBs              map[string]*output.InfluxDB //needed to get statistics
	runtimeStatsRunning bool
	TagMap              map[string]string
	Fields              map[string]interface{}
	bps                 *client.BatchPoints
	chExit              chan bool
	mutex               sync.Mutex
	RtMeasName          string //devices measurement name
	GvmMeasName         string //Self agent GoVirtualMachine measurement name
	OutMeasName         string //Output DB's measurement name
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
	sm.OutDBs = make(map[string]*output.InfluxDB)

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

	log.Infof("Self monitoring TAGS inheritance set to : %t", sm.cfg.InheritDeviceTags)

	// Measurement Names
	sm.RtMeasName = "selfmon_device_stats"
	sm.GvmMeasName = "selfmon_gvm"
	sm.OutMeasName = "selfmon_outdb_stats"

	if len(sm.cfg.Prefix) > 0 {
		sm.RtMeasName = fmt.Sprintf("%sselfmon_device_stats", sm.cfg.Prefix)
		sm.GvmMeasName = fmt.Sprintf("%sselfmon_gvm", sm.cfg.Prefix)
		sm.OutMeasName = fmt.Sprintf("%sselfmon_outdb_stats", sm.cfg.Prefix)
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

// SetOutDB set the output devices for query its statistics
func (sm *SelfMon) SetOutDB(odb map[string]*output.InfluxDB) {
	sm.OutDBs = odb
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
func (sm *SelfMon) AddDeviceMetrics(deviceid string, fields map[string]interface{}, devtags map[string]string) {
	if !sm.IsInitialized() {
		return
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	tagMap := make(map[string]string)
	for k, v := range sm.TagMap {
		tagMap[k] = v
	}
	if sm.cfg.InheritDeviceTags {
		for k, v := range devtags {
			tagMap[k] = v
		}
	}

	tagMap["device"] = deviceid
	now := time.Now()
	pt, err := client.NewPoint(
		sm.RtMeasName,
		tagMap,
		fields,
		now)
	if err != nil {
		log.Warnf("Error on compute Stats data Point %+v for device %s: Error:%s", fields, deviceid, err)
		return
	}

	(*sm.bps).AddPoint(pt)
}

// End Release the SelMon Object
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

	go sm.reportStats(wg)
}

// StopGather for stopping selfmonitori goroutine
func (sm *SelfMon) StopGather() {
	if sm.cfg.Enabled {
		sm.chExit <- true
	}
}

func (sm *SelfMon) getOutDBStats() {

	now := time.Now()

	for dbname, db := range sm.OutDBs {

		stats := db.GetResetStats()

		tm := make(map[string]string)
		fields := make(map[string]interface{})

		for k, v := range sm.TagMap {
			tm[k] = v
		}
		tm["outdb"] = dbname

		fields["points_sent"] = stats.PSent
		fields["points_sent_max"] = stats.PSentMax

		fields["write_sent"] = stats.WriteSent
		fields["write_error"] = stats.WriteErrors
		sec := stats.WriteTime.Seconds()
		fields["write_time"] = sec
		fields["write_time_max"] = stats.WriteTimeMax.Seconds()

		if stats.WriteSent > 0 {
			fields["points_sent_avg"] = float64(stats.PSent) / float64(stats.WriteSent)
			fields["write_time_avg"] = sec / float64(stats.WriteSent)
		}

		pt, err := client.NewPoint(sm.OutMeasName, tm, fields, now)
		if err != nil {
			log.Warnf("Error on compute Stats data Point %+v for database %s: Error:%s", fields, dbname, err)
			return
		}

		//add data to the batchpoint
		sm.addDataPoint(pt)
	}

}

func (sm *SelfMon) getRuntimeStats() {

	lastSampleTime := time.Now()
	var lastPauseNs uint64
	var lastNumGc uint32

	nsInMs := float64(time.Millisecond)
	memStats := &runtime.MemStats{}
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
		var totalPause float64
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

	pt, err := client.NewPoint(
		sm.GvmMeasName,
		sm.TagMap,
		sm.Fields,
		now,
	)
	if err != nil {
		log.Warnf("Error on compute Stats data Point %+v for GVB : Error:%s", sm.Fields, err)
		return
	}

	//add data to the batchpoint
	sm.addDataPoint(pt)

}

func (sm *SelfMon) reportStats(wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	log.Info("SELFMON: Beginning  selfmonitor process for device")

	s := time.Tick(time.Duration(sm.cfg.Freq) * time.Second)
	for {
		//Get BVM stats
		sm.getRuntimeStats()
		//
		sm.getOutDBStats()
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
