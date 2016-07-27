package main

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"os"
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
	bps                 *client.BatchPoints
}

func (sm *SelfMonConfig) Init() {
	//Init extra tags
	if len(sm.ExtraTags) > 0 {
		sm.TagMap = make(map[string]string)
		for _, tag := range sm.ExtraTags {
			s := strings.Split(tag, "=")
			key, value := s[0], s[1]
			sm.TagMap[key] = value
		}
	}
}

func (sm *SelfMonConfig) Report(metric string, value float64, timestamp time.Time) error {
	pt, _ := client.NewPoint(
		metric,
		sm.TagMap,
		map[string]interface{}{
			"value": value,
		},
		timestamp,
	)

	(*sm.bps).AddPoint(pt)
	return nil
}

func (sm *SelfMonConfig) ReportStats(wg *sync.WaitGroup) {
	if sm.runtimeStatsRunning {
		fmt.Fprintf(os.Stderr, "Runtime stats is already running\n")
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

		sm.Report(fmt.Sprintf("%s.goroutines", prefix), float64(runtime.NumGoroutine()), now)
		sm.Report(fmt.Sprintf("%s.memory.allocated", prefix), float64(memStats.Alloc), now)
		sm.Report(fmt.Sprintf("%s.memory.mallocs", prefix), float64(memStats.Mallocs), now)
		sm.Report(fmt.Sprintf("%s.memory.frees", prefix), float64(memStats.Frees), now)
		sm.Report(fmt.Sprintf("%s.memory.gc.total_pause", prefix), float64(memStats.PauseTotalNs)/nsInMs, now)
		sm.Report(fmt.Sprintf("%s.memory.heap", prefix), float64(memStats.HeapAlloc), now)
		sm.Report(fmt.Sprintf("%s.memory.stack", prefix), float64(memStats.StackInuse), now)

		if lastPauseNs > 0 {
			pauseSinceLastSample := memStats.PauseTotalNs - lastPauseNs
			sm.Report(fmt.Sprintf("%s.memory.gc.pause_per_second", prefix), float64(pauseSinceLastSample)/nsInMs/sleep.Seconds(), now)
			sm.Report(fmt.Sprintf("%s.memory.gc.pause_per_interval", prefix), float64(pauseSinceLastSample)/nsInMs, now)
		}
		lastPauseNs = memStats.PauseTotalNs

		countGc := int(memStats.NumGC - lastNumGc)
		if lastNumGc > 0 {
			diff := float64(countGc)
			diffTime := now.Sub(lastSampleTime).Seconds()
			sm.Report(fmt.Sprintf("%s.memory.gc.gc_per_second", prefix), diff/diffTime, now)
			sm.Report(fmt.Sprintf("%s.memory.gc.gc_per_interval", prefix), diff, now)

		}

		if countGc > 0 {
			if countGc > 256 {
				fmt.Fprintf(os.Stderr, "We're missing some gc pause times")
				countGc = 256
			}
			var totalPause float64 = 0
			for i := 0; i < countGc; i++ {
				idx := int((memStats.NumGC-uint32(i))+255) % 256
				pause := float64(memStats.PauseNs[idx])
				totalPause += pause
				//	sm.Report(fmt.Sprintf("%s.memory.gc.pause", prefix), pause/nsInMs, now)
			}
			sm.Report(fmt.Sprintf("%s.memory.gc.pause_per_interval", prefix), totalPause/nsInMs, now)
		}

		lastNumGc = memStats.NumGC
		lastSampleTime = now

		//BatchPoint Send
		sm.Influx.Send(sm.bps)

		time.Sleep(sleep)
	}
	wg.Done()
}
