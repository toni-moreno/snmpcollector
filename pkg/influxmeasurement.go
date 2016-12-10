package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/soniah/gosnmp"
)

/*
type InfluxMeasurementCfg struct {
	id          string   //name of the key in the config array
	Name        string   `toml:"name"`
	Fields      []string `toml:"fields"`
	GetMode     string   `toml:"getmode"` //0=value 1=indexed
	IndexOID    string   `toml:"indexoid"`
	IndexTag    string   `toml:"indextag"`
	fieldMetric []*SnmpMetricCfg
}*/

//Init initialize the measurement configuration
func (mc *InfluxMeasurementCfg) Init(name string, MetricCfg *map[string]*SnmpMetricCfg) error {
	mc.ID = name
	//validate config values
	if len(mc.Name) == 0 {
		return errors.New("Name not set in measurement Config " + mc.ID)
	}
	if len(mc.Fields) == 0 {
		return errors.New("No Fields added to measurement " + mc.ID)
	}

	switch mc.GetMode {
	case "indexed":
		if len(mc.IndexOID) == 0 {
			return errors.New("Indexed measurement with no IndexOID in measurement Config " + mc.ID)
		}
		if len(mc.IndexTag) == 0 {
			return errors.New("Indexed measurement with no IndexTag configuredin measurement " + mc.ID)
		}
		if !strings.HasPrefix(mc.IndexOID, ".") {
			return errors.New("Bad BaseOid format:" + mc.IndexOID + " in metric Config " + mc.ID)
		}

	case "value":
	default:
		return errors.New("Unknown GetMode" + mc.GetMode + " in measurement Config " + mc.ID)
	}

	log.Infof("processing measurement key: %s ", name)
	log.Debugf("%+v", mc)
	for _, f_val := range mc.Fields {
		log.Debugf("looking for measurement %s : fields: %s : Report %t", mc.Name, f_val.ID, f_val.Report)
		if val, ok := (*MetricCfg)[f_val.ID]; ok {
			if val.DataSrcType == "STRINGEVAL" {
				mc.evalMetric = append(mc.evalMetric, val)
				log.Debugf("EVAL metric found measurement %s : fields: %s ", mc.Name, f_val.ID)
			} else {
				mc.fieldMetric = append(mc.fieldMetric, val)
			}
		} else {
			log.Warnf("measurement field  %s NOT FOUND in Metrics Database !", f_val.ID)
		}
	}
	//check if fieldMetric
	if len(mc.fieldMetric) == 0 {
		var s string
		for _, v := range mc.Fields {
			s += ";" + v.ID
		}
		return errors.New("No metrics found with names" + s + " in measurement Config " + mc.ID)
	}
	return nil
}

//InfluxMeasurement the runtime measurement config
type InfluxMeasurement struct {
	cfg              *InfluxMeasurementCfg
	ID               string
	MetricTable      map[string]map[string]*SnmpMetric //snmpMetric mapped with metric_names and Index
	snmpOids         []string
	OidSnmpMap       map[string]*SnmpMetric //snmpMetric mapped with real OID's
	Filterlabels     map[string]string      // `json:"-"`
	AllIndexedLabels map[string]string      //`json:"-"` //all available values on the remote device
	CurIndexedLabels map[string]string      //`json:"-"`
	idxPosInOID      int
	Filter           *MeasFilterCfg
	log              *logrus.Logger
	snmpClient       *gosnmp.GoSNMP
	DisableBulk      bool
	GetData          func() (int64, int64, error)        `json:"-"`
	Walk             func(string, gosnmp.WalkFunc) error `json:"-"`
}

//NewInfluxMeasurement creates object with config , log + goSnmp client
func NewInfluxMeasurement(c *InfluxMeasurementCfg, l *logrus.Logger, cli *gosnmp.GoSNMP, db bool) (*InfluxMeasurement, error) {
	m := &InfluxMeasurement{ID: c.ID, cfg: c, log: l, snmpClient: cli, DisableBulk: db}
	err := m.Init()
	return m, err
}

/*Init does:
 *inicialize AllIndexesLabels
 *Assign CurIndexedLabels to all Labels (until filters set)
 *init MetricTable
 */
