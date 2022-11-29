package selfmon

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	metric "github.com/influxdata/telegraf/metric"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/config"
)

var log *logrus.Logger

// SetLogger set log output
func SetLogger(l *logrus.Logger) {
	log = l
}

// SelfMon configuration for self monitoring
// selfmon allows to gather data from all outputs and send metrics to a definid output
// selfmon should run within a goroutine and it should send metrics using the default backend
// that should be already defined as a SinkDB with an attached backend
type SelfMon struct {
	cfg                 *config.SelfMonConfig
	Output              *output.SinkDB
	OutDBs              map[string]*output.SinkDB // needed to get statistics
	runtimeStatsRunning bool
	TagMap              map[string]string
	tick                *time.Ticker

	chExit      chan bool
	mutex       sync.Mutex
	RtMeasName  string // devices measurement name
	GvmMeasName string // Self agent GoVirtualMachine measurement name
	OutMeasName string // Output DB's measurement name
	initialized bool
	imutex      sync.Mutex
	// memory for GVM data colletion
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
	if sm.CheckAndSetInitialized() {
		log.Info("Self monitoring thread  already Initialized (skipping Initialization)")
		return
	}

	// declare all the available outputs to gather metrics from
	// Init extra tags
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
	// initialize ticker with the current time
	// we should review if we should sync with the output ticker as it can send metrics witha freq delay
	// if the output sent ticks before the selfmon ticker
	if sm.cfg.Freq <= 0 {
		log.Infof("No freq defined, defaulting to 60s")
		sm.cfg.Freq = 60
	}
	sm.tick = time.NewTicker(time.Duration(sm.cfg.Freq) * time.Second)
}

// SetOutDB set the output devices for query its statistics
func (sm *SelfMon) SetOutDB(odb map[string]*output.SinkDB) {
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
func (sm *SelfMon) SetOutput(val *output.SinkDB) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.Output = val
}

