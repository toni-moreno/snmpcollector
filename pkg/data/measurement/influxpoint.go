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
		k := m.MetricTable["0"]
		var t time.Time
		Fields := make(map[string]interface{})
		for _, v_mtr := range k {
			if v_mtr.CookedValue == nil {
				m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", v_mtr.ID, m.cfg.ID, hostTags, v_mtr)
				metError++
				continue
			}
			if v_mtr.Valid == false {
				m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", v_mtr.ID, m.cfg.ID, hostTags, v_mtr)
				continue
			}
			if v_mtr.Report == metric.NeverReport {
				m.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.ID, m.cfg.ID)
				continue
			}
			if v_mtr.Report == metric.OnNonZeroReport {
				if v_mtr.CookedValue == 0.0 {
					m.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.ID, m.cfg.ID)
					continue
				}
			}
			m.Debugf("generating field for %s value %f ", v_mtr.GetFieldName(), v_mtr.CookedValue)
			m.Debugf("DEBUG METRIC %+v", v_mtr)
			Fields[v_mtr.GetFieldName()] = v_mtr.CookedValue
			t = v_mtr.CurTime
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
		}

	case "indexed", "indexed_it":
		var t time.Time
		for k_idx, v_idx := range m.MetricTable {
			m.Debugf("generating influx point for indexed %s", k_idx)
			//copy tags and add index tag
			Tags := make(map[string]string)
			for k_t, v_t := range hostTags {
				Tags[k_t] = v_t
			}
			Tags[m.cfg.IndexTag] = k_idx
			m.Debugf("IDX :%+v", v_idx)
			Fields := make(map[string]interface{})
			for _, v_mtr := range v_idx {
				v_mtr.PrintDebugCfg()
				if v_mtr.IsTag() == true {
					if v_mtr.CookedValue == nil {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", v_mtr.ID, m.cfg.ID, Tags, v_mtr)
						metError++ //not sure if an tag error should be count as metric
						continue
					}
					if v_mtr.Valid == false {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", v_mtr.ID, m.cfg.ID, hostTags, v_mtr)
						continue
					}
					if v_mtr.Report == metric.NeverReport {
						m.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.ID, m.cfg.ID)
						continue
					}

					var tag string
					switch v := v_mtr.CookedValue.(type) {
					case float64:
						//most of times these will be integers
						tag = strconv.FormatInt(int64(v), 10)
					default:
						//assume string
						tag = v.(string)
					}
					if v_mtr.Report == metric.OnNonZeroReport {
						if tag == "0" {
							m.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.ID, m.cfg.ID)
							continue
						}
					}
					m.Debugf("generating Tag for Metric: %s : tagname: %s", v_mtr.GetFieldName(), tag)
					Tags[v_mtr.GetFieldName()] = tag
				} else {
					if v_mtr.CookedValue == nil {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", v_mtr.ID, m.cfg.ID, Tags, v_mtr)
						metError++
						continue
					}
					if v_mtr.Valid == false {
						m.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has obsolete data => See Metric Runtime [ %+v ]", v_mtr.ID, m.cfg.ID, hostTags, v_mtr)
						continue
					}
					if v_mtr.Report == metric.NeverReport {
						m.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.ID, m.cfg.ID)
						continue
					}
					if v_mtr.Report == metric.OnNonZeroReport {
						if v_mtr.CookedValue == 0.0 {
							m.Debugf("REPORT on non zero in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.ID, m.cfg.ID)
							continue
						}
					}
					m.Debugf("generating field for Metric: %s : value %#v", v_mtr.GetFieldName(), v_mtr.CookedValue)
					m.Debugf("DEBUG METRIC %+v", v_mtr)
					Fields[v_mtr.GetFieldName()] = v_mtr.CookedValue
				}
				metSent++
				//reported gathered time for the measurment is choosed as the last field time
				t = v_mtr.CurTime
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
				m.Debugf("GENERATED INFLUX POINT[%s] index [%s]: %+v", m.cfg.Name, k_idx, pt)
				ptarray = append(ptarray, pt)
				measSent++
			}
		}

	}

	return metSent, metError, measSent, measError, ptarray

}
