package measurement

import (
	"strconv"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/toni-moreno/snmpcollector/pkg/data/metric"
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
		Fields := make(map[string]interface{})
		for _, vMtr := range k.Data {
			if vMtr.CookedValue == nil {
				m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", vMtr.ID, m.cfg.ID, hostTags, vMtr)
				metError++
				continue
			}
			if vMtr.Valid == false {
				m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", vMtr.ID, m.cfg.ID, hostTags, vMtr)
				continue
			}
			if vMtr.Report == metric.NeverReport {
				m.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", vMtr.ID, m.cfg.ID)
				continue
			}
			if vMtr.Report == metric.OnNonZeroReport {
				if vMtr.CookedValue == 0.0 {
					m.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", vMtr.ID, m.cfg.ID)
					continue
				}
			}
			m.Debugf("generating field for %s value %f ", vMtr.GetFieldName(), vMtr.CookedValue)
			m.Debugf("DEBUG METRIC %+v", vMtr)
			Fields[vMtr.GetFieldName()] = vMtr.CookedValue
			t = vMtr.CurTime
			metSent++
		}
		m.Debugf("FIELDS:%+v", Fields)

		pt, err := client.NewPoint(
			m.cfg.Name,
			hostTags,
			Fields,
			t,
		)
		if err != nil {
			m.Warnf("error in influx point building:%s", err)
			measError++
		} else {
			m.Debugf("GENERATED INFLUX POINT[%s] value: %+v", m.cfg.Name, pt)
			ptarray = append(ptarray, pt)
			measSent++
			k.Valid = true
		}

	case "indexed", "indexed_it":
		var t time.Time
		for idx, vIdx := range m.MetricTable.Row {
			m.Debugf("generating influx point for indexed %s", idx)
			//copy tags and add index tag
			Tags := make(map[string]string)
			for kT, vT := range hostTags {
				Tags[kT] = vT
			}
			Tags[m.cfg.IndexTag] = idx
			m.Debugf("IDX :%+v", vIdx)
			Fields := make(map[string]interface{})
			for _, vMtr := range vIdx.Data {
				vMtr.PrintDebugCfg()
				if vMtr.IsTag() == true {
					if vMtr.CookedValue == nil {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", vMtr.ID, m.cfg.ID, Tags, vMtr)
						metError++ //not sure if an tag error should be count as metric
						continue
					}
					if vMtr.Valid == false {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", vMtr.ID, m.cfg.ID, hostTags, vMtr)
						continue
					}
					if vMtr.Report == metric.NeverReport {
						m.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", vMtr.ID, m.cfg.ID)
						continue
					}

					var tag string
					switch v := vMtr.CookedValue.(type) {
					case float64:
						//most of times these will be integers
						tag = strconv.FormatInt(int64(v), 10)
					default:
						//assume string
						tag = v.(string)
					}
					if vMtr.Report == metric.OnNonZeroReport {
						if tag == "0" {
							m.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", vMtr.ID, m.cfg.ID)
							continue
						}
					}
					m.Debugf("generating Tag for Metric: %s : tagname: %s", vMtr.GetFieldName(), tag)
					Tags[vMtr.GetFieldName()] = tag
				} else {
					if vMtr.CookedValue == nil {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", vMtr.ID, m.cfg.ID, Tags, vMtr)
						metError++
						continue
					}
					if vMtr.Valid == false {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", vMtr.ID, m.cfg.ID, hostTags, vMtr)
						continue
					}
					if vMtr.Report == metric.NeverReport {
						m.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", vMtr.ID, m.cfg.ID)
						continue
					}
					if vMtr.Report == metric.OnNonZeroReport {
						if vMtr.CookedValue == 0.0 {
							m.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", vMtr.ID, m.cfg.ID)
							continue
						}
					}
					m.Debugf("generating field for Metric: %s : value %#v", vMtr.GetFieldName(), vMtr.CookedValue)
					m.Debugf("DEBUG METRIC %+v", vMtr)
					Fields[vMtr.GetFieldName()] = vMtr.CookedValue
				}
				metSent++
				//reported gathered time for the measurment is choosed as the last field time
				t = vMtr.CurTime
			}
			//here we can chek Fields names prior to send data
			m.Debugf("FIELDS:%+v TAGS:%+v", Fields, Tags)
			pt, err := client.NewPoint(
				m.cfg.Name,
				Tags,
				Fields,
				t,
			)
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