func (m *InfluxMeasurement) Init() error {

	var err error
	//Init snmp methods
	if m.cfg.GetMode == "value" {
		m.GetData = m.SnmpGetData
	} else {
		m.GetData = m.SnmpWalkData
		switch {
		case m.snmpClient.Version == gosnmp.Version1 || m.DisableBulk:
			m.Walk = m.snmpClient.Walk
		default:
			m.Walk = m.snmpClient.BulkWalk
		}
	}
	//

	//loading all posible values in 	m.AllIndexedLabels
	if m.cfg.GetMode == "indexed" {
		m.idxPosInOID = len(m.cfg.IndexOID)
		m.log.Infof("Loading Indexed values in : %s", m.cfg.ID)
		m.AllIndexedLabels, err = m.loadIndexedLabels()
		if err != nil {
			m.log.Errorf("Error while trying to load Indexed Labels on for measurement %s for baseOid %s : ERROR: %s", m.cfg.ID, m.cfg.IndexOID, err)
			return err
		}
		//Final Selected Indexes are All Indexed
		m.CurIndexedLabels = m.AllIndexedLabels
	}

	/********************************
	 * Initialize Metric Runtime data in one array m-values
	 * ******************************/
	m.log.Debug("Initialize OID measurement per label => map of metric object per field | OID array [ready to send to the walk device] | OID=>Metric MAP")
	m.InitMetricTable()
	return nil
}

/*func (m *InfluxMeasurement) SetDisableBulk(disable bool) {
	m.DisableBulk = disable
	m.log.Debugf("Disable Snmp Bulk Queries to measurment %s: %t", m.cfg.ID, disable)
}*/

