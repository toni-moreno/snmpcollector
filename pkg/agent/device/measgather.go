package device

import (
	"github.com/toni-moreno/snmpcollector/pkg/data/measurement"
	"sync"
	"time"
)

func (d *SnmpDevice) measConcurrentGatherAndSend() {
	var totalGets int64
	var totalErrors int64

	startSnmpStats := time.Now()
	var wg sync.WaitGroup
	for _, m := range d.Measurements {
		wg.Add(1)
		go func(m *measurement.Measurement) {
			defer wg.Done()
			bpts, _ := d.Influx.BP()
			d.Debugf("-------Processing measurement : %s", m.ID)

			nGets, nErrors, _ := m.GetData()
			totalGets += nGets
			totalErrors += nErrors

			m.ComputeEvaluatedMetrics()
			m.ComputeOidConditionalMetrics()

			if nGets > 0 {
				d.addGets(nGets)
			}
			if nErrors > 0 {
				d.addErrors(nErrors)
			}
			//prepare batchpoint
			points := m.GetInfluxPoint(d.TagMap)
			if bpts != nil {
				(*bpts).AddPoints(points)
				//send data
				d.Influx.Send(bpts)
			} else {
				d.Warnf("Can not send data to the output DB becaouse of batchpoint creation error")
			}
		}(m)
	}
	wg.Wait()
	elapsedSnmpStats := time.Since(startSnmpStats)
	d.Infof("snmp pooling took [%s] SNMP: Gets [%d] Errors [%d]", elapsedSnmpStats, totalGets, totalErrors)
	d.setGatherStats(startSnmpStats, elapsedSnmpStats)
	if d.selfmon != nil {
		fields := map[string]interface{}{
			"process_t": elapsedSnmpStats.Seconds(),
			"getsent":   totalGets,
			"geterror":  totalErrors,
		}
		d.selfmon.AddDeviceMetrics(d.cfg.ID, fields)
	}

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

		if nGets > 0 {
			d.addGets(nGets)
		}
		if nErrors > 0 {
			d.addErrors(nErrors)
		}
		//prepare batchpoint
		points := m.GetInfluxPoint(d.TagMap)
		if bpts != nil {
			(*bpts).AddPoints(points)
		}
	}

	elapsedSnmpStats := time.Since(startSnmpStats)
	d.Infof("snmp pooling took [%s] SNMP: Gets [%d] Errors [%d]", elapsedSnmpStats, totalGets, totalErrors)
	d.setGatherStats(startSnmpStats, elapsedSnmpStats)
	if d.selfmon != nil {
		fields := map[string]interface{}{
			"process_t": elapsedSnmpStats.Seconds(),
			"getsent":   totalGets,
			"geterror":  totalErrors,
		}
		d.selfmon.AddDeviceMetrics(d.cfg.ID, fields)
	}
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
	d.Infof("influx send took [%s]", elapsedInfluxStats)

}