// AddDeviceMetrics add data from devices
func (sm *SelfMon) AddMetrics(t string, id string, fields map[string]interface{}, devtags, statustags map[string]string) {
	if !sm.IsInitialized() {
		return
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	var tmetrics []telegraf.Metric

	// Selfmon tags
	tagMap := make(map[string]string)
	for k, v := range sm.TagMap {
		tagMap[k] = v
	}
	// device user configured extra tags (only if inherited)
	if sm.cfg.InheritDeviceTags {
		for k, v := range devtags {
			tagMap[k] = v
		}
	}
	// status tags for device
	for k, v := range statustags {
		tagMap[k] = v
	}

	switch t {
	case "measurement":
		tagMap["meas_name"] = id
	case "device":
		tagMap["device"] = id
	}

	tagMap["type"] = t

	now := time.Now()
	tmetric := metric.New(sm.RtMeasName, tagMap, fields, now)
	// Validate that at least len of fields > 0 and len of tags > 0
	if len(fields) == 0 {
		log.Warnf("error, empty fields: [%d]", len(fields))
		return
	}
	tmetrics = append(tmetrics, tmetric)
	lmet, err := sm.Output.SendToBuffer(tmetrics, "selfmon"+sm.RtMeasName)
	if err != nil {
		log.Errorf("unable to send metrics to the buffer - dropped metrics: %d, err: %s", lmet, err)
	}
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

// getOutDBStats gather stats from all outputs, we need to query them and add the metrics to the output buffer
func (sm *SelfMon) getOutDBStats() {
	now := time.Now()
	var tmetrics []telegraf.Metric

	for dbname, sinkdb := range sm.OutDBs {
		stats := sinkdb.GetResetStats()

		tags := make(map[string]string)
		fields := make(map[string]interface{})

		for k, v := range sm.TagMap {
			tags[k] = v
		}
		tags["outdb"] = dbname

		fields["fields_sent"] = stats.FieldSent
		fields["fields_sent_max"] = stats.FieldSentMax

		fields["points_sent"] = stats.PSent
		fields["points_dropped"] = stats.PDropped
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

		tmetric := metric.New(sm.OutMeasName, tags, fields, now)
		// Validate that at least len of fields > 0 and len of tags > 0
		if len(tags) == 0 || len(fields) == 0 {
			log.Warnf("error, empty tags [%d] or fields: [%d]", len(tags), len(fields))
			continue
		}
		tmetrics = append(tmetrics, tmetric)
	}
	sm.Output.SendToBuffer(tmetrics, "selfmon"+sm.OutMeasName)
}

func (sm *SelfMon) getRuntimeStats() {

	var tmetrics []telegraf.Metric

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

	// HEAP

	fields["mem.heapAlloc"] = float64(memStats.HeapAlloc)       // HeapAlloc is bytes of allocated heap objects.
	fields["mem.heapSys"] = float64(memStats.HeapSys)           // HeapSys is bytes of heap memory obtained from the OS.
	fields["mem.heapIdle"] = float64(memStats.HeapIdle)         // HeapIdle is bytes in idle (unused) spans.
	fields["mem.heapInUse"] = float64(memStats.HeapInuse)       // HeapInuse is bytes in in-use spans.
	fields["mem.heapReleased"] = float64(memStats.HeapReleased) // HeapReleased is bytes of physical memory returned to the OS.
	fields["mem.heapObjects"] = float64(memStats.HeapReleased)  // HeapObjects is the number of allocated heap objects.

	// STACK/MSPAN/MCACHE

	fields["mem.stackInuse"] = float64(memStats.StackInuse)   // StackInuse is bytes in stack spans.
	fields["mem.mSpanInuse"] = float64(memStats.MSpanInuse)   // MSpanInuse is bytes of allocated mspan structures.
	fields["mem.mCacheInuse"] = float64(memStats.MCacheInuse) // MCacheInuse is bytes of allocated mcache structures.

	// Pause Count
	fields["gc.total_pause_ns"] = float64(memStats.PauseTotalNs) / nsInMs

	if sm.lastPauseNs > 0 {
		pauseSinceLastSample := memStats.PauseTotalNs - sm.lastPauseNs
		pauseInterval := float64(pauseSinceLastSample) / nsInMs
		fields["gc.pause_per_interval"] = pauseInterval
		fields["gc.pause_per_second"] = pauseInterval / diffTime
		//		log.Debugf("SELFMON:Diftime(%f) |PAUSE X INTERVAL: %f | PAUSE X SECOND %f", diffTime, pauseInterval, pauseInterval/diffTime)
	}
	sm.lastPauseNs = memStats.PauseTotalNs

	// GC Count
	countGc := int(memStats.NumGC - sm.lastNumGc)
	if sm.lastNumGc > 0 {
		diff := float64(countGc)
		fields["gc.gc_per_second"] = diff / diffTime
		fields["gc.gc_per_interval"] = diff
	}
	sm.lastNumGc = memStats.NumGC
	sm.lastSampleTime = now

	tmetric := metric.New(sm.GvmMeasName, sm.TagMap, fields, now)
	// Validate that at least len of fields > 0 and len of tags > 0
	if len(fields) == 0 {
		log.Warnf("error, empty fields: [%d]", len(fields))
	}
	tmetrics = append(tmetrics, tmetric)
	sm.Output.SendToBuffer(tmetrics, "selfmon"+sm.GvmMeasName)
}

func (sm *SelfMon) reportStats(wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	log.Info("SELFMON: Beginning  selfmonitor process for device")

	sm.lastSampleTime = time.Now()
	for {
		select {
		case <-sm.tick.C:
			// Get BVM stats
			sm.getRuntimeStats()
			sm.getOutDBStats()
		case <-sm.chExit:
			log.Infof("SELFMON: EXIT from SelfMonitoring Gather process ")
			sm.runtimeStatsRunning = false
			return
		}
	}
}
