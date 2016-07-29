package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
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

	log.Info("processing measurement key: ", name)
	log.Debug("%+v", mc)
	for _, f_val := range mc.Fields {
		log.Debug("looking for measure ", mc.Name, " fields: ", f_val)
		if val, ok := (*MetricCfg)[f_val]; ok {
			log.Debug("Found ok!")
			//map is correct
			mc.fieldMetric = append(mc.fieldMetric, val)
		} else {
			log.Warn("measurement field ", f_val, " NOT FOUND in Metrics Database !")
		}
	}
	//check if fieldMetric
	if len(mc.fieldMetric) == 0 {
		var s string
		for _, v := range mc.Fields {
			s += v
		}
		return errors.New("No metrics found with names" + s + " in measurement Config " + mc.ID)
	}
	return nil
}

/*
type MeasFilterCfg struct {
	fType       string //file/OidCondition
	FileName    string
	enableAlias bool
	OIDCond     string
	condType    string
	condValue   string
}*/

//InfluxMeasurement the runtime measurement config
type InfluxMeasurement struct {
	cfg              *InfluxMeasurementCfg
	values           map[string]map[string]*SnmpMetric //snmpMetric mapped with metric_names and Index
	snmpOids         []string
	oidSnmpMap       map[string]*SnmpMetric //snmpMetric mapped with real OID's
	Filterlabels     map[string]string
	AllIndexedLabels map[string]string //all available values on the remote device
	CurIndexedLabels map[string]string
	Filter           *MeasFilterCfg
	log              *logrus.Logger
	numValOrig       int //num of values in the full index
	numValFlt        int //num of final indexed values after filter applyed
}

func (m *InfluxMeasurement) printConfig() {
	if m.Filter != nil {
		switch m.Filter.FType {
		case "file":
			fmt.Printf(" ----------------------------------------------------------\n")
			fmt.Printf(" File Filter: %s ( EnableAlias: %t)\n [ TOTAL: %d| FILTERED: %d]", m.Filter.FileName, m.Filter.EnableAlias, m.numValOrig, m.numValFlt)
			fmt.Printf(" ----------------------------------------------------------\n")
		case "OIDCondition":
			fmt.Printf(" ----------------------------------------------------------\n")
			fmt.Printf(" OID Condition Filter: %s ( [%s] %s) [ TOTAL: %d| FILTERED: %d] \n", m.Filter.OIDCond, m.Filter.CondType, m.Filter.CondValue, m.numValOrig, m.numValFlt)
			fmt.Printf(" ----------------------------------------------------------\n")
		}

	}
	for _, v := range m.cfg.fieldMetric {
		fmt.Printf("\t*Metric[%s]\tName[%s]\tOID:%s\t(%s) \n", v.ID, v.FieldName, v.BaseOID, v.DataSrcType)
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
		k := m.values["0"]
		var t time.Time
		Fields := make(map[string]interface{})
		for _, v_mtr := range k {
			m.log.Debugf("generating field for %s value %s ", v_mtr.cfg.FieldName, v_mtr.cookedValue)
			m.log.Debugf("DEBUG METRIC %+v", v_mtr)
			Fields[v_mtr.cfg.FieldName] = v_mtr.cookedValue
			t = v_mtr.curTime
		}
		m.log.Debug("FIELDS:%+v", Fields)
		//	m.log.Debug("TAGS:%+v", FullTags)

		pt, err := client.NewPoint(
			m.cfg.Name,
			//FullTags,
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
		for k_idx, v_idx := range m.values {
			m.log.Debugf("generating influx point for indexed %s", k_idx)
			//copy tags and add index tag
			Tags := make(map[string]string)
			//for k_t, v_t := range FullTags {
			for k_t, v_t := range hostTags {
				Tags[k_t] = v_t
			}
			Tags[m.cfg.IndexTag] = k_idx
			m.log.Debugf("IDX :%+v", v_idx)
			Fields := make(map[string]interface{})
			for _, v_mtr := range v_idx {
				m.log.Debugf("DEBUG METRIC %+v", v_mtr.cfg)
				m.log.Debugf("generating field for Metric: %s", v_mtr.cfg.FieldName)
				Fields[v_mtr.cfg.FieldName] = v_mtr.cookedValue
				t = v_mtr.curTime
			}
			m.log.Debugf("FIELDS:%+v", Fields)
			m.log.Debugf("TAGS:%+v", Tags)
			pt, err := client.NewPoint(
				m.cfg.Name,
				Tags,
				Fields,
				t,
			)
			if err != nil {
				m.log.Warnf("error in influx point creation :%s", err)
			} else {
				m.log.Debugf("DEBUG INFLUX POINT[%s] index [%s]: %+v", m.cfg.Name, k_idx, pt)
				ptarray = append(ptarray, pt)
			}
		}

	}

	return ptarray

}

/*
SnmpBulkData GetSNMP Data
*/

func (m *InfluxMeasurement) SnmpBulkData(snmp *gosnmp.GoSNMP) (int64, int64, error) {

	now := time.Now()
	var sent int64
	var errs int64

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.log.Debugf("received SNMP  pdu:%+v", pdu)
		if pdu.Value == nil {
			m.log.Warnf("no value retured by pdu :%+v", pdu)
			return nil //if error return the bulk process will stop
		}
		if metric, ok := m.oidSnmpMap[pdu.Name]; ok {
			m.log.Debugln("OK measurement ", m.cfg.ID, "SNMP RESULT OID", pdu.Name, "MetricFound", pdu.Value)
			metric.setRawData(pduVal2Int64(pdu), now)
		} else {
			m.log.Debugf("returned OID from device: %s  Not Found in measurement /metric list: %+v", pdu.Name, m.cfg.ID)
		}
		return nil
	}
	for _, v := range m.cfg.fieldMetric {
		if err := snmp.BulkWalk(v.BaseOID, setRawData); err != nil {
			m.log.Errorf("SNMP (%s) for OID (%s) get error: %s\n", snmp.Target, v.BaseOID, err)
			errs++
		}
		sent++
	}

	return sent, errs, nil
}

