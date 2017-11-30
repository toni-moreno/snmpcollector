package measurement

import (
	"fmt"

	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/filter"
	"github.com/toni-moreno/snmpcollector/pkg/data/metric"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

var (
	confDir string              //Needed to get File Filters data
	dbc     *config.DatabaseCfg //Needed to get Custom Filter  data
)

// SetConfDir  enable load File Filters from anywhere in the our FS.
func SetConfDir(dir string) {
	confDir = dir
	metric.SetConfDir(dir)
}

// SetDB load database config to load data if needed (used in filters)
func SetDB(db *config.DatabaseCfg) {
	dbc = db
	metric.SetDB(db)
}

//Measurement the runtime measurement config
type Measurement struct {
	cfg              *config.MeasurementCfg
	ID               string
	MName            string
	TagName          string
	MetricTable      *MetricTable
	snmpOids         []string
	OidSnmpMap       map[string]*metric.SnmpMetric `json:"-"` //snmpMetric mapped with real OID's
	AllIndexedLabels map[string]string             //`json:"-"` //all available values on the remote device
	CurIndexedLabels map[string]string             //`json:"-"`
	idxPosInOID      int
	idx2PosInOID     int
	curIdxPos        int //used in Walk functions could be variable depending on the Index (or IndexTag)
	FilterCfg        *config.MeasFilterCfg
	Filter           filter.Filter
	log              *logrus.Logger
	snmpClient       *gosnmp.GoSNMP
	DisableBulk      bool                                `json:"-"`
	GetData          func() (int64, int64, int64, error) `json:"-"`
	Walk             func(string, gosnmp.WalkFunc) error `json:"-"`
}

//New  creates object with config , log + goSnmp client
func New(c *config.MeasurementCfg, l *logrus.Logger, cli *gosnmp.GoSNMP, db bool) (*Measurement, error) {
	m := &Measurement{ID: c.ID, MName: c.Name, cfg: c, log: l, snmpClient: cli, DisableBulk: db}
	err := m.Init()
	return m, err
}

// InvalidateMetrics Invalidate all MetricTable metrics
func (m *Measurement) InvalidateMetrics() {
	//invalidate normal metrics
	m.MetricTable.InvalidateTable()
}

/*Init does:
 *inicialize AllIndexesLabels
 *Assign CurIndexedLabels to all Labels (until filters set)
 *init MetricTable
 */
func (m *Measurement) Init() error {
	//Init snmp methods
	switch m.cfg.GetMode {
	case "value":
		m.GetData = m.SnmpGetData
	default:
		m.GetData = m.SnmpWalkData
	}

	switch {
	case m.snmpClient.Version == gosnmp.Version1 || m.DisableBulk:
		m.Walk = m.snmpClient.Walk
	default:
		m.Walk = m.snmpClient.BulkWalk
	}

	//loading all posible values in 	m.AllIndexedLabels
	if m.cfg.GetMode == "indexed" || m.cfg.GetMode == "indexed_it" {
		m.idxPosInOID = len(m.cfg.IndexOID)
		m.TagName = m.cfg.IndexTag
		if (m.cfg.GetMode) == "indexed_it" {
			m.idx2PosInOID = len(m.cfg.TagOID)
		}
		m.Infof("Loading Indexed values")
		il, err := m.loadIndexedLabels()

		if err != nil {
			m.Errorf("Error while trying to load Indexed Labels on for measurement : for baseOid %s : ERROR: %s", m.cfg.IndexOID, err)
			return err
		}
		m.AllIndexedLabels = il
		//Final Selected Indexes are All Indexed
		m.CurIndexedLabels = m.AllIndexedLabels
	}

	/********************************
	 * Initialize Metric Runtime data in one array m-values
	 * ******************************/
	m.Debug("Initialize OID measurement per label => map of metric object per field | OID array [ready to send to the walk device] | OID=>Metric MAP")
	m.MetricTable = NewMetricTable(m.cfg, m.log, m.CurIndexedLabels)
	return nil
}

// SetSnmpClient set a GoSNMP client to the Measurement
func (m *Measurement) SetSnmpClient(cli *gosnmp.GoSNMP) {

	m.snmpClient = cli

	switch {
	case m.snmpClient.Version == gosnmp.Version1 || m.DisableBulk:
		m.Walk = m.snmpClient.Walk
	default:
		m.Walk = m.snmpClient.BulkWalk
	}
}

// GetMode Returns mode info
func (m *Measurement) GetMode() string {
	return m.cfg.GetMode
}

// InitBuildRuntime init
func (m *Measurement) InitBuildRuntime() {

	switch m.cfg.GetMode {
	case "value":
		m.snmpOids, m.OidSnmpMap = m.MetricTable.GetSnmpMaps()
	default:
		m.OidSnmpMap = m.MetricTable.GetSnmpMap()
	}

}

// AddFilter attach a filtering process to the measurement
func (m *Measurement) AddFilter(f *config.MeasFilterCfg) error {
	var err error
	if m.cfg.GetMode == "value" {
		return fmt.Errorf("Error this measurement %s  is not indexed(snmptable) not Filter apply ", m.cfg.ID)
	}
	if f == nil {
		return fmt.Errorf("Error invalid  NIL  filter on measurment %s ", m.cfg.ID)
	}

	m.FilterCfg = f
	switch m.FilterCfg.FType {
	case "file":
		m.Filter = filter.NewFileFilter(m.FilterCfg.FilterName, m.FilterCfg.EnableAlias, m.log)
		err = m.Filter.Init(confDir)
		if err != nil {
			return fmt.Errorf("Error invalid File Filter : %s", err)
		}
	case "OIDCondition":
		cond, err2 := dbc.GetOidConditionCfgByID(m.FilterCfg.FilterName)
		if err2 != nil {
			m.Errorf("Error getting filter id %s OIDCondition [id: %s ] data : %s", m.FilterCfg.ID, m.FilterCfg.FilterName, err)
		}

		if cond.IsMultiple {
			m.Filter = filter.NewOidMultipleFilter(cond.OIDCond, m.log)
			err = m.Filter.Init(m.Walk, dbc)
			if err != nil {
				return fmt.Errorf("Error invalid Multiple Condition Filter : %s", err)
			}
		} else {
			m.Filter = filter.NewOidFilter(cond.OIDCond, cond.CondType, cond.CondValue, m.log)
			err = m.Filter.Init(m.Walk)
			if err != nil {
				return fmt.Errorf("Error invalid OID condition Filter : %s", err)
			}
		}

	case "custom":
		m.Filter = filter.NewCustomFilter(m.FilterCfg.FilterName, m.FilterCfg.EnableAlias, m.log)
		err = m.Filter.Init(dbc)
		if err != nil {
			return fmt.Errorf("Error invalid Custom Filter : %s", err)
		}
	default:
		return fmt.Errorf("Invalid Filter Type %s for measurement: %s", m.FilterCfg.FType, m.cfg.ID)
	}

	err = m.Filter.Update()
	if err != nil {
		m.Errorf("Error while trying to apply file Filter  ERROR: %s", err)
	}

	//now we have the 	m.Filterlabels array initialized with only those values which we will need
	//Loading final Values to query with snmp
	m.CurIndexedLabels = m.Filter.MapLabels(m.AllIndexedLabels)
	m.MetricTable = NewMetricTable(m.cfg, m.log, m.CurIndexedLabels)
	return err
}

// UpdateFilter reload indexed with filters
func (m *Measurement) UpdateFilter() (bool, error) {
	var err error
	//var newfilterlabels map[string]string

	if m.cfg.GetMode == "value" {
		return false, fmt.Errorf("Error this measurement %s  is not indexed(snmptable) not Filter apply ", m.cfg.ID)
	}

	//fist update  all indexed--------
	m.Infof("Re Loading Indexed values")
	il, err2 := m.loadIndexedLabels()

	if err2 != nil {
		m.Errorf("Error while trying to reload Indexed Labels for baseOid %s : ERROR: %s", m.cfg.IndexOID, err)
		return false, err
	}
	m.AllIndexedLabels = il
	if m.Filter == nil {
		m.Debugf("There is no filter configured in this measurement %s", m.cfg.ID)
		//check if curindexed different of AllIndexed
		delIndexes := utils.DiffKeyValuesInMap(m.CurIndexedLabels, m.AllIndexedLabels)
		newIndexes := utils.DiffKeyValuesInMap(m.AllIndexedLabels, m.CurIndexedLabels)

		if len(newIndexes) == 0 && len(delIndexes) == 0 {
			//no changes on the Filter
			m.Infof("No changes found on the Index for this measurement")
			return false, nil
		}
		m.CurIndexedLabels = m.AllIndexedLabels

		m.Debugf("NEW INDEXES: %+v", newIndexes)
		m.Debugf("DELETED INDEXES: %+v", delIndexes)

		if len(delIndexes) > 0 {
			m.MetricTable.Pop(delIndexes)
		}
		if len(newIndexes) > 0 {
			m.MetricTable.Push(newIndexes)
		}
		return true, nil
	}
	//----------------
	m.Infof("Applying filter : [ %s ] on measurement", m.FilterCfg.ID)

	err = m.Filter.Update()
	if err != nil {
		m.Errorf("Error while trying to apply Filter : ERROR: %s", err)
		return false, err
	}
	//check if all values have been filtered to send a warnign message.
	if m.Filter.Count() == 0 {
		m.Warnf("WARNING after applying filter no values on this measurement will be sent")
	}
	//check if newfilterlabels are different than previous.

	//now we have the 	m.Filter,m.ls array initialized with only those values which we will need
	//Loading final Values to query with snmp
	newIndexedLabels := m.Filter.MapLabels(m.AllIndexedLabels)

	delIndexes := utils.DiffKeyValuesInMap(m.CurIndexedLabels, newIndexedLabels)
	newIndexes := utils.DiffKeyValuesInMap(newIndexedLabels, m.CurIndexedLabels)

	if len(newIndexes) == 0 && len(delIndexes) == 0 {
		//no changes on the Filter
		m.Infof("No changes on the filter %s ", m.FilterCfg.FType)
		return false, nil
	}

	m.Debugf("NEW INDEXES: %+v", newIndexes)
	m.Debugf("DELETED INDEXES: %+v", delIndexes)

	m.CurIndexedLabels = newIndexedLabels

	if len(delIndexes) > 0 {
		m.MetricTable.Pop(delIndexes)
	}
	if len(newIndexes) > 0 {
		m.MetricTable.Push(newIndexes)
	}

	return true, nil
}

/*
SnmpBulkData GetSNMP Data
*/

// SnmpWalkData get data with snmpwalk
func (m *Measurement) SnmpWalkData() (int64, int64, int64, error) {

	now := time.Now()
	var gathered int64
	var processed int64
	var errors int64

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.Debugf("received SNMP  pdu:%+v", pdu)
		gathered++
		if pdu.Value == nil {
			m.Warnf("no value retured by pdu :%+v", pdu)
			errors++
			return nil //if error return the bulk process will stop
		}
		if metr, ok := m.OidSnmpMap[pdu.Name]; ok {
			m.Debugf("OK measurement %s SNMP RESULT OID %s MetricFound", pdu.Name, pdu.Value)
			processed++
			metr.SetRawData(pdu, now)
		} else {
			m.Debugf("returned OID from device: %s  Not Found in measurement /metr list: %+v", pdu.Name, m.cfg.ID)
		}
		return nil
	}

	for _, v := range m.cfg.FieldMetric {
		if err := m.Walk(v.BaseOID, setRawData); err != nil {
			m.Errorf("SNMP WALK (%s) for OID (%s) get error: %s\n", m.snmpClient.Target, v.BaseOID, err)
			errors += int64(m.MetricTable.Len())
		}
	}

	return gathered, processed, errors, nil
}