func (m *InfluxMeasurement) PushMetricTable(p map[string]string) error {
	if m.cfg.GetMode == "value" {
		return fmt.Errorf("Can not push new values in a measurement type value : %s", m.cfg.ID)
	}
	for key, label := range p {
		idx := make(map[string]*SnmpMetric)
		m.log.Infof("initializing [indexed] metric cfg for [%s/%s]", key, label)
		for k, smcfg := range m.cfg.fieldMetric {
			metric, err := NewSnmpMetric(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [indexed] fields metric  %d: Error: %s ", k, err)
				continue
			}
			metric.SetLogger(m.log)
			metric.RealOID += "." + key
			idx[smcfg.ID] = metric
		}
		for k, smcfg := range m.cfg.evalMetric {
			metric, err := NewSnmpMetric(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [indexed] [evaluated] fields metric  %d: Error: %s ", k, err)
				continue
			}
			metric.SetLogger(m.log)
			metric.RealOID = m.cfg.ID + "." + smcfg.ID + "." + key //unique identificator for this metric
			idx[smcfg.ID] = metric
		}
		//setup visibility on db for each metric
		for k, v := range idx {
			report := true
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

func (m *InfluxMeasurement) PopMetricTable(p map[string]string) error {
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
func (m *InfluxMeasurement) InitMetricTable() {
	m.MetricTable = make(map[string]map[string]*SnmpMetric)

	//create metrics.
	switch m.cfg.GetMode {
	case "value":
		//for each field
		idx := make(map[string]*SnmpMetric)
		for k, smcfg := range m.cfg.fieldMetric {
			m.log.Debugf("initializing [value]metric cfgi %s", smcfg.ID)
			metric, err := NewSnmpMetric(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [value] field metric %d : Error: %s ", k, err)
				continue
			}
			metric.SetLogger(m.log)
			idx[smcfg.ID] = metric
		}
		for k, smcfg := range m.cfg.evalMetric {
			m.log.Debugf("initializing [value] [evaluated] metric cfg %s", smcfg.ID)
			metric, err := NewSnmpMetric(smcfg)
			if err != nil {
				m.log.Errorf("ERROR on create new [value] [evaluated] field metric %d : Error: %s ", k, err)
				continue
			}
			metric.SetLogger(m.log)
			metric.RealOID = m.cfg.ID + "." + smcfg.ID
			idx[smcfg.ID] = metric
		}
		//setup visibility on db for each metric
		for k, v := range idx {
			report := true
			for _, r := range m.cfg.Fields {
				if r.ID == k {
					report = r.Report
					break
				}
			}
			v.Report = report
		}
		m.MetricTable["0"] = idx

	case "indexed":
		//for each field an each index (previously initialized)
		for key, label := range m.CurIndexedLabels {
			idx := make(map[string]*SnmpMetric)
			m.log.Debugf("initializing [indexed] metric cfg for [%s/%s]", key, label)
			for k, smcfg := range m.cfg.fieldMetric {
				metric, err := NewSnmpMetric(smcfg)
				if err != nil {
					m.log.Errorf("ERROR on create new [indexed] fields metric  %d: Error: %s ", k, err)
					continue
				}
				metric.SetLogger(m.log)
				metric.RealOID += "." + key
				idx[smcfg.ID] = metric
			}
			for k, smcfg := range m.cfg.evalMetric {
				metric, err := NewSnmpMetric(smcfg)
				if err != nil {
					m.log.Errorf("ERROR on create new [indexed] [evaluated] fields metric  %d: Error: %s ", k, err)
					continue
				}
				metric.SetLogger(m.log)
				metric.RealOID = m.cfg.ID + "." + smcfg.ID + "." + key //unique identificator for this metric
				idx[smcfg.ID] = metric
			}
			//setup visibility on db for each metric
			for k, v := range idx {
				report := true
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

func (m *InfluxMeasurement) InitBuildRuntime() {
	m.snmpOids = []string{}
	m.OidSnmpMap = make(map[string]*SnmpMetric)
	//metric level
	for kIdx, vIdx := range m.MetricTable {
		m.log.Debugf("KEY iDX %s", kIdx)
		//index level
		for kM, vM := range vIdx {
			m.log.Debugf("KEY METRIC %s OID %s", kM, vM.RealOID)
			if vM.cfg.DataSrcType != "STRINGEVAL" {
				//this array is used in SnmpGetData to send IOD's to the end device
				// so it can not contain any other thing than OID's
				// on string eval it contains a identifier not OID
				m.snmpOids = append(m.snmpOids, vM.RealOID)
			}

			m.OidSnmpMap[vM.RealOID] = vM
		}
	}
}

func (m *InfluxMeasurement) AddFilter(filter *MeasFilterCfg) error {
	var err error
	if m.cfg.GetMode == "value" {
		return fmt.Errorf("Error this measurement %s  is not indexed(snmptable) not Filter apply ", m.cfg.ID)
	}
	if filter == nil {
		return fmt.Errorf("Error invalid  NIL  filter on measurment %s ", m.cfg.ID)
	}

	m.Filter = filter
	switch m.Filter.FType {
	case "file":
		m.Filterlabels, err = m.applyFileFilter(m.Filter.FileName, m.Filter.EnableAlias)
		if err != nil {
			m.log.Errorf("Error while trying to apply file Filter  for measurement %s: ERROR: %s", m.cfg.ID, err)
		}
	case "OIDCondition":
		m.Filterlabels, err = m.applyOIDCondFilter(m.Filter.OIDCond, m.Filter.CondType, m.Filter.CondValue)
		if err != nil {
			m.log.Errorf("Error while trying to apply condition Filter  for measurement %s: ERROR: %s", m.cfg.ID, err)
		}
	default:
		return fmt.Errorf("Invalid Filter Type %s for measurement: %s", m.Filter.FType, m.cfg.ID)
	}
	//now we have the 	m.Filterlabels array initialized with only those values which we will need
	//Loading final Values to query with snmp
	m.CurIndexedLabels = m.filterIndexedLabels(m.Filter.FType, m.Filterlabels)

	m.InitMetricTable()
	return err
}

func (m *InfluxMeasurement) UpdateFilter() (bool, error) {
	var err error
	var newfilterlabels map[string]string

	if m.cfg.GetMode == "value" {
		return false, fmt.Errorf("Error this measurement %s  is not indexed(snmptable) not Filter apply ", m.cfg.ID)
	}

	//fist update  all indexed--------
	m.log.Infof("Re Loading Indexed values in : %s", m.cfg.ID)
	m.AllIndexedLabels, err = m.loadIndexedLabels()
	if err != nil {
		m.log.Errorf("Error while trying to reload Indexed Labels on for measurement %s for baseOid %s : ERROR: %s", m.cfg.ID, m.cfg.IndexOID, err)
		return false, err
	}
	if m.Filter == nil {
		m.log.Debugf("There is no filter configured in this measurement %s", m.cfg.ID)
		//check if curindexed different of AllIndexed
		delIndexes := diffKeyValuesInMap(m.CurIndexedLabels, m.AllIndexedLabels)
		newIndexes := diffKeyValuesInMap(m.AllIndexedLabels, m.CurIndexedLabels)

		if len(newIndexes) == 0 && len(delIndexes) == 0 {
			//no changes on the Filter
			m.log.Infof("No changes on the Index for measurement: %s", m.cfg.ID)
			return false, nil
		}
		m.CurIndexedLabels = m.AllIndexedLabels

		m.log.Debug("NEW INDEXES: %+v", newIndexes)
		m.log.Debug("DELETED INDEXES: %+v", delIndexes)

		if len(delIndexes) > 0 {
			m.PopMetricTable(delIndexes)
		}
		if len(newIndexes) > 0 {
			m.PushMetricTable(newIndexes)
		}
		return true, nil
	}
	//----------------
	switch m.Filter.FType {
	case "file":
		newfilterlabels, err = m.applyFileFilter(m.Filter.FileName, m.Filter.EnableAlias)
		if err != nil {
			m.log.Errorf("Error while trying to apply file Filter  for measurement %s: ERROR: %s", m.cfg.ID, err)
		}
	case "OIDCondition":
		newfilterlabels, err = m.applyOIDCondFilter(m.Filter.OIDCond, m.Filter.CondType, m.Filter.CondValue)
		if err != nil {
			m.log.Errorf("Error while trying to apply condition Filter  for measurement %s: ERROR: %s", m.cfg.ID, err)
		}
	default:
		return false, fmt.Errorf("Invalid Filter Type %s for measurement: %s", m.Filter.FType, m.cfg.ID)
	}
	//check if newfilterlabels are diferent than previous.

	//now we have the 	m.Filter,m.ls array initialized with only those values which we will need
	//Loading final Values to query with snmp
	newIndexedLabels := m.filterIndexedLabels(m.Filter.FType, newfilterlabels)

	delIndexes := diffKeyValuesInMap(m.CurIndexedLabels, newIndexedLabels)
	newIndexes := diffKeyValuesInMap(newIndexedLabels, m.CurIndexedLabels)

	if len(newIndexes) == 0 && len(delIndexes) == 0 {
		//no changes on the Filter
		m.log.Infof("No changes on the filter %s for measurement: %s", m.Filter.FType, m.cfg.ID)
		return false, nil
	}

	m.log.Debug("NEW INDEXES: %+v", newIndexes)
	m.log.Debug("DELETED INDEXES: %+v", delIndexes)

	m.Filterlabels = newfilterlabels
	m.CurIndexedLabels = newIndexedLabels

	if len(delIndexes) > 0 {
		m.PopMetricTable(delIndexes)
	}
	if len(newIndexes) > 0 {
		m.PushMetricTable(newIndexes)
	}

	return true, nil
}

func (m *InfluxMeasurement) printConfig() {
	if m.cfg.GetMode == "indexed" {
		fmt.Printf("-----------------------------------------------------------\n")
		fmt.Printf(" ** Indexed by OID: %s (TagName: %s) **\n", m.cfg.IndexOID, m.cfg.IndexTag)
		fmt.Printf("-----------------------------------------------------------\n")
	}

	if m.Filter != nil {
		switch m.Filter.FType {
		case "file":
			fmt.Printf(" ----------------------------------------------------------\n")
			fmt.Printf(" File Filter: %s ( EnableAlias: %t)\n [ TOTAL: %d| NON FILTERED: %d]", m.Filter.FileName, m.Filter.EnableAlias, len(m.AllIndexedLabels), len(m.Filterlabels))
			fmt.Printf(" ----------------------------------------------------------\n")
		case "OIDCondition":
			fmt.Printf(" ----------------------------------------------------------\n")
			fmt.Printf(" OID Condition Filter: %s ( [%s] %s) [ TOTAL: %d| NON FILTERED: %d] \n", m.Filter.OIDCond, m.Filter.CondType, m.Filter.CondValue, len(m.AllIndexedLabels), len(m.Filterlabels))
			fmt.Printf(" ----------------------------------------------------------\n")
		}
	}

	for _, v := range m.cfg.fieldMetric {
		if v.IsTag == true {
			fmt.Printf("\t*TAG[%s]\tTagName[%s]\tOID:%s\t(%s) \n", v.ID, v.FieldName, v.BaseOID, v.DataSrcType)
		} else {
			fmt.Printf("\t*Metric[%s]\tName[%s]\tOID:%s\t(%s) \n", v.ID, v.FieldName, v.BaseOID, v.DataSrcType)
		}
	}

	for _, v := range m.cfg.evalMetric {
		if v.IsTag == true {
			fmt.Printf("\t*EVALUATED TAG[%s]\tTagName[%s]\t(%s)EVAL:%s \n", v.ID, v.FieldName, v.DataSrcType, v.ExtraData)
		} else {
			fmt.Printf("\t*EVALUATED Metric[%s]\tName[%s]\t(%s)EVAL:%s \n", v.ID, v.FieldName, v.DataSrcType, v.ExtraData)
		}
	}

	if m.cfg.GetMode == "indexed" {
		fmt.Printf(" ---------------------------------------------------------\n")
		for k, v := range m.CurIndexedLabels {
			fmt.Printf("\t\tIndex[%s / %s]\n", k, v)
		}
	}
}

//GetInfluxPoint get points from measuremnetsl
func (m *InfluxMeasurement) GetInfluxPoint(hostTags map[string]string) []*client.Point {
	var ptarray []*client.Point

	switch m.cfg.GetMode {
	case "value":
		k := m.MetricTable["0"]
		var t time.Time
		Fields := make(map[string]interface{})
		for _, v_mtr := range k {
			if v_mtr.CookedValue == nil {
				m.log.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", v_mtr.cfg.ID, m.cfg.ID, hostTags, v_mtr)
				continue
			}
			if v_mtr.Report == false {
				m.log.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.cfg.ID, m.cfg.ID)
				continue
			}
			m.log.Debugf("generating field for %s value %f ", v_mtr.cfg.FieldName, v_mtr.CookedValue)
			m.log.Debugf("DEBUG METRIC %+v", v_mtr)
			Fields[v_mtr.cfg.FieldName] = v_mtr.CookedValue
			t = v_mtr.CurTime
		}
		m.log.Debug("FIELDS:%+v", Fields)

		pt, err := client.NewPoint(
			m.cfg.Name,
			hostTags,
			Fields,
			t,
		)
		if err != nil {
			m.log.Warnf("error in influx point building:%s", err)
		} else {
			m.log.Debugf("GENERATED INFLUX POINT[%s] value: %+v", m.cfg.Name, pt)
			ptarray = append(ptarray, pt)
		}

	case "indexed":
		var t time.Time
		for k_idx, v_idx := range m.MetricTable {
			m.log.Debugf("generating influx point for indexed %s", k_idx)
			//copy tags and add index tag
			Tags := make(map[string]string)
			for k_t, v_t := range hostTags {
				Tags[k_t] = v_t
			}
			Tags[m.cfg.IndexTag] = k_idx
			m.log.Debugf("IDX :%+v", v_idx)
			Fields := make(map[string]interface{})
			for _, v_mtr := range v_idx {
				m.log.Debugf("DEBUG METRIC %+v", v_mtr.cfg)
				if v_mtr.cfg.IsTag == true {
					if v_mtr.CookedValue == nil {
						m.log.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", v_mtr.cfg.ID, m.cfg.ID, Tags, v_mtr)
						continue
					}
					if v_mtr.Report == false {
						m.log.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.cfg.ID, m.cfg.ID)
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
					m.log.Debugf("generating Tag for Metric: %s : tagname: %s", v_mtr.cfg.FieldName, tag)
					Tags[v_mtr.cfg.FieldName] = tag
				} else {
					if v_mtr.CookedValue == nil {
						m.log.Warnf("Warning METRIC ID [%s] from MEASUREMENT[ %s ] with TAGS [%+v] has no valid data => See Metric Runtime [ %+v ]", v_mtr.cfg.ID, m.cfg.ID, Tags, v_mtr)
						continue
					}
					if v_mtr.Report == false {
						m.log.Debugf("REPORT is FALSE in METRIC ID [%s] from MEASUREMENT[ %s ] won't be reported to the output backend", v_mtr.cfg.ID, m.cfg.ID)
						continue
					}
					m.log.Debugf("generating field for Metric: %s : value %f", v_mtr.cfg.FieldName, v_mtr.CookedValue.(float64))
					Fields[v_mtr.cfg.FieldName] = v_mtr.CookedValue
				}

				t = v_mtr.CurTime
			}
			m.log.Debugf("FIELDS:%+v TAGS:%+v", Fields, Tags)
			pt, err := client.NewPoint(
				m.cfg.Name,
				Tags,
				Fields,
				t,
			)
			if err != nil {
				m.log.Warnf("error in influx point creation :%s", err)
			} else {
				m.log.Debugf("GENERATED INFLUX POINT[%s] index [%s]: %+v", m.cfg.Name, k_idx, pt)
				ptarray = append(ptarray, pt)
			}
		}

	}

	return ptarray

}

/*
SnmpBulkData GetSNMP Data
*/

func (m *InfluxMeasurement) SnmpWalkData() (int64, int64, error) {

	now := time.Now()
	var sent int64
	var errs int64

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.log.Debugf("received SNMP  pdu:%+v", pdu)
		sent++
		if pdu.Value == nil {
			m.log.Warnf("no value retured by pdu :%+v", pdu)
			errs++
			return nil //if error return the bulk process will stop
		}
		if metric, ok := m.OidSnmpMap[pdu.Name]; ok {
			m.log.Debugln("OK measurement ", m.cfg.ID, "SNMP RESULT OID", pdu.Name, "MetricFound", pdu.Value)
			metric.setRawData(pdu, now)
		} else {
			m.log.Debugf("returned OID from device: %s  Not Found in measurement /metric list: %+v", pdu.Name, m.cfg.ID)
		}
		return nil
	}

	for _, v := range m.cfg.fieldMetric {
		if err := m.Walk(v.BaseOID, setRawData); err != nil {
			m.log.Errorf("SNMP WALK (%s) for OID (%s) get error: %s\n", m.snmpClient.Target, v.BaseOID, err)
		}
	}

	return sent, errs, nil
}

func (m *InfluxMeasurement) ComputeEvaluatedMetrics() {
	if m.cfg.evalMetric == nil {
		m.log.Infof("Not EVAL metrics exist on measurement %s", m.cfg.ID)
		return
	}
	switch m.cfg.GetMode {
	case "value":
		parameters := make(map[string]interface{})
		m.log.Debugf("Building parrameters array for index measurement %s", m.cfg.ID)
		parameters["NR"] = len(m.CurIndexedLabels) //Number of rows (like awk)
		parameters["NF"] = len(m.cfg.fieldMetric)  //Number of fields ( like awk)
		//getting all values to the array
		for _, v := range m.cfg.fieldMetric {
			if metric, ok := m.OidSnmpMap[v.BaseOID]; ok {
				m.log.Debugf("OK Field metric found %s with FieldName %s", metric.cfg.ID, metric.cfg.FieldName)
				parameters[v.FieldName] = metric.CookedValue
			} else {
				m.log.Debugf("Evaluated metric not Found for Eval key %s", v.BaseOID)
			}
		}
		m.log.Debugf("PARAMETERS: %+v", parameters)
		//compute Evalutated metrics
		for _, v := range m.cfg.evalMetric {
			evalkey := m.cfg.ID + "." + v.ID
			if metric, ok := m.OidSnmpMap[evalkey]; ok {
				m.log.Debugln("OK Evaluated metric found", m.cfg.ID, "Eval KEY", evalkey)
				metric.Compute(parameters)
				parameters[v.FieldName] = metric.CookedValue
			} else {
				m.log.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
			}
		}
	case "indexed":
		for key, val := range m.CurIndexedLabels {
			//building parameters array
			parameters := make(map[string]interface{})
			m.log.Debugf("Building parrameters array for index %s/%s", key, val)
			parameters["NR"] = len(m.CurIndexedLabels) //Number of rows (like awk)
			parameters["NF"] = len(m.cfg.fieldMetric)  //Number of fields ( like awk)
			//TODO: add other common variables => Elapsed , etc
			//getting all values to the array
			for _, v := range m.cfg.fieldMetric {
				if metric, ok := m.OidSnmpMap[v.BaseOID+"."+key]; ok {
					m.log.Debugf("OK Field metric found %s with FieldName %s", metric.cfg.ID, metric.cfg.FieldName)
					//TODO: validate all posibles values of CookedValue
					parameters[v.FieldName] = metric.CookedValue
				} else {
					m.log.Debugf("Evaluated metric not Found for Eval key %s")
				}
			}
			m.log.Debugf("PARAMETERS: %+v", parameters)
			//compute Evalutated metrics
			for _, v := range m.cfg.evalMetric {
				evalkey := m.cfg.ID + "." + v.ID + "." + key
				if metric, ok := m.OidSnmpMap[evalkey]; ok {
					m.log.Debugln("OK Evaluated metric found", m.cfg.ID, "Eval KEY", evalkey)
					metric.Compute(parameters)
					parameters[v.ID] = metric.CookedValue
				} else {
					m.log.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
				}
			}
		}
	}
}

/*
GetSnmpData GetSNMP Data
*/

func (m *InfluxMeasurement) SnmpGetData() (int64, int64, error) {

	now := time.Now()
	var sent int64
	var errs int64
	l := len(m.snmpOids)
	for i := 0; i < l; i += maxOids {
		end := i + maxOids
		if end > l {
			end = len(m.snmpOids)
			sent += (int64(end) - int64(i))
		}
		m.log.Debugf("Getting snmp data from %d to %d", i, end)
		//	log.Printf("DEBUG oids:%+v", m.snmpOids)
		//	log.Printf("DEBUG oidmap:%+v", m.OidSnmpMap)
		pkt, err := m.snmpClient.Get(m.snmpOids[i:end])
		if err != nil {
			m.log.Debugf("selected OIDS %+v", m.snmpOids[i:end])
			m.log.Errorf("SNMP (%s) for OIDs (%d/%d) get error: %s\n", m.snmpClient.Target, i, end, err)
			errs++
			continue
		}

		for _, pdu := range pkt.Variables {
			m.log.Debugf("DEBUG pdu:%+v", pdu)
			if pdu.Value == nil {
				errs++
				continue
			}
			oid := pdu.Name
			val := pdu.Value
			if metric, ok := m.OidSnmpMap[oid]; ok {
				m.log.Debugf("OK measurement %s SNMP result OID: %s MetricFound: %+v ", m.cfg.ID, oid, val)
				metric.setRawData(pdu, now)
			} else {
				m.log.Errorln("OID", oid, "Not Found in measurement", m.cfg.ID)
			}
		}
	}

	return sent, errs, nil
}

func (m *InfluxMeasurement) loadIndexedLabels() (map[string]string, error) {

	m.log.Debugf("Looking up column names %s ", m.cfg.IndexOID)

	allindex := make(map[string]string)

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.log.Debugf("received SNMP  pdu:%+v", pdu)
		if pdu.Value == nil {
			m.log.Warnf("no value retured by pdu :%+v", pdu)
			return nil //if error return the bulk process will stop
		}
		if len(pdu.Name) < m.idxPosInOID {
			m.log.Warnf("Received PDU OID smaller  than minimal index(%d) positionretured by pdu :%+v", m.idxPosInOID, pdu)
			return nil //if error return the bulk process will stop
		}
		//i := strings.LastIndex(pdu.Name, ".")
		suffix := pdu.Name[m.idxPosInOID+1:]

		if m.cfg.IndexAsValue == true {
			allindex[suffix] = suffix
			return nil
		}
		var name string
		switch pdu.Type {
		case gosnmp.OctetString:
			name = string(pdu.Value.([]byte))
			m.log.Debugf("Got the following OctetString index for [%s/%s]", suffix, name)
		case gosnmp.Integer, gosnmp.Counter32, gosnmp.Counter64, gosnmp.Gauge32, gosnmp.Uinteger32:
			name = strconv.FormatInt(pduVal2Int64(pdu), 10)
			m.log.Debugf("Got the following Numeric index for [%s/%s]", suffix, name)
		default:
			m.log.Errorf("Error in IndexedLabel  IndexLabel %s ERR: Not String or numeric Value", m.cfg.IndexOID)
		}
		allindex[suffix] = name
		return nil
	}
	err := m.Walk(m.cfg.IndexOID, setRawData)
	if err != nil {
		m.log.Errorf("SNMP WALK error: %s", err)
		return allindex, err
	}
	return allindex, nil
}

/*
 filterIndexedLabels construct the final index array from all index and filters
*/
func (m *InfluxMeasurement) filterIndexedLabels(f_mode string, L map[string]string) map[string]string {
	curIndexedLabels := make(map[string]string, len(m.Filterlabels))

	switch f_mode {
	case "file":
		//file filter should compare with all indexed labels with the value (name)
		for k_f, v_f := range L {
			for k_l, v_l := range m.AllIndexedLabels {
				if k_f == v_l {
					if len(v_f) > 0 {
						// map[k_l]v_f (alias to key of the label
						curIndexedLabels[k_l] = v_f
					} else {
						//map[k_l]v_l (original name)
						curIndexedLabels[k_l] = v_l
					}

				}
			}
		}

	case "OIDCondition":
		for k_f, _ := range L {
			for k_l, v_l := range m.AllIndexedLabels {
				if k_f == k_l {
					curIndexedLabels[k_l] = v_l
				}
			}
		}

		//confition filter should comapre with all indexed label with the key (number)
	}
	//could be posible to a delete of the non needed arrays  m.AllIndexedLabels //m.Filterlabels
	return curIndexedLabels
}

func (m *InfluxMeasurement) applyOIDCondFilter(oidCond string, typeCond string, valueCond string) (map[string]string, error) {

	m.log.Infof("Apply Condition Filter: Looking up column names in: Condition %s", oidCond)

	idxPosInOID := len(oidCond)

	filterlabels := make(map[string]string)

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.log.Debugf("received SNMP  pdu:%+v", pdu)
		if pdu.Value == nil {
			m.log.Warnf("no value retured by pdu :%+v", pdu)
			return nil //if error return the bulk process will stop
		}
		var vci int64
		var value int64
		var cond bool

		switch {
		case typeCond == "notmatch":
			//m.log.Debugf("PDU: %+v", pdu)
			str := pduVal2str(pdu)
			var re = regexp.MustCompile(valueCond)
			matched := re.MatchString(str)
			m.log.Debugf("Evaluated notmatch condition  value: %s | filter: %s | result : %t", str, valueCond, !matched)
			cond = !matched
		case typeCond == "match":
			//m.log.Debugf("PDU: %+v", pdu)
			str := pduVal2str(pdu)
			var re = regexp.MustCompile(valueCond)
			matched := re.MatchString(str)
			m.log.Debugf("Evaluated match condition  value: %s | filter: %s | result : %t", str, valueCond, matched)
			cond = matched
		case strings.Contains(typeCond, "n"):
			//undesrstand valueCondition as numeric
			vc, err := strconv.Atoi(valueCond)
			if err != nil {
				m.log.Warnf("only accepted numeric value as value condition  current : %s  for TypeCond %s", valueCond, typeCond)
				return nil
			}
			vci = int64(vc)
			value = pduVal2Int64(pdu)
			fallthrough
		case typeCond == "neq":
			cond = (value == vci)
		case typeCond == "nlt":
			cond = (value < vci)
		case typeCond == "ngt":
			cond = (value > vci)
		case typeCond == "nge":
			cond = (value >= vci)
		case typeCond == "nle":
			cond = (value <= vci)

		default:
			m.log.Errorf("Error in Condition filter OidCondition: %s Type: %s ValCond: %s ", oidCond, typeCond, valueCond)
		}
		if cond == true {
			if len(pdu.Name) < idxPosInOID {
				m.log.Warnf("Received PDU OID smaller  than minimal index(%d) positionretured by pdu :%+v", m.idxPosInOID, pdu)
				return nil //if error return the bulk process will stop
			}
			suffix := pdu.Name[idxPosInOID+1:]
			filterlabels[suffix] = ""
		}

		return nil
	}
	err := m.Walk(oidCond, setRawData)
	if err != nil {
		m.log.Errorf("SNMP version-1 walk error : %s", err)
		return filterlabels, err
	}

	return filterlabels, nil
}

func (m *InfluxMeasurement) applyFileFilter(file string, enableAlias bool) (map[string]string, error) {
	m.log.Infof("apply File filter : %s Enable Alias: %t", file, enableAlias)
	filterlabels := make(map[string]string)
	if len(file) == 0 {
		return filterlabels, errors.New("No file configured error ")
	}
	data, err := ioutil.ReadFile(filepath.Join(confDir, file))
	if err != nil {
		m.log.Errorf("ERROR on open file %s: error: %s", filepath.Join(confDir, file), err)
		return filterlabels, err
	}

	for l_num, line := range strings.Split(string(data), "\n") {
		//		log.Println("LINIA:", line)
		// strip comments
		comment := strings.Index(line, "#")
		if comment >= 0 {
			line = line[:comment]
		}
		if len(line) == 0 {
			continue
		}
		f := strings.Fields(line)
		switch len(f) {
		case 1:
			filterlabels[f[0]] = ""

		case 2:
			if enableAlias {
				filterlabels[f[0]] = f[1]
			} else {
				filterlabels[f[0]] = ""
			}

		default:
			m.log.Warnf("wrong number of parameters in file: %s Lnum: %s num : %s line: %s", file, l_num, len(f), line)
		}
	}
	return filterlabels, nil
}