/*
GetSnmpData GetSNMP Data
*/

func (m *InfluxMeasurement) SnmpGetData(snmp *gosnmp.GoSNMP) (int64, int64, error) {

	now := time.Now()
	var sent int64
	var errs int64
	l := len(m.snmpOids)
	for i := 0; i < l; i += maxOids {
		end := i + maxOids
		if end > l {
			end = len(m.snmpOids)
		}
		m.log.Debugf("Getting snmp data from %d to %d", i, end)
		//	log.Printf("DEBUG oids:%+v", m.snmpOids)
		//	log.Printf("DEBUG oidmap:%+v", m.oidSnmpMap)
		pkt, err := snmp.Get(m.snmpOids[i:end])
		if err != nil {
			m.log.Debugf("selected OIDS %+v", m.snmpOids[i:end])
			m.log.Errorf("SNMP (%s) for OIDs (%d/%d) get error: %s\n", snmp.Target, i, end, err)
			errs++
			continue

		}
		sent++
		for _, pdu := range pkt.Variables {
			m.log.Debugf("DEBUG pdu:%+v", pdu)
			if pdu.Value == nil {
				continue
			}
			oid := pdu.Name
			val := pdu.Value
			if metric, ok := m.oidSnmpMap[oid]; ok {
				m.log.Debugf("OK measurement %s SNMP result OID: %s MetricFound: %s ", m.cfg.ID, oid, val)
				metric.setRawData(pduVal2Int64(pdu), now)
			} else {
				m.log.Errorln("OID", oid, "Not Found in measurement", m.cfg.ID)
			}
		}
	}

	return sent, errs, nil
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key, _ := range encountered {
		result = append(result, key)
	}
	return result
}

func (m *InfluxMeasurement) loadIndexedLabels(c *SnmpDevice) error {
	client := c.snmpClient
	m.log.Debugf("Looking up column names for: %s  NAMES %s ", c.cfg.Host, m.cfg.IndexOID)
	pdus, err := client.BulkWalkAll(m.cfg.IndexOID)
	if err != nil {
		m.log.Fatalln("SNMP bulkwalk error", err)
	}
	m.AllIndexedLabels = make(map[string]string)
	m.numValOrig = 0
	for _, pdu := range pdus {
		switch pdu.Type {
		case gosnmp.OctetString:
			i := strings.LastIndex(pdu.Name, ".")
			suffix := pdu.Name[i+1:]
			name := string(pdu.Value.([]byte))
			m.AllIndexedLabels[suffix] = name
			m.numValOrig++
			m.log.Debugf("Got the following index for %c :[%s/%s]", c.cfg.Host, suffix, name)
		default:
			m.log.Errorf("Error in IndexedLabel for host: %s  IndexLabel %s ERR: Not String", c.cfg.Host, m.cfg.IndexOID)
		}
	}
	return nil
}

