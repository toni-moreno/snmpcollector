package measurement

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
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
	TagName          []string
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
	MultiIndexMeas   []*Measurement
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

	// If GetMode is multiindex, we can reuse a simple measurement to init each one and get all funcions...?
	switch {
	case m.snmpClient.Version == gosnmp.Version1 || m.DisableBulk:
		m.Walk = m.snmpClient.Walk
	default:
		m.Walk = m.snmpClient.BulkWalk
	}

	if m.cfg.GetMode == "indexed_multiple" {
		// Create, init and store the var into base measurement
		err := m.InitMultiIndex()
		if err != nil {
			return err
		}
		// Load all dependencies and load base measurement
		err = m.LoadMultiIndex()
		if err != nil {
			return err
		}
		return nil
	}

	//loading all posible values in 	m.AllIndexedLabels
	if m.cfg.GetMode == "indexed" || m.cfg.GetMode == "indexed_it" || m.cfg.GetMode == "indexed_mit" {
		m.idxPosInOID = len(m.cfg.IndexOID)
		m.TagName = append([]string{}, m.cfg.IndexTag)
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

// InitMultiIndex initializes measurements from MultiIndexCfg
func (m *Measurement) InitMultiIndex() error {

	// Create an array of measurements, based on length of indexed_multiple:

	multimeas := []*Measurement{}

	// Go over all defined measuremens in MultiIndex...
	for _, v := range m.cfg.MultiIndexCfg {
		// Create a new measurement cfg from multiindex fields...
		mcfg := config.MeasurementCfg{
			ID:             m.ID + ".." + v.Label,
			Name:           v.Label,
			GetMode:        v.GetMode,
			IndexOID:       v.IndexOID,
			TagOID:         v.TagOID,
			MultiTagOID:    v.MultiTagOID,
			IndexTag:       v.IndexTag,
			IndexTagFormat: v.IndexTagFormat,
			Description:    v.Description,
			Fields:         m.cfg.Fields,
			FieldMetric:    m.cfg.FieldMetric,
		}

		//create entirely new measurement based on provided CFG
		mm, err := New(&mcfg, m.log, m.snmpClient, m.DisableBulk)
		if err != nil {
			return err
		}

		//append it with order
		multimeas = append(multimeas, mm)
	}

	// Save it in memory...
	m.MultiIndexMeas = multimeas
	return nil
}

// AddMultiFilter - initializes filter for each existin measurement defined in MultiIndexMeas
func (m *Measurement) AddMultiFilter() {
	for _, meas := range m.MultiIndexMeas {
		meas.AddFilter(meas.FilterCfg, false)
	}
}

// UpdateMultiFilter - updates filter for each existin measurement defined in MultiIndexMeas
func (m *Measurement) UpdateMultiFilter() {
	for _, meas := range m.MultiIndexMeas {
		meas.UpdateFilter()
	}
}

// BuildMultiIndexLabels - builds the multi index labels. Returns the CurIndexedLabels and TagName from processed result
func (m *Measurement) BuildMultiIndexLabels() (map[string]string, []string, error) {

	// Declare array of MultiIndexFormat
	allindex := MultiIndexFormatArray{}

	// Start deep copy of defined indexed measurements
	for i, mindex := range m.MultiIndexMeas {
		ci := make(map[string]string)
		// as CurIndexedLabels is a map, need to make a safe copy of its value in new map
		for i, k := range mindex.CurIndexedLabels {
			ci[i] = k
		}
		iformat := &MultiIndexFormat{
			CurIndexedLabels: ci,
			TagName:          mindex.TagName,
			Index:            i,
			DepDesc:          m.cfg.MultiIndexCfg[i].Dependency,
			Label:            mindex.ID,
		}
		if len(iformat.DepDesc) > 0 {
			err := iformat.GetDepMultiParams()
			if err != nil {
				return nil, nil, err
			}
		}
		allindex = append(allindex, iformat)
	}

	sort.Sort(MultiIndexFormatArray(allindex))

	m.Debugf("Starting to process %d multiindex", len(allindex))

	// Process dependencies | si: sort indexed | index: index info
	for si, index := range allindex {
		// Load on index Dependency all dependency params
		m.Debugf("[%s] - GOT INDEX MULTIPARAMS --> %+v", index.Label, index.Dependency)

		if index.Dependency != nil {
			if index.Dependency.Index > len(allindex)-1 {
				return nil, nil, fmt.Errorf("[%s] - Dependency is out of index range - %d [len: %d], read from %s", m.ID, index.Dependency.Index, len(allindex)-1, index.Label)
			}
			//Read the real id from dependency one, as it is orderered, it is guaranteed that it has already been processed
			ri, err := allindex.GetDepIndex(index.Dependency.Index)
			if err != nil {
				return nil, nil, err
			}

			//Check if the index is itself, just skip it
			if si == ri {
				m.Warnf("[%s] - Detected same IDX on dependency index, skipping it", index.Label)
				continue
			}

			// Set CurIndexedLabels, ci: current index, cv: current value
			for ci, cv := range index.CurIndexedLabels {
				section := ci
				if index.Dependency.Start != -1 {
					section, err = sectionDotSlice(ci, index.Dependency.Start, index.Dependency.End)
					if err != nil {
						m.Warnf("[%] - SectionDotSlice Error: %s", index.Label, err)
						return nil, nil, err
					}
				}
				// check if section is found and split results by '|', dv : dependency value
				if dv, ok := allindex[ri].CurIndexedLabels[section]; ok {
					index.CurIndexedLabels[ci] = dv + "|" + cv
				} else {
					fill := ""
					fillt := ""
					switch index.Dependency.Strategy[1] {
					case "SKIP":
						delete(index.CurIndexedLabels, ci)
					case "FILL":
						if len(index.Dependency.Strategy) == 3 {
							fillt = index.Dependency.Strategy[2]
						}
						fallthrough
					default:
						//fill need to be the number of retrieved tagNames
						for i := 0; i < len(allindex[ri].TagName); i++ {
							fill += fillt + "|"
						}
						index.CurIndexedLabels[ci] = fill + cv
					}
				}
			}

			// Set TagName
			index.TagName = append(allindex[ri].TagName, index.TagName...)

		} else {
			m.Debugf("[%s] - has no dependency", index.Label)
		}
	}

	// As the result is built with index, we can use a simply split(.) to mantain an array order
	// Examples:
	// s = strings.Split(m.cfg.MultiIndexResult ".")
	// s = []string{"1", "2", "IDX{0}", "45", "IDX{1}", "4"}
	// s = []string{"IDX{0}", "4", "IDX{1}", "5", "6"}
	// s = []string{"IDX{0}"}

	s := strings.Split(m.cfg.MultiIndexResult, ".")

	resindex, suffix, err := BuildParseResults(allindex, s)
	if err != nil {
		return nil, nil, err
	}

	// Start process of merge all generated IDX...
	// Go over all indexes, start from the first and go over merging it
	// Values will be stored as MultiIndexFormat, as it includes all indexes and tags...

	// Exampple:
	// [(1.1) = "TAG1A", (1.2) = "TAG1B"]
	// [(2.1) = "TAG2A", (2.2) = "TAG2B"]

	//RESULT:
	// [(1.1.2.1)="TAG1A|TAG2A"]
	// [(1.1.2.2)="TAG1A|TAG2B"]
	// [(1.2.2.1)="TAG1B|TAG2A"]
	// [(1.2.2.2)="TAG1B|TAG2B"]

	// Exampple:
	// [(1.1) = "TAG1A", (1.2) = "TAG1B"]
	// [(2.1) = "TAG2A", (2.2) = "TAG2B"]
	// [(3.1) = "TAG3A", (3.2) = "TAG3B"]

	//RESULT:
	// [(1.1.2.1.3.1)="TAG1A|TAG2A|TAG3A"]
	// [(1.1.2.1.3.1)="TAG1A|TAG2A|TAG3B"]
	// [(1.1.2.2.3.1)="TAG1A|TAG2B|TAG3A"]
	// [(1.1.2.2.3.2)="TAG1A|TAG2B|TAG3B"]
	// [(1.2.2.1.3.1)="TAG1B|TAG2A|TAG3A"]
	// [(1.2.2.1.3.2)="TAG1B|TAG2A|TAG3B"]
	// [(1.2.2.1.3.1)="TAG1B|TAG2B|TAG3A"]
	// [(1.2.2.1.3.1)="TAG1B|TAG2B|TAG3B"]

	// Merge process...
	fresult, errmerge := MergeResults(resindex, suffix)
	if errmerge != nil {
		return nil, nil, errmerge
	}
	m.Debugf("Got final multiindex result index labels: %+v", fresult.CurIndexedLabels)
	m.Debugf("Got final multiindex result tag names: %+v", fresult.TagName)

	return fresult.CurIndexedLabels, fresult.TagName, nil
}

// LoadMultiIndex loads the multiindex with all attached measurements
func (m *Measurement) LoadMultiIndex() error {

	// Load MultiIndex labels based on dependencies
	mil, tag, err := m.BuildMultiIndexLabels()
	if err != nil {
		return err
	}

	// Finally, set up base measurement
	m.TagName = tag
	m.AllIndexedLabels = mil
	m.CurIndexedLabels = mil
	m.MetricTable = NewMetricTable(m.cfg, m.log, mil)

	m.InitBuildRuntime()
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

// CheckInitFilter loads measurement filter on measurement if name/label is matched
func (m *Measurement) CheckInitFilter(f *config.MeasFilterCfg) (bool, bool) {
	// check if filter must be applied on base measurement
	if m.cfg.GetMode != "value" && f.IDMeasurementCfg == m.ID {
		m.FilterCfg = f
		return true, false
	}
	// check if filter must be applied on multi index
	if m.cfg.GetMode == "indexed_multiple" {
		for _, mi := range m.MultiIndexMeas {
			// to be unique, multiple filter is defined as base<ID>..index<ID>
			if f.IDMeasurementCfg == mi.ID {
				mi.FilterCfg = f
				return true, true
			}
		}
	}
	return false, false
}

// AddFilter attach a filtering process to the measurement, but it is not initialized
func (m *Measurement) AddFilter(f *config.MeasFilterCfg, multi bool) error {
	var err error
	if m.cfg.GetMode == "value" {
		return fmt.Errorf("Error this measurement %s  is not indexed(snmptable) not Filter apply ", m.cfg.ID)
	}

	// If multi, all multi indexes must be reloaded and reload main measurement
	if multi {
		m.AddMultiFilter()
		err := m.LoadMultiIndex()
		if err != nil {
			return err
		}
	}

	// If multi and the current measurement doesn't have any filter to be applied, just skip it
	if m.FilterCfg == nil {
		if multi {
			m.Infof("No filter on base base measurement, skipping %s", m.cfg.ID)
			return err
		}
		return fmt.Errorf("Error invalid  NIL  filter on measurment %s ", m.cfg.ID)
	}

	// Filter is already set on check and init filter
	// m.FilterCfg = f

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

	if m.cfg.GetMode == "value" {
		return false, fmt.Errorf("Error this measurement %s  is not indexed(snmptable) not Filter apply ", m.cfg.ID)
	}

	//fist update  all indexed--------
	m.Infof("Re Loading Indexed values")

	// if its indexed_multiple, we need to update internal filters and create the new metric table on based one
	if m.cfg.GetMode == "indexed_multiple" {
		m.UpdateMultiFilter()
		il, _, err := m.BuildMultiIndexLabels()
		if err != nil {
			return false, err
		}
		m.AllIndexedLabels = il
	} else {
		il, err2 := m.loadIndexedLabels()

		if err2 != nil {
			m.Errorf("Error while trying to reload Indexed Labels for baseOid %s : ERROR: %s", m.cfg.IndexOID, err)
			return false, err
		}
		m.AllIndexedLabels = il
	}

	// Reload measurement indexes
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
		m.Debugf("DEBUG pdu [%+v] || Value type %T [%x]", pdu, pdu.Value, pdu.Type)
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
			m.Debugf("returned OID from device: %s  Not Found in measurement /metr list: %+v, %v", pdu.Name, m.cfg.ID, m.OidSnmpMap)
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

	//copy the input
	switch m.cfg.GetMode {
	case "value":
		parameters := make(map[string]interface{})
		//copy of the catalog map
		for k, v := range catalog {
			parameters[k] = v
		}

		m.Debugf("Building parrameters array for index measurement %s", m.cfg.ID)
		parameters["NFR"] = len(m.AllIndexedLabels)                          //Number of non filtered rows
		parameters["NR"] = len(m.CurIndexedLabels)                           //Number of current rows (like awk) --after filtered applied  --
		parameters["NF"] = len(m.cfg.FieldMetric) + len(m.cfg.OidCondMetric) //Number of fields ( like awk)
		//getting all values to the array
		for _, v := range m.cfg.FieldMetric {
			if metr, ok := m.OidSnmpMap[v.BaseOID]; ok {
				metr.GetEvaluableVariables(parameters)
			} else {
				m.Debugf("Evaluated metric not Found for Eval key %s", v.BaseOID)
			}
		}
		for _, v := range m.cfg.OidCondMetric {
			RealOID := m.cfg.ID + "." + v.ID
			if metr, ok := m.OidSnmpMap[RealOID]; ok {
				metr.GetEvaluableVariables(parameters)
			} else {
				m.Debugf("Evaluated OIDCondition metric not Found for Eval key %s", RealOID)
			}
		}
		m.Debugf("PARAMETERS: %+v", parameters)
		//compute Evalutated metrics
		for _, v := range m.cfg.EvalMetric {
			evalkey := m.cfg.ID + "." + v.ID
			if metr, ok := m.OidSnmpMap[evalkey]; ok {
				m.Debugf("OK Evaluated metric found %s Eval KEY", evalkey)
				metr.Compute(parameters)
				metr.GetEvaluableVariables(parameters)
			} else {
				m.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
			}
		}
	case "indexed", "indexed_it", "indexed_mit", "indexed_multiple":
		for key, val := range m.CurIndexedLabels {
			parameters := make(map[string]interface{})
			//copy of the catalog map
			for k, v := range catalog {
				parameters[k] = v
			}
			//building parameters array
			m.Debugf("Building parrameters array for index %s/%s", key, val)
			parameters["NFR"] = len(m.AllIndexedLabels) //Number of non filtered rows
			parameters["NR"] = len(m.CurIndexedLabels)  //Number of rows (like awk)
			parameters["NF"] = len(m.cfg.FieldMetric)   //Number of fields ( like awk)
			//TODO: add other common variables => Elapsed , etc
			//getting all values to the array
			for _, v := range m.cfg.FieldMetric {
				if metr, ok := m.OidSnmpMap[v.BaseOID+"."+key]; ok {
					m.Debugf("OK Field metric found %s with FieldName %s", metr.GetID(), metr.GetFieldName())
					metr.GetEvaluableVariables(parameters)
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
					metr.GetEvaluableVariables(parameters)
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
			m.Debugf("DEBUG pdu [%+v] || Value type %T [%x] ", pdu, pdu.Value, pdu.Type)
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
		name := "ErrorOnGetIdxValue"
		switch pdu.Type {
		case gosnmp.OctetString:
			name = pdu.Value.(string)
			m.Debugf("Got the following OctetString index for [%s/%s]", suffix, name)
		case gosnmp.Counter32, gosnmp.Counter64, gosnmp.Gauge32, gosnmp.Uinteger32:
			name = strconv.FormatUint(snmp.PduVal2UInt64(pdu), 10)
			m.Debugf("Got the following Numeric index for [%s/%s]", suffix, name)
		case gosnmp.Integer:
			name = strconv.FormatInt(snmp.PduVal2Int64(pdu), 10)
			m.Debugf("Got the following Numeric index for [%s/%s]", suffix, name)
		case gosnmp.IPAddress:
			var err error
			name, err = snmp.PduVal2IPaddr(pdu)
			m.Debugf("Got the following IPaddress index for [%s/%s]", suffix, name)
			if err != nil {
				m.Errorf("Error on  IndexedLabel  IPAddress  to string conversion: %s", err)
			}
		default:
			m.Errorf("Error in IndexedLabel  IndexLabel %s ERR: Not String or numeric or IPAddress Value", m.cfg.IndexOID)
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
	if m.cfg.GetMode == "indexed" {
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

	switch m.cfg.GetMode {
	case "indexed_it":
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

		if len(allindexOrigin) != len(allindexIt) {
			m.Warnf("Not all indexes have been indirected\n First Idx [%+v]\n Tagged Idx [ %+v]", allindexOrigin, allindexIt)
		}
		return allindexIt, nil

	case "indexed_mit":
		//Make another copy of origin, we need to mantain always a base origin index
		allindexRes := make(map[string]string, len(allindexOrigin))
		for k, v := range allindexOrigin {
			allindexRes[k] = v
		}

		// Go over all defined multipletagoid
		for k, tagcfg := range m.cfg.MultiTagOID {
			//initialize all index again
			allindex = make(map[string]string)
			//Store the last position to use it on allindex
			m.curIdxPos = len(tagcfg.TagOID)
			err = m.Walk(tagcfg.TagOID, setRawData)
			if err != nil {
				m.Errorf("SNMP WALK over IndexOID error: %s", err)
				return allindex, err
			}
			// Create a base map, we need it to mantain the key1 as the first index, so we can reference it back to as the core index
			allindexIt := make(map[string]string)
			// Start logic of match
			for key1, val1 := range allindexRes {
				// As we only need the value to keep going on the different tables, we can create a custom check field if the index is not the same as original
				// It is used on qos to get parent cfg oids
				check := formatTag(m.log, tagcfg.IndexFormat, map[string]string{"IDX1": key1, "VAL1": val1}, "VAL1")
				if val2, ok := allindex[check]; ok {
					//Only apply formatTag based on the last index...
					if k == len(m.cfg.MultiTagOID)-1 {
						allindexIt[key1] = formatTag(m.log, m.cfg.IndexTagFormat, map[string]string{"IDX1": key1, "VAL1": val1, "IDX2": val1, "VAL2": val2}, "VAL2")
						continue
					}
					allindexIt[key1] = val2
				} else {
					// Set debug due to large logs generated. This case is generic, so a not match can be normal
					m.Debugf("[%d] - There is not valid index : %s on TagOID : %s", k, val1, tagcfg.TagOID)
				}
			}
			allindexRes = allindexIt
		}

		if len(allindexOrigin) != len(allindexRes) {
			m.Warnf("Not all indexes have been indirected\n First Idx [%+v]\n Tagged Idx [ %+v]", allindexOrigin, allindexRes)
		}

		return allindexRes, nil

	default:
		return allindex, fmt.Errorf("Uknown provided getmode %s on measurement %s", m.cfg.GetMode, m.ID)
	}
}
