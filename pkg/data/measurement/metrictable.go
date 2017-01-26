package measurement

import (
	"fmt"
	"github.com/toni-moreno/snmpcollector/pkg/data/metric"
)

// PushMetricTable add map
func (m *Measurement) PushMetricTable(p map[string]string) error {
	if m.cfg.GetMode == "value" {
		return fmt.Errorf("Can not push new values in a measurement type value : %s", m.cfg.ID)
	}
	for key, label := range p {
		idx := make(map[string]*metric.SnmpMetric)
		m.log.Infof("initializing [indexed] metric cfg for [%s/%s]", key, label)
		for k, smcfg := range m.cfg.FieldMetric {
			metr, err := metric.New(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [indexed] fields metric  %d: Error: %s ", k, err)
				continue
			}
			metr.SetLogger(m.log)
			metr.RealOID += "." + key
			idx[smcfg.ID] = metr
		}
		for k, smcfg := range m.cfg.EvalMetric {
			metr, err := metric.New(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [indexed] [evaluated] fields metric  %d: Error: %s ", k, err)
				continue
			}
			metr.SetLogger(m.log)
			metr.RealOID = m.cfg.ID + "." + smcfg.ID + "." + key //unique identificator for this metric
			idx[smcfg.ID] = metr
		}
		//setup visibility on db for each metric
		for k, v := range idx {
			report := metric.AlwaysReport
			for _, r := range m.cfg.Fields {
				if r.ID == k {
					report = r.Report
					break
				}
			}
			v.Report = report
		}
		m.MetricTable[label] = idx
	}
	return nil
}

// PopMetricTable add
func (m *Measurement) PopMetricTable(p map[string]string) error {
	if m.cfg.GetMode == "value" {
		return fmt.Errorf("Can not pop values in a measurement type value : %s", m.cfg.ID)
	}
	for key, label := range p {
		m.log.Infof("removing [indexed] metric cfg for [%s/%s]", key, label)
		delete(m.MetricTable, label)
	}
	return nil
}

/* InitMetricTable
 */
func (m *Measurement) InitMetricTable() {
	m.MetricTable = make(map[string]map[string]*metric.SnmpMetric)

	//create metrics.
	switch m.cfg.GetMode {
	case "value":
		//for each field
		idx := make(map[string]*metric.SnmpMetric)
		for k, smcfg := range m.cfg.FieldMetric {
			m.log.Debugf("initializing [value]metric cfgi %s", smcfg.ID)
			metr, err := metric.New(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [value] field metric %d : Error: %s ", k, err)
				continue
			}
			metr.SetLogger(m.log)
			idx[smcfg.ID] = metr
		}
		for k, smcfg := range m.cfg.EvalMetric {
			m.log.Debugf("initializing [value] [evaluated] metric cfg %s", smcfg.ID)
			metr, err := metric.New(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [value] [evaluated] field metric %d : Error: %s ", k, err)
				continue
			}
			metr.SetLogger(m.log)
			metr.RealOID = m.cfg.ID + "." + smcfg.ID
			idx[smcfg.ID] = metr
		}
		for k, smcfg := range m.cfg.OidCondMetric {
			m.log.Debugf("initializing [value] [oid condition evaluated] metric cfg %s", smcfg.ID)
			metr, err := metric.NewWithLog(smcfg, m.log)
			if err != nil {
				m.log.Errorf("ERROR on create new [value] [oid condition evaluated] field metric %d : Error: %s ", k, err)
				continue
			}
			metr.RealOID = m.cfg.ID + "." + smcfg.ID
			idx[smcfg.ID] = metr
		}
		//setup visibility on db for each metric
		for k, v := range idx {
			report := metric.AlwaysReport
			for _, r := range m.cfg.Fields {
				if r.ID == k {
					report = r.Report
					break
				}
			}
			v.Report = report
		}
		m.MetricTable["0"] = idx

	case "indexed", "indexed_it":
		//for each field an each index (previously initialized)
		for key, label := range m.CurIndexedLabels {
			idx := make(map[string]*metric.SnmpMetric)
			m.log.Debugf("initializing [indexed] metric cfg for [%s/%s]", key, label)
			for k, smcfg := range m.cfg.FieldMetric {
				metr, err := metric.New(smcfg)
				if err != nil {
					m.log.Errorf("ERROR on create new [indexed] fields metric  %d: Error: %s ", k, err)
					continue
				}
				metr.SetLogger(m.log)
				metr.RealOID += "." + key
				idx[smcfg.ID] = metr
			}
			for k, smcfg := range m.cfg.EvalMetric {
				metr, err := metric.New(smcfg)
				if err != nil {
					m.log.Errorf("ERROR on create new [indexed] [evaluated] fields metric  %d: Error: %s ", k, err)
					continue
				}
				metr.SetLogger(m.log)
				metr.RealOID = m.cfg.ID + "." + smcfg.ID + "." + key //unique identificator for this metric
				idx[smcfg.ID] = metr
			}
			//setup visibility on db for each metric
			for k, v := range idx {
				report := metric.AlwaysReport
				for _, r := range m.cfg.Fields {
					if r.ID == k {
						report = r.Report
						break
					}
				}
				v.Report = report
			}
			m.MetricTable[label] = idx
		}

	default:
		m.log.Errorf("Unknown Measurement GetMode Config :%s", m.cfg.GetMode)
	}
}