// ComputeOidConditionalMetrics take OID contitional metrics and computes true value
func (m *Measurement) ComputeOidConditionalMetrics() {
	if m.cfg.OidCondMetric == nil {
		m.Infof("Not Oid CONDITIONEVAL metrics exist on this measurement")
		return
	}
	switch m.cfg.GetMode {
	case "value":
		//compute Evalutated metrics
		for _, v := range m.cfg.OidCondMetric {
			evalkey := m.cfg.ID + "." + v.ID
			if metr, ok := m.OidSnmpMap[evalkey]; ok {
				m.Debugf("OK OidCondition  metric found %s Eval KEY", evalkey)
				metr.Compute(m.Walk, dbc)
			} else {
				m.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
			}
		}
	default:
		m.Warnf("Warning there is CONDITIONAL metrics on indexed measurements!!")
	}
}

// ComputeEvaluatedMetrics take evaluated metrics and computes them from the other values
func (m *Measurement) ComputeEvaluatedMetrics(catalog map[string]interface{}) {
	if m.cfg.EvalMetric == nil {
		m.Infof("Not EVAL metrics exist on  this measurement")
		return
	}

	parameters := make(map[string]interface{})
	//copy of the catalog map
	for k, v := range catalog {
		parameters[k] = v
	}

	//copy the input
	switch m.cfg.GetMode {
	case "value":
		m.Debugf("Building parrameters array for index measurement %s", m.cfg.ID)
		parameters["NR"] = len(m.CurIndexedLabels) //Number of rows (like awk)
		parameters["NF"] = len(m.cfg.FieldMetric)  //Number of fields ( like awk)
		//getting all values to the array
		for _, v := range m.cfg.FieldMetric {
			if metr, ok := m.OidSnmpMap[v.BaseOID]; ok {
				m.Debugf("OK Field metric found %s with FieldName %s", metr.ID, metr.GetFieldName())
				if metr.Valid == true { //only valid for compute if it has been updated last
					parameters[v.FieldName] = metr.CookedValue
				}
			} else {
				m.Debugf("Evaluated metric not Found for Eval key %s", v.BaseOID)
			}
		}
		m.Debugf("PARAMETERS: %+v", parameters)
		//compute Evalutated metrics
		for _, v := range m.cfg.EvalMetric {
			evalkey := m.cfg.ID + "." + v.ID
			if metr, ok := m.OidSnmpMap[evalkey]; ok {
				m.Debugf("OK Evaluated metric found %s Eval KEY", evalkey)
				metr.Compute(parameters)
				if metr.Valid == true { // this computed metric could be used to compute others only if valid
					parameters[v.FieldName] = metr.CookedValue
				}
			} else {
				m.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
			}
		}
	case "indexed", "indexed_it":
		for key, val := range m.CurIndexedLabels {
			//building parameters array
			m.Debugf("Building parrameters array for index %s/%s", key, val)
			parameters["NR"] = len(m.CurIndexedLabels) //Number of rows (like awk)
			parameters["NF"] = len(m.cfg.FieldMetric)  //Number of fields ( like awk)
			//TODO: add other common variables => Elapsed , etc
			//getting all values to the array
			for _, v := range m.cfg.FieldMetric {
				if metr, ok := m.OidSnmpMap[v.BaseOID+"."+key]; ok {
					m.Debugf("OK Field metric found %s with FieldName %s", metr.ID, metr.GetFieldName())
					//TODO: validate all posibles values of CookedValue
					if metr.Valid == true {
						parameters[v.FieldName] = metr.CookedValue
					}
				} else {
					m.Debugf("Evaluated metric not Found for Eval key %s")
				}
			}
			m.Debugf("PARAMETERS: %+v", parameters)
			//compute Evalutated metrics
			for _, v := range m.cfg.EvalMetric {
				evalkey := m.cfg.ID + "." + v.ID + "." + key
				if metr, ok := m.OidSnmpMap[evalkey]; ok {
					m.Debugf("OK Evaluated metric found %s Eval KEY", evalkey)
					metr.Compute(parameters)
					if metr.Valid == true {
						parameters[v.FieldName] = metr.CookedValue
					}
				} else {
					m.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
				}
			}
		}
	}
}

