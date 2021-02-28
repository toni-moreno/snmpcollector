package measurement

import (
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/v2"
)

//GetInfluxPoint get points from measuremnetsl
func (m *Measurement) GetInfluxPoint(hostTags map[string]string) (int64, int64, int64, int64, []*client.Point) {
	var metSent int64
	var metError int64
	var measSent int64
	var measError int64
	var ptarray []*client.Point

	switch m.cfg.GetMode {
	case "value":
		k := m.MetricTable.Row["0"]
		var t time.Time
		//copy tags and add index tag
		Tags := make(map[string]string)
		for kT, vT := range hostTags {
			Tags[kT] = vT
		}
		Fields := make(map[string]interface{})
		for _, vMtr := range k.Data {
			me := vMtr.ImportFieldsAndTags(m.cfg.ID, Fields, Tags)
			metError += me
			//check again if metric is valid
			if vMtr.Valid == true {
				t = vMtr.CurTime
			} else {
				m.Debugf("SKIPPING TS due to invalid %s metric %s", m.cfg.ID, vMtr.CurTime)
			}
		}
		metSent += int64(len(Fields))
		m.Debugf("FIELDS:%+v", Fields)

		pt, err := client.NewPoint(m.cfg.Name, Tags, Fields, t)
		if err != nil {
			m.Warnf("error in influx point building:%s", err)
			measError++
		} else {
			m.Debugf("GENERATED INFLUX POINT[%s] value: %+v", m.cfg.Name, pt)
			ptarray = append(ptarray, pt)
			measSent++
			k.Valid = true
		}

	case "indexed", "indexed_it", "indexed_mit", "indexed_multiple":
		var t time.Time
		for idx, vIdx := range m.MetricTable.Row {
			m.Debugf("generating influx point for indexed %s", idx)
			//copy tags and add index tag
			Tags := make(map[string]string)
			for kT, vT := range hostTags {
				Tags[kT] = vT
			}
			//Need to check that the lengt of stags is the same as m.tagName
			// The split must be only applied on indexed_multiple measurements
			stags := []string{idx}

			if m.cfg.GetMode == "indexed_multiple" {
				stags = strings.Split(idx, "|")
				if len(stags) != len(m.TagName) {
					m.Errorf("Tags %+v - doesn't match with generated tags %+v. Error in generating point", m.TagName, stags)
					return metSent, metError, measSent, measError, ptarray
				}
			}
			for k, v := range m.TagName {
				Tags[v] = stags[k]
			}
			m.Debugf("IDX :%+v", vIdx)
			Fields := make(map[string]interface{})
			for _, vMtr := range vIdx.Data {
				me := vMtr.ImportFieldsAndTags(m.cfg.ID, Fields, Tags)
				metError += me
				//check again if metric is valid
				if vMtr.Valid == true {
					t = vMtr.CurTime
				} else {
					m.Debugf("SKIPPING TS due to invalid %s metric %s", m.cfg.ID, vMtr.CurTime)
				}
			}
			metSent += int64(len(Fields))
			//here we can chek Fields names prior to send data
			m.Debugf("FIELDS:%+v TAGS:%+v", Fields, Tags)
			pt, err := client.NewPoint(m.cfg.Name, Tags, Fields, t)
			if err != nil {
				m.Warnf("error in influx point creation :%s", err)
				measError++
			} else {
				m.Debugf("GENERATED INFLUX POINT[%s] index [%s]: %+v", m.cfg.Name, idx, pt)
				ptarray = append(ptarray, pt)
				measSent++
				vIdx.Valid = true
			}
		}

	}

	return metSent, metError, measSent, measError, ptarray

}
