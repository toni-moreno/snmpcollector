package measurement

import (
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	metric "github.com/influxdata/telegraf/metric"
)

// GetInfluxPoint get points from measuremnetsl
func (m *Measurement) GenMetrics(hostTags map[string]string) (int64, int64, int64, int64, []telegraf.Metric) {
	var metSent int64
	var metError int64
	var measSent int64
	var measError int64
	var tmetrics []telegraf.Metric

	switch m.cfg.GetMode {
	case "value":
		k := m.MetricTable.Row["0"]
		var t time.Time
		// copy tags and add index tag
		Tags := make(map[string]string)
		for kT, vT := range hostTags {
			Tags[kT] = vT
		}
		Fields := make(map[string]interface{})
		for _, vMtr := range k.Data {
			me := vMtr.ImportFieldsAndTags(m.cfg.ID, Fields, Tags)
			metError += me
			// check again if metric is valid
			if vMtr.Valid {
				t = vMtr.CurTime
			} else {
				m.Log.Debugf("SKIPPING TS due to invalid %s metric %s", m.cfg.ID, vMtr.CurTime)
			}
		}
		metSent += int64(len(Fields))
		m.Log.Debugf("FIELDS:%+v", Fields)

		tmetric := metric.New(m.cfg.Name, Tags, Fields, t)

		// Validate that at least len of fields > 0
		if len(Fields) == 0 {
			m.Log.Warnf("error, empty fields: [%d]", len(Tags), len(Fields))
			measError++
			break
		}

		// the SNMPCollector can generate emtpy tags in certains configurations (i.e: fill() on multiindex measurements)
		// and it is not allowed on IPL. The retrieve tmetric can return an empty tag value,
		// so we should ensure that this is removed from the final metric
		for tk, tv := range tmetric.Tags() {
			if tv == "" {
				tmetric.RemoveTag(tk)
			}
		}
		measSent++
		k.Valid = true
		tmetrics = append(tmetrics, tmetric)
	case "indexed", "indexed_it", "indexed_mit", "indexed_multiple":
		var t time.Time
		for idx, vIdx := range m.MetricTable.Row {
			m.Log.Debugf("generating influx point for indexed %s", idx)
			// copy tags and add index tag
			Tags := make(map[string]string)
			for kT, vT := range hostTags {
				Tags[kT] = vT
			}
			// Need to check that the lengt of stags is the same as m.tagName
			// The split must be only applied on indexed_multiple measurements
			stags := []string{idx}

			if m.cfg.GetMode == "indexed_multiple" {
				stags = strings.Split(idx, "|")
				if len(stags) != len(m.TagName) {
					m.Log.Errorf("Tags %+v - doesn't match with generated tags %+v. Error in generating point", m.TagName, stags)
					return metSent, metError, measSent, measError, tmetrics
				}
			}
			for k, v := range m.TagName {
				Tags[v] = stags[k]
			}
			m.Log.Debugf("IDX :%+v", vIdx)
			Fields := make(map[string]interface{})
			for _, vMtr := range vIdx.Data {
				me := vMtr.ImportFieldsAndTags(m.cfg.ID, Fields, Tags)
				metError += me
				// check again if metric is valid
				if vMtr.Valid {
					t = vMtr.CurTime
				} else {
					m.Log.Debugf("SKIPPING TS due to invalid %s metric %s", m.cfg.ID, vMtr.CurTime)
				}
			}
			metSent += int64(len(Fields))
			// here we can chek Fields names prior to send data
			m.Log.Debugf("FIELDS:%+v TAGS:%+v", Fields, Tags)
			tmetric := metric.New(m.cfg.Name, Tags, Fields, t)

			// Validate that at least len of fields > 0
			if len(Fields) == 0 {
				m.Log.Warnf("error, empty fields: [%d]", len(Tags), len(Fields))
				measError++
				continue
			}
			// the SNMPCollector can generate emtpy tags in certains configurations (i.e: fill() on multiindex measurements)
			// and it is not allowed on IPL. The retrieve tmetric can return an empty tag value,
			// so we should ensure that this is removed from the final metric
			for tk, tv := range tmetric.Tags() {
				if tv == "" {
					tmetric.RemoveTag(tk)
				}
			}
			measSent++
			vIdx.Valid = true
			tmetrics = append(tmetrics, tmetric)
		}
	}
	return metSent, metError, measSent, measError, tmetrics
}