/*
GetSnmpData GetSNMP Data
*/

// SnmpGetData get Snmp data with snmpget
func (m *Measurement) SnmpGetData() (int64, int64, int64, error) {

	now := time.Now()
	var sent int64
	var errs int64
	l := len(m.snmpOids)
	for i := 0; i < l; i += snmp.MaxOids {
		end := i + snmp.MaxOids
		if end > l {
			end = len(m.snmpOids)
		}
		m.Debugf("Getting snmp data from %d to %d", i, end)
		//	log.Printf("DEBUG oids:%+v", m.snmpOids)
		//	log.Printf("DEBUG oidmap:%+v", m.OidSnmpMap)
		pkt, err := m.snmpClient.Get(m.snmpOids[i:end])
		if err != nil {
			m.Debugf("selected OIDS %+v", m.snmpOids[i:end])
			m.Errorf("SNMP (%s) for OIDs (%d/%d) get error: %s\n", m.snmpClient.Target, i, end, err)
			errs++
			continue
		}

		for _, pdu := range pkt.Variables {
			m.Debugf("DEBUG pdu:%+v", pdu)
			if pdu.Value == nil {
				errs++
				continue
			}
			oid := pdu.Name
			val := pdu.Value
			if metr, ok := m.OidSnmpMap[oid]; ok {
				m.Debugf("OK measurement %s SNMP result OID: %s MetricFound: %+v ", m.cfg.ID, oid, val)
				sent++
				metr.SetRawData(pdu, now)
			} else {
				m.Errorf("OID %s Not Found in measurement %s", oid, m.cfg.ID)
			}
		}
	}

	return int64(l), sent, errs, nil
}

