package metric

import (
	"fmt"

	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

// MetricRow Measurment row type
type MetricRow struct {
	Valid bool
	Data  map[string]*SnmpMetric
}

// Invalidate set invalid all metrics on the row
func (mr *MetricRow) Invalidate() {
	mr.Valid = false
	for _, m := range mr.Data {
		m.Valid = false
	}
}

// NewMetricRow create a new metric Row
func NewMetricRow() *MetricRow {
	mr := MetricRow{Valid: true}
	mr.Data = make(map[string]*SnmpMetric)
	return &mr
}

// Add Add a new metric in the row
func (mr *MetricRow) Add(id string, m *SnmpMetric) {
	mr.Data[id] = m
}

// SetVisible set visible only the selected metrics
func (mr *MetricRow) SetVisible(ar map[string]int) {
	for k1, v := range mr.Data {
		for k2, vis := range ar {
			if k1 == k2 {
				v.Report = vis
				break
			}
		}
	}
}

// MetricTable Sequence of metric rows with headers
type MetricTable struct {
	Header  map[string]interface{}
	visible map[string]int
	log     utils.Logger
	cfg     *config.MeasurementCfg
	Row     map[string]*MetricRow
}

// Log For MetricTable OBject.

// Debugf info
func (mt *MetricTable) Debugf(expr string, vars ...interface{}) {
	expr2 := "METRICTABLE for MEASUREMENT [" + mt.cfg.ID + "]  " + expr
	mt.log.Debugf(expr2, vars...)
}

// Debug info
func (mt *MetricTable) Debug(expr string) {
	expr2 := "METRICTABLE for MEASUREMENT [" + mt.cfg.ID + "] " + expr
	mt.log.Debug(expr2)
}

// Infof info
func (mt *MetricTable) Infof(expr string, vars ...interface{}) {
	expr2 := "METRICTABLE for MEASUREMENT [" + mt.cfg.ID + "]  " + expr
	mt.log.Infof(expr2, vars...)
}

// Errorf info
func (mt *MetricTable) Errorf(expr string, vars ...interface{}) {
	expr2 := "METRICTABLE for MEASUREMENT [" + mt.cfg.ID + "]  " + expr
	mt.log.Errorf(expr2, vars...)
}

// Warnf log warn info
func (mt *MetricTable) Warnf(expr string, vars ...interface{}) {
	expr2 := "METRICTABLE  for MEASUREMENT [" + mt.cfg.ID + "]" + expr
	mt.log.Warnf(expr2, vars...)
}

// AddRow add a new row to the metricTable
func (mt *MetricTable) AddRow(id string, mr *MetricRow) {
	mt.Row[id] = mr
}

// Len get number of rows for the MetricTable
func (mt *MetricTable) Len() int {
	return len(mt.Row)
}

// InvalidateTable set invalid each row in the table
func (mt *MetricTable) InvalidateTable() {
	for _, r := range mt.Row {
		r.Invalidate()
	}
}

// GetSnmpMaps get an  OID array  and a metric Object OID mapped
func (mt *MetricTable) GetSnmpMaps() ([]string, map[string]*SnmpMetric) {
	snmpOids := []string{}
	OidSnmpMap := make(map[string]*SnmpMetric)
	for idx, row := range mt.Row {
		mt.Debugf("KEY iDX %s", idx)
		// index level
		for kM, vM := range row.Data {
			mt.Debugf("KEY METRIC %s OID %s", kM, vM.RealOID)
			t := vM.GetDataSrcType()
			switch t {
			case "STRINGEVAL":
			default:
				// this array is used in SnmpGetData to send IOD's to the end device
				// so it can not contain any other thing than OID's
				// on string eval it contains a identifier not OID
				snmpOids = append(snmpOids, vM.RealOID)
			}
			OidSnmpMap[vM.RealOID] = vM
		}
	}
	return snmpOids, OidSnmpMap
}

// GetSnmpMap get and snmpmetric OID map
func (mt *MetricTable) GetSnmpMap() map[string]*SnmpMetric {
	OidSnmpMap := make(map[string]*SnmpMetric)
	for idx, row := range mt.Row {
		mt.Debugf("KEY iDX %s", idx)
		// index level
		for kM, vM := range row.Data {
			mt.Debugf("KEY METRIC %s OID %s", kM, vM.RealOID)
			OidSnmpMap[vM.RealOID] = vM
		}
	}
	return OidSnmpMap
}

// NewMetricTable create a new MetricTable
func NewMetricTable(c *config.MeasurementCfg, l utils.Logger, CurIndexedLabels map[string]string) *MetricTable {
	mt := MetricTable{}
	mt.Init(c, l, CurIndexedLabels)
	return &mt
}

// Init Initialize the MetricTable Object
func (mt *MetricTable) Init(c *config.MeasurementCfg, l utils.Logger, CurIndexedLabels map[string]string) {
	mt.cfg = c
	mt.log = l
	mt.Row = make(map[string]*MetricRow)
	mt.visible = make(map[string]int, len(mt.cfg.Fields))
	mt.Header = make(map[string]interface{}, len(mt.cfg.Fields))
	for _, r := range mt.cfg.Fields {
		mt.visible[r.ID] = r.Report
		for _, val := range mt.cfg.FieldMetric {
			if r.ID == val.ID {
				mt.Header[val.FieldName] = val.GetMetricHeader(r.Report)
				continue
			}
		}
		for _, val := range mt.cfg.EvalMetric {
			if r.ID == val.ID {
				mt.Header[val.FieldName] = val.GetMetricHeader(r.Report)
				continue
			}
		}
		for _, val := range mt.cfg.OidCondMetric {
			if r.ID == val.ID {
				mt.Header[val.FieldName] = val.GetMetricHeader(r.Report)
				continue
			}
		}
	}

	// create metrics.
	switch mt.cfg.GetMode {
	case "value":
		// for each field
		idx := NewMetricRow()
		for k, smcfg := range mt.cfg.FieldMetric {
			mt.Debugf("initializing [value]metric cfgi %s", smcfg.ID)
			metr, err := New(smcfg, mt.log)
			if err != nil {
				mt.Errorf("ERROR on create new [value] field metric %d : Error: %s ", k, err)
				continue
			}
			metr.SetLogger(mt.log)
			idx.Add(smcfg.ID, metr)
		}
		for k, smcfg := range mt.cfg.EvalMetric {
			mt.Debugf("initializing [value] [evaluated] metric cfg %s", smcfg.ID)
			metr, err := New(smcfg, mt.log)
			if err != nil {
				mt.Errorf("ERROR on create new [value] [evaluated] field metric %d : Error: %s ", k, err)
				continue
			}
			metr.SetLogger(mt.log)
			metr.RealOID = mt.cfg.ID + "." + smcfg.ID
			idx.Add(smcfg.ID, metr)
		}
		for k, smcfg := range mt.cfg.OidCondMetric {
			mt.Debugf("initializing [value] [oid condition evaluated] metric cfg %s", smcfg.ID)
			metr, err := New(smcfg, mt.log)
			if err != nil {
				mt.Errorf("ERROR on create new [value] [oid condition evaluated] field metric %d : Error: %s ", k, err)
				continue
			}
			metr.RealOID = mt.cfg.ID + "." + smcfg.ID
			idx.Add(smcfg.ID, metr)
		}
		// setup visibility on db for each metric

		idx.SetVisible(mt.visible)
		mt.AddRow("0", idx)

	case "indexed", "indexed_it", "indexed_mit", "indexed_multiple":
		// for each field an each index (previously initialized)
		for key, label := range CurIndexedLabels {
			idx := NewMetricRow()
			mt.Debugf("initializing [indexed] metric cfg for [%s/%s]", key, label)
			for k, smcfg := range mt.cfg.FieldMetric {
				metr, err := New(smcfg, mt.log)
				if err != nil {
					mt.Errorf("ERROR on create new [indexed] fields metric  %d: Error: %s ", k, err)
					continue
				}
				metr.SetLogger(mt.log)
				metr.RealOID += "." + key
				idx.Add(smcfg.ID, metr)
			}
			for k, smcfg := range mt.cfg.EvalMetric {
				metr, err := New(smcfg, mt.log)
				if err != nil {
					mt.Errorf("ERROR on create new [indexed] [evaluated] fields metric  %d: Error: %s ", k, err)
					continue
				}
				metr.SetLogger(mt.log)
				metr.RealOID = mt.cfg.ID + "." + smcfg.ID + "." + key // unique identificator for this metric
				idx.Add(smcfg.ID, metr)
			}
			// setup visibility on db for each metric
			idx.SetVisible(mt.visible)
			mt.AddRow(label, idx)
		}

	default:
		mt.Errorf("Unknown Measurement GetMode Config :%s", mt.cfg.GetMode)
	}
}

// Pop remove MetricRows from the MetricTable
func (mt *MetricTable) Pop(p map[string]string) error {
	if mt.cfg.GetMode == "value" {
		return fmt.Errorf("Can not pop values in a measurement type value : %s", mt.cfg.ID)
	}
	for key, label := range p {
		mt.Infof("removing [indexed] metric cfg for [%s/%s]", key, label)
		delete(mt.Row, label)
	}
	return nil
}

// Push add New MetricRows to the MetricTable
func (mt *MetricTable) Push(p map[string]string) error {
	if mt.cfg.GetMode == "value" {
		return fmt.Errorf("Can not push new values in a measurement type value : %s", mt.cfg.ID)
	}
	for key, label := range p {
		idx := NewMetricRow()
		mt.Infof("initializing [indexed] metric cfg for [%s/%s]", key, label)
		for k, smcfg := range mt.cfg.FieldMetric {
			metr, err := New(smcfg, mt.log)
			if err != nil {
				mt.Errorf("ERROR on create new [indexed] fields metric  %d: Error: %s ", k, err)
				continue
			}
			metr.SetLogger(mt.log)
			metr.RealOID += "." + key
			idx.Add(smcfg.ID, metr)
		}
		for k, smcfg := range mt.cfg.EvalMetric {
			metr, err := New(smcfg, mt.log)
			if err != nil {
				mt.Errorf("ERROR on create new [indexed] [evaluated] fields metric  %d: Error: %s ", k, err)
				continue
			}
			metr.SetLogger(mt.log)
			metr.RealOID = mt.cfg.ID + "." + smcfg.ID + "." + key // unique identificator for this metric
			idx.Add(smcfg.ID, metr)
		}
		idx.SetVisible(mt.visible)
		mt.AddRow(label, idx)
	}
	return nil
}