/*
 filterIndexedLabels construct the final index array from all index and filters
*/
func (m *InfluxMeasurement) filterIndexedLabels(f_mode string) error {
	m.CurIndexedLabels = make(map[string]string, len(m.Filterlabels))

	switch f_mode {
	case "file":
		//file filter should compare with all indexed labels with the value (name)
		for k_f, v_f := range m.Filterlabels {
			for k_l, v_l := range m.AllIndexedLabels {
				if k_f == v_l {
					if len(v_f) > 0 {
						// map[k_l]v_f (alias to key of the label
						m.CurIndexedLabels[k_l] = v_f
					} else {
						//map[k_l]v_l (original name)
						m.CurIndexedLabels[k_l] = v_l
					}

				}
			}
		}

	case "OIDCondition":
		for k_f, _ := range m.Filterlabels {
			for k_l, v_l := range m.AllIndexedLabels {
				if k_f == k_l {
					m.CurIndexedLabels[k_l] = v_l
				}
			}
		}

		//confition filter should comapre with all indexed label with the key (number)
	}

	//could be posible to a delete of the non needed arrays  m.AllIndexedLabels //m.Filterlabels
	return nil
}

//IndexedLabels
func (m *InfluxMeasurement) IndexedLabels() error {
	m.CurIndexedLabels = m.AllIndexedLabels
	return nil
}

func (m *InfluxMeasurement) applyOIDCondFilter(c *SnmpDevice, oidCond string, typeCond string, valueCond string) error {
	client := c.snmpClient
	m.log.Infof("Apply Condition Filter: Looking up column names in: %s Condition %s", c.cfg.Host, oidCond)
	pdus, err := client.BulkWalkAll(oidCond)
	if err != nil {
		m.log.Fatalf("SNMP bulkwalk error : %s", err)
	}
	m.Filterlabels = make(map[string]string)
	vc, err := strconv.Atoi(valueCond)
	if err != nil {
		return errors.New("only accepted numeric value as value condition current :" + valueCond)
	}
	m.numValFlt = 0
	vci := int64(vc)
	for _, pdu := range pdus {
		value := pduVal2Int64(pdu)
		var cond bool
		switch typeCond {
		case "eq":
			cond = (value == vci)
		case "lt":
			cond = (value < vci)
		case "gt":
			cond = (value > vci)
		case "ge":
			cond = (value >= vci)
		case "le":
			cond = (value <= vci)
		default:
			m.log.Errorf("Error in Condition filter for host: %s OidCondition: %s Type: %s ValCond: %s ", c.cfg.Host, oidCond, typeCond, valueCond)
		}
		if cond == true {
			i := strings.LastIndex(pdu.Name, ".")
			suffix := pdu.Name[i+1:]
			m.Filterlabels[suffix] = ""
			m.numValFlt++
		}

	}
	return nil
}

func (m *InfluxMeasurement) applyFileFilter(file string, enableAlias bool) error {
	m.log.Infof("apply File filter : %s Enable Alias: %s", file, enableAlias)
	if len(file) == 0 {
		return errors.New("File error ")
	}
	data, err := ioutil.ReadFile(filepath.Join(confDir, file))
	if err != nil {
		m.log.Fatal(err)
	}
	m.Filterlabels = make(map[string]string)
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
			m.Filterlabels[f[0]] = ""

		case 2:
			if enableAlias {
				m.Filterlabels[f[0]] = f[1]
			} else {
				m.Filterlabels[f[0]] = ""
			}

		default:
			m.log.Warnf("wrong number of parameters in file: %s Lnum: %s num : %s line: %s", file, l_num, len(f), line)
		}
	}
	return nil
}
