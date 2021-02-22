package selfmon

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
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
	bps                 *client.BatchPoints
	chExit              chan bool
	mutex               sync.Mutex
	RtMeasName          string //devices measurement name
	GvmMeasName         string //Self agent GoVirtualMachine measurement name
	OutMeasName         string //Output DB's measurement name
	initialized         bool
	imutex              sync.Mutex
	//memory for GVM data colletion
	lastSampleTime time.Time
	lastPauseNs    uint64
	lastNumGc      uint32
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

		fields["fields_sent"] = stats.FieldSent
		fields["fields_sent_max"] = stats.FieldSentMax

		fields["points_sent"] = stats.PSent
		fields["points_sent_max"] = stats.PSentMax

		fields["write_sent"] = stats.WriteSent
		fields["write_error"] = stats.WriteErrors
		sec := stats.WriteTime.Seconds()
		fields["write_time"] = sec
		fields["write_time_max"] = stats.WriteTimeMax.Seconds()

		fields["buffer_percent_used"] = stats.BufferPercentUsed

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

	nsInMs := float64(time.Millisecond)
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	fields := make(map[string]interface{})

	now := time.Now()
	diffTime := now.Sub(sm.lastSampleTime).Seconds()

	fields["runtime_goroutines"] = float64(runtime.NumGoroutine())
	fields["mem.alloc"] = float64(memStats.Alloc)
	fields["mem.mallocs"] = float64(memStats.Mallocs)
	fields["mem.frees"] = float64(memStats.Frees)
	fields["mem.sys"] = float64(memStats.Sys)

	//HEAP

	fields["mem.heapAlloc"] = float64(memStats.HeapAlloc)       //HeapAlloc is bytes of allocated heap objects.
	fields["mem.heapSys"] = float64(memStats.HeapSys)           // HeapSys is bytes of heap memory obtained from the OS.
	fields["mem.heapIdle"] = float64(memStats.HeapIdle)         // HeapIdle is bytes in idle (unused) spans.
	fields["mem.heapInUse"] = float64(memStats.HeapInuse)       // HeapInuse is bytes in in-use spans.
	fields["mem.heapReleased"] = float64(memStats.HeapReleased) // HeapReleased is bytes of physical memory returned to the OS.
	fields["mem.heapObjects"] = float64(memStats.HeapReleased)  // HeapObjects is the number of allocated heap objects.

	//STACK/MSPAN/MCACHE

	fields["mem.stackInuse"] = float64(memStats.StackInuse)   // StackInuse is bytes in stack spans.
	fields["mem.mSpanInuse"] = float64(memStats.MSpanInuse)   // MSpanInuse is bytes of allocated mspan structures.
	fields["mem.mCacheInuse"] = float64(memStats.MCacheInuse) // MCacheInuse is bytes of allocated mcache structures.

	//Pause Count
	fields["gc.total_pause_ns"] = float64(memStats.PauseTotalNs) / nsInMs

	if sm.lastPauseNs > 0 {
		pauseSinceLastSample := memStats.PauseTotalNs - sm.lastPauseNs
		pauseInterval := float64(pauseSinceLastSample) / nsInMs
		fields["gc.pause_per_interval"] = pauseInterval
		fields["gc.pause_per_second"] = pauseInterval / diffTime
		//		log.Debugf("SELFMON:Diftime(%f) |PAUSE X INTERVAL: %f | PAUSE X SECOND %f", diffTime, pauseInterval, pauseInterval/diffTime)
	}
	sm.lastPauseNs = memStats.PauseTotalNs

	//GC Count
	countGc := int(memStats.NumGC - sm.lastNumGc)
	if sm.lastNumGc > 0 {
		diff := float64(countGc)
		fields["gc.gc_per_second"] = diff / diffTime
		fields["gc.gc_per_interval"] = diff
	}
	sm.lastNumGc = memStats.NumGC

	sm.lastSampleTime = now

	pt, err := client.NewPoint(
		sm.GvmMeasName,
		sm.TagMap,
		fields,
		now,
	)
	if err != nil {
		log.Warnf("Error on compute Stats data Point %+v for GVB : Error:%s", fields, err)
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
	sm.lastSampleTime = time.Now()
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
