package device

import (
	"sync"
	"time"

	"github.com/toni-moreno/snmpcollector/pkg/data/measurement"
)

func (d *SnmpDevice) measConcurrentGatherAndSend() {
	startSnmpStats := time.Now()
	var wg sync.WaitGroup
	for _, m := range d.Measurements {
		wg.Add(1)
		go func(m *measurement.Measurement) {
			defer wg.Done()
			bpts, _ := d.OutDB.BP()
			d.Debugf("-------Processing measurement : %s", m.ID)

			nGets, nProcs, nErrs, _ := m.GetData()
			d.stats.UpdateSnmpGetStats(nGets, nProcs, nErrs)

			m.ComputeOidConditionalMetrics()
			m.ComputeEvaluatedMetrics(d.VarMap)

			//prepare batchpoint
			metSent, metError, measSent, measError, points := m.GetOutDBPoint(d.TagMap)
			d.stats.AddMeasStats(metSent, metError, measSent, measError)
			startOutDBStats := time.Now()
			if bpts != nil {
				(*bpts).AddPoints(points)
				//send data
				d.OutDB.Send(bpts)
			} else {
				d.Warnf("Can not send data to the output DB becaouse of batchpoint creation error")
			}
			elapsedOutDBStats := time.Since(startOutDBStats)
			d.stats.AddSentDuration(startOutDBStats, elapsedOutDBStats)

		}(m)
	}
	wg.Wait()
	elapsedSnmpStats := time.Since(startSnmpStats)
	d.stats.SetGatherDuration(startSnmpStats, elapsedSnmpStats)
}

func (d *SnmpDevice) measSeqGatherAndSend() {
	var tnGets int64
	var tnProc int64
	var tnErrors int64
	bpts, _ := d.OutDB.BP()
	startSnmpStats := time.Now()
	for _, m := range d.Measurements {

		d.Debugf("-------Processing measurement : %s", m.ID)

		nGets, nProc, nErrors, _ := m.GetData()
		tnGets += nGets
		tnProc += nProc
		tnErrors += nErrors

		m.ComputeOidConditionalMetrics()
		m.ComputeEvaluatedMetrics(d.VarMap)

		//prepare batchpoint
		metSent, metError, measSent, measError, points := m.GetOutDBPoint(d.TagMap)
		d.stats.AddMeasStats(metSent, metError, measSent, measError)
		if bpts != nil {
			(*bpts).AddPoints(points)
		}
	}

	elapsedSnmpStats := time.Since(startSnmpStats)
	d.stats.UpdateSnmpGetStats(tnGets, tnProc, tnErrors)
	d.stats.SetGatherDuration(startSnmpStats, elapsedSnmpStats)
	/*************************
	 *
	 * Send data to OutDBDB process
	 *
	 ***************************/

	startOutDBStats := time.Now()
	if bpts != nil {
		d.OutDB.Send(bpts)
	} else {
		d.Warnf("Can not send data to the output DB becaouse of batchpoint creation error")
	}
	elapsedOutDBStats := time.Since(startOutDBStats)
	d.stats.AddSentDuration(startOutDBStats, elapsedOutDBStats)

}