func (m *Measurement) loadIndexedLabels() (map[string]string, error) {

	m.Debugf("Looking up column names %s ", m.cfg.IndexOID)

	allindex := make(map[string]string)

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.Debugf("received SNMP  pdu:%+v", pdu)
		if pdu.Value == nil {
			m.Warnf("no value retured by pdu :%+v", pdu)
			return nil //if error return the bulk process will stop
		}
		if len(pdu.Name) < m.curIdxPos+1 {
			m.Warnf("Received PDU OID smaller  than minimal index(%d) positionretured by pdu :%+v", m.curIdxPos, pdu)
			return nil //if error return the bulk process will stop
		}
		//i := strings.LastIndex(pdu.Name, ".")
		suffix := pdu.Name[m.curIdxPos+1:]

		if m.cfg.IndexAsValue == true {
			allindex[suffix] = suffix
			return nil
		}
		var name string
		switch pdu.Type {
		case gosnmp.OctetString:
			name = string(pdu.Value.([]byte))
			m.Debugf("Got the following OctetString index for [%s/%s]", suffix, name)
		case gosnmp.Counter32, gosnmp.Counter64, gosnmp.Gauge32, gosnmp.Uinteger32:
			name = strconv.FormatUint(snmp.PduVal2UInt64(pdu), 10)
			m.Debugf("Got the following Numeric index for [%s/%s]", suffix, name)
		case gosnmp.Integer:
			name = strconv.FormatInt(snmp.PduVal2Int64(pdu), 10)
			m.Debugf("Got the following Numeric index for [%s/%s]", suffix, name)
		default:
			m.Errorf("Error in IndexedLabel  IndexLabel %s ERR: Not String or numeric Value", m.cfg.IndexOID)
		}
		allindex[suffix] = name
		return nil
	}
	//needed to get data for different indexes
	m.curIdxPos = m.idxPosInOID
	err := m.Walk(m.cfg.IndexOID, setRawData)
	if err != nil {
		m.Errorf("LOADINDEXEDLABELS - SNMP WALK error: %s", err)
		return allindex, err
	}
	if m.cfg.GetMode != "indexed_it" {
		for k, v := range allindex {
			allindex[k] = formatTag(m.log, m.cfg.IndexTagFormat, map[string]string{"IDX1": k, "VAL1": v}, "VAL1")
		}
		return allindex, nil
	}
	// INDIRECT INDEXED
	//backup old index
	allindexOrigin := make(map[string]string, len(allindex))
	for k, v := range allindex {
		allindexOrigin[k] = v
	}

	//initialize allindex again
	allindex = make(map[string]string)
	m.curIdxPos = m.idx2PosInOID
	err = m.Walk(m.cfg.TagOID, setRawData)
	if err != nil {
		m.Errorf("SNMP WALK over IndexOID error: %s", err)
		return allindex, err
	}

	//At this point we have Indirect indexes on allindex_origin and values on allindex
	// Example:
	// allindexOrigin["1"]="9008"
	//    key1="1"
	//    val1="9008"
	// allindex["9008"]="eth0"
	//    key2="9008"
	//    val2="eth0"
	m.Debugf("ORIGINAL INDEX: %+v", allindexOrigin)
	m.Debugf("INDIRECT  INDEX : %+v", allindex)

	allindexIt := make(map[string]string)
	for key1, val1 := range allindexOrigin {
		if val2, ok := allindex[val1]; ok {
			allindexIt[key1] = formatTag(m.log, m.cfg.IndexTagFormat, map[string]string{"IDX1": key1, "VAL1": val1, "IDX2": val1, "VAL2": val2}, "VAL2")
		} else {
			m.Warnf("There is not valid index : %s on TagOID : %s", val1, m.cfg.TagOID)
		}
	}
	//-----------------------------------

	if len(allindexOrigin) != len(allindexIt) {
		m.Warnf("Not all indexes have been indirected\n First Idx [%+v]\n Tagged Idx [ %+v]", allindexOrigin, allindexIt)
	}
	return allindexIt, nil
}
