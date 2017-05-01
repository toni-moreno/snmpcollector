package device

import (
	"github.com/toni-moreno/snmpcollector/pkg/data/measurement"
	"sync"
	"time"
)

func (d *SnmpDevice) measConcurrentGatherAndSend() {
	startSnmpStats := time.Now()
	var wg sync.WaitGroup
	for _, m := range d.Measurements {
		wg.Add(1)
		go func(m *measurement.Measurement) {
			defer wg.Done()
			bpts, _ := d.Influx.BP()
			d.Debugf("-------Processing measurement : %s", m.ID)

			nGets, nErrors, _ := m.GetData()

			m.ComputeEvaluatedMetrics()
			m.ComputeOidConditionalMetrics()

			if nGets > 0 {
				d.stats.AddGets(nGets)
			}
			if nErrors > 0 {
				d.stats.AddErrors(nErrors)
			}
			//prepare batchpoint
			points := m.GetInfluxPoint(d.TagMap)
			startInfluxStats := time.Now()
			if bpts != nil {
				(*bpts).AddPoints(points)
				//send data
				d.Influx.Send(bpts)
			} else {
				d.Warnf("Can not send data to the output DB becaouse of batchpoint creation error")
			}
			elapsedInfluxStats := time.Since(startInfluxStats)
			d.stats.AddSentDuration(elapsedInfluxStats)

		}(m)
	}
	wg.Wait()
	elapsedSnmpStats := time.Since(startSnmpStats)
	d.stats.SetGatherDuration(startSnmpStats, elapsedSnmpStats)
}

func (d *SnmpDevice) measSeqGatherAndSend() {
	var totalGets int64
	var totalErrors int64
	bpts, _ := d.Influx.BP()
	startSnmpStats := time.Now()
	for _, m := range d.Measurements {

		d.Debugf("-------Processing measurement : %s", m.ID)

		nGets, nErrors, _ := m.GetData()
		totalGets += nGets
		totalErrors += nErrors

		m.ComputeEvaluatedMetrics()
		m.ComputeOidConditionalMetrics()

		//prepare batchpoint
		points := m.GetInfluxPoint(d.TagMap)
		if bpts != nil {
			(*bpts).AddPoints(points)
		}
	}

	if totalGets > 0 {
		d.stats.AddGets(totalGets)
	}
	if totalErrors > 0 {
		d.stats.AddErrors(totalErrors)
	}
	elapsedSnmpStats := time.Since(startSnmpStats)
	d.stats.SetGatherDuration(startSnmpStats, elapsedSnmpStats)
	/*************************
	 *
	 * Send data to InfluxDB process
	 *
	 ***************************/

	startInfluxStats := time.Now()
	if bpts != nil {
		d.Influx.Send(bpts)
	} else {
		d.Warnf("Can not send data to the output DB becaouse of batchpoint creation error")
	}
	elapsedInfluxStats := time.Since(startInfluxStats)
	d.stats.AddSentDuration(elapsedInfluxStats)

}
