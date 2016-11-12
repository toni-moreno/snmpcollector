package main

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"runtime"
	"strings"
	"sync"
	"time"
)

//SelfMonConfig configuration for self monitoring
type SelfMonConfig struct {
	Enabled             bool     `toml:"enabled"`
	Freq                int      `toml:"freq"`
	Prefix              string   `toml:"prefix"`
	ExtraTags           []string `toml:"extra-tags"`
	Influx              *InfluxDB
	runtimeStatsRunning bool
	TagMap              map[string]string
	Fields              map[string]interface{}
	bps                 *client.BatchPoints
}

func (sm *SelfMonConfig) Init() {
	//Init extra tags
	if len(sm.ExtraTags) > 0 {
		sm.TagMap = make(map[string]string)
		for _, tag := range sm.ExtraTags {
			s := strings.Split(tag, "=")
			if len(s) == 2 {
				key, value := s[0], s[1]
				sm.TagMap[key] = value
			} else {
				log.Errorf("Error on tag  definition TAG=VALUE [ %s ] for SelfMon", tag)
			}
		}
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
}

func (sm *SelfMonConfig) ReportStats(wg *sync.WaitGroup) {
	if sm.runtimeStatsRunning {
		log.Error("Runtime stats is already running")
		return
	}

	sm.runtimeStatsRunning = true
	go sm.reportRuntimeStats(time.Duration(sm.Freq*1000000000), wg)
}

func (sm *SelfMonConfig) reportRuntimeStats(sleep time.Duration, wg *sync.WaitGroup) {

	//sm.bps = sm.Influx.BP()

	memStats := &runtime.MemStats{}
	lastSampleTime := time.Now()
	var lastPauseNs uint64 = 0
	var lastNumGc uint32 = 0
	prefix := sm.Prefix

	nsInMs := float64(time.Millisecond)

	for sm.runtimeStatsRunning {
		//BatchPoint Init
		sm.bps = sm.Influx.BP()

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
			sm.Fields["gc.pause_per_second"] = float64(pauseSinceLastSample) / nsInMs / sleep.Seconds()
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
		metricname := "selmon_gvm"
		if len(prefix) > 0 {
			metricname = fmt.Sprintf("%sselfmon_gvm", prefix)
		}
		pt, _ := client.NewPoint(
			metricname,
			sm.TagMap,
			sm.Fields,
			now,
		)
		(*sm.bps).AddPoint(pt)
		//BatchPoint Send
		sm.Influx.Send(sm.bps)

		time.Sleep(sleep)
	}
	wg.Done()
}
