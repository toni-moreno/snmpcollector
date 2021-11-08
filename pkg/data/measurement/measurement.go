package measurement

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/agent/bus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/filter"
	"github.com/toni-moreno/snmpcollector/pkg/data/metric"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/stats"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

var (
	confDir string              // Needed to get File Filters data
	dbc     *config.DatabaseCfg // Needed to get Custom Filter  data
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

// Measurement the runtime measurement config
type Measurement struct {
	cfg          *config.MeasurementCfg
	measFilters  []string
	mFilters     map[string]*config.MeasFilterCfg
	snmpOids     []string
	idxPosInOID  int
	idx2PosInOID int
	curIdxPos    int // used in Walk functions could be variable depending on the Index (or IndexTag)
	// rtData protect this struct between reads from UI (ToJSON) and writes from GatherLoop
	rtData sync.RWMutex
	// initialized is used to determine if gather could be done
	initialized bool

	ID      string
	MName   string
	TagName []string
	// MetricTable data from OidSnmpMap structured to be passed to the UI (with ToJSON).
	// We use pointers, so the data is the same here and OidSnmpMap.
	MetricTable *metric.MetricTable
	// OidSnmpMap store values returned from the snmp queries
	OidSnmpMap       map[string]*metric.SnmpMetric `json:"-"` // snmpMetric mapped with real OID's
	AllIndexedLabels map[string]string             //`json:"-"` //all available values on the remote device
	CurIndexedLabels map[string]string             //`json:"-"`
	FilterCfg        *config.MeasFilterCfg
	Filter           filter.Filter
	Log              utils.Logger `json:"-"`
	snmpClient       *snmp.Client
	MultiIndexMeas   []*Measurement
	// Enabled is true if this measurement should gather metrics. This value is controlled by the device
	Active    bool
	Connected bool
	// Measurement statistics
	stats     stats.GatherStats  // Runtime Internal statistic
	Stats     *stats.GatherStats // Public info for thread safe accessing to the data ()
	statsData sync.RWMutex
}

// GetBasicStats get basic info for this measurement
func (m *Measurement) GetBasicStats() *stats.GatherStats {
	m.statsData.RLock()
	defer m.statsData.RUnlock()
	return m.Stats
}

// GetBasicStats get basic info for this device
func (m *Measurement) getBasicStats() *stats.GatherStats {
	stat := m.stats.ThSafeCopy()
	stat.TagMap = m.stats.TagMap
	stat.NumMeasurements = 1
	stat.NumMetrics = len(m.OidSnmpMap)
	return stat
}

func (m *Measurement) SetStats(st stats.GatherStats) {
	// TODO: st contains a lock and its being passed as value. Could create a problem?
	m.stats = st
	m.statsData.Lock()
	m.Stats = m.getBasicStats()
	m.statsData.Unlock()
}

func New(c *config.MeasurementCfg, measFilters []string, mFilters map[string]*config.MeasFilterCfg, active bool, l utils.Logger) *Measurement {
	return &Measurement{
		ID:          c.ID,
		MName:       c.Name,
		cfg:         c,
		measFilters: measFilters,
		mFilters:    mFilters,
		Log:         l,
		Active:      active,
	}
}

// InvalidateMetrics mark as old (Valid=False) all the metrics in the table
func (m *Measurement) InvalidateMetrics() {
	m.MetricTable.InvalidateTable()
}

/*Init does:
 *inicialize AllIndexesLabels
 *Assign CurIndexedLabels to all Labels (until filters set)
 *init MetricTable
 * This function actually connects to the device to gather values
 */
func (m *Measurement) Init() error {
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
		m.InitFilters()
		return nil
	}

	// loading all posible values in 	m.AllIndexedLabels
	if m.cfg.GetMode == "indexed" || m.cfg.GetMode == "indexed_it" || m.cfg.GetMode == "indexed_mit" {
		m.idxPosInOID = len(m.cfg.IndexOID)
		m.TagName = append([]string{}, m.cfg.IndexTag)
		if (m.cfg.GetMode) == "indexed_it" {
			m.idx2PosInOID = len(m.cfg.TagOID)
		}
		m.Log.Infof("Loading Indexed values")
		il, err := m.loadIndexedLabels()
		if err != nil {
			return fmt.Errorf("trying to load Indexed Labels for baseOid %s: %s", m.cfg.IndexOID, err)
		}
		m.AllIndexedLabels = il
		// Final Selected Indexes are All Indexed
		m.CurIndexedLabels = m.AllIndexedLabels
	}

	/********************************
	 * Initialize Metric Runtime data in one array m-values
	 * ******************************/
	m.Log.Debug("Initialize OID measurement per label => map of metric object per field | OID array [ready to send to the walk device] | OID=>Metric MAP")
	m.MetricTable = metric.NewMetricTable(m.cfg, m.Log, m.CurIndexedLabels)

	m.InitFilters()

	m.stats.SetActive(m.Active)
	// init public queriable stats
	m.statsData.Lock()
	m.Stats = m.getBasicStats()
	m.statsData.Unlock()
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

		// create entirely new measurement based on provided CFG
		mm := New(&mcfg, m.measFilters, m.mFilters, m.Active, m.Log)
		mm.SetSNMPClient(*m.snmpClient)
		err := mm.Init()
		if err != nil {
			return fmt.Errorf("init multi measurement %s..%s", m.ID, v.Label)
		}

		// append it with order
		multimeas = append(multimeas, mm)
	}

	// Save it in memory...
	m.MultiIndexMeas = multimeas
	return nil
}

// AddMultiFilter - initializes filter for each existing measurement defined in MultiIndexMeas
func (m *Measurement) AddMultiFilter() {
	for _, meas := range m.MultiIndexMeas {
		meas.AddFilter(meas.FilterCfg, false)
	}
}

// UpdateMultiFilter - updates filter for each existing measurement defined in MultiIndexMeas
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

	m.Log.Debugf("Starting to process %d multiindex", len(allindex))

	// Process dependencies | si: sort indexed | index: index info
	for si, index := range allindex {
		// Load on index Dependency all dependency params
		m.Log.Debugf("[%s] - GOT INDEX MULTIPARAMS --> %+v", index.Label, index.Dependency)

		if index.Dependency != nil {
			if index.Dependency.Index > len(allindex)-1 {
				return nil, nil, fmt.Errorf("[%s] - Dependency is out of index range - %d [len: %d], read from %s", m.ID, index.Dependency.Index, len(allindex)-1, index.Label)
			}
			// Read the real id from dependency one, as it is orderered, it is guaranteed that it has already been processed
			ri, err := allindex.GetDepIndex(index.Dependency.Index)
			if err != nil {
				return nil, nil, err
			}

			// Check if the index is itself, just skip it
			if si == ri {
				m.Log.Warnf("[%s] - Detected same IDX on dependency index, skipping it", index.Label)
				continue
			}

			// Set CurIndexedLabels, ci: current index, cv: current value
			for ci, cv := range index.CurIndexedLabels {
				section := ci
				if index.Dependency.Start != -1 {
					section, err = sectionDotSlice(ci, index.Dependency.Start, index.Dependency.End)
					if err != nil {
						m.Log.Warnf("[%] - SectionDotSlice Error: %s", index.Label, err)
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
						// fill need to be the number of retrieved tagNames
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
			m.Log.Debugf("[%s] - has no dependency", index.Label)
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

	// RESULT:
	// [(1.1.2.1)="TAG1A|TAG2A"]
	// [(1.1.2.2)="TAG1A|TAG2B"]
	// [(1.2.2.1)="TAG1B|TAG2A"]
	// [(1.2.2.2)="TAG1B|TAG2B"]

	// Exampple:
	// [(1.1) = "TAG1A", (1.2) = "TAG1B"]
	// [(2.1) = "TAG2A", (2.2) = "TAG2B"]
	// [(3.1) = "TAG3A", (3.2) = "TAG3B"]

	// RESULT:
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
	m.Log.Debugf("Got final multiindex result index labels: %+v", fresult.CurIndexedLabels)
	m.Log.Debugf("Got final multiindex result tag names: %+v", fresult.TagName)

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
	m.MetricTable = metric.NewMetricTable(m.cfg, m.Log, mil)

	m.InitBuildRuntime()
	return nil
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
			m.Log.Infof("No filter on base base measurement, skipping %s", m.cfg.ID)
			return err
		}
		return fmt.Errorf("Error invalid  NIL  filter on measurment %s ", m.cfg.ID)
	}

	// Filter is already set on check and init filter
	// m.FilterCfg = f

	switch m.FilterCfg.FType {
	case "file":
		m.Filter = filter.NewFileFilter(m.FilterCfg.FilterName, m.FilterCfg.EnableAlias, m.Log)
		err = m.Filter.Init(confDir)
		if err != nil {
			return fmt.Errorf("Error invalid File Filter : %s", err)
		}
	case "OIDCondition":
		cond, err2 := dbc.GetOidConditionCfgByID(m.FilterCfg.FilterName)
		if err2 != nil {
			m.Log.Errorf("Error getting filter id %s OIDCondition [id: %s ] data : %s", m.FilterCfg.ID, m.FilterCfg.FilterName, err)
		}

		if cond.IsMultiple {
			m.Filter = filter.NewOidMultipleFilter(cond.OIDCond, m.Log)
			err = m.Filter.Init(m.snmpClient.Walk, dbc)
			if err != nil {
				return fmt.Errorf("Error invalid Multiple Condition Filter : %s", err)
			}
		} else {
			m.Filter = filter.NewOidFilter(cond.OIDCond, cond.CondType, cond.CondValue, m.Log)
			err = m.Filter.Init(m.snmpClient.Walk)
			if err != nil {
				return fmt.Errorf("Error invalid OID condition Filter : %s", err)
			}
		}

	case "custom":
		m.Filter = filter.NewCustomFilter(m.FilterCfg.FilterName, m.FilterCfg.EnableAlias, m.Log)
		err = m.Filter.Init(dbc)
		if err != nil {
			return fmt.Errorf("Error invalid Custom Filter : %s", err)
		}
	default:
		return fmt.Errorf("Invalid Filter Type %s for measurement: %s", m.FilterCfg.FType, m.cfg.ID)
	}

	err = m.Filter.Update()
	if err != nil {
		m.Log.Errorf("Error while trying to apply file Filter  ERROR: %s", err)
	}

	// now we have the 	m.Filterlabels array initialized with only those values which we will need
	// Loading final Values to query with snmp
	m.CurIndexedLabels = m.Filter.MapLabels(m.AllIndexedLabels)
	m.MetricTable = metric.NewMetricTable(m.cfg, m.Log, m.CurIndexedLabels)
	return err
}

// UpdateFilter reload indexed with filters
func (m *Measurement) UpdateFilter() (bool, error) {
	var err error

	if m.cfg.GetMode == "value" {
		return false, fmt.Errorf("Error this measurement %s  is not indexed(snmptable) not Filter apply ", m.cfg.ID)
	}

	// fist update  all indexed--------
	m.Log.Infof("Re Loading Indexed values")

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
			m.Log.Errorf("Error while trying to reload Indexed Labels for baseOid %s : ERROR: %s", m.cfg.IndexOID, err2)
			return false, err2
		}
		m.AllIndexedLabels = il
	}

	// Reload measurement indexes
	if m.Filter == nil {
		m.Log.Debugf("There is no filter configured in this measurement %s", m.cfg.ID)
		// check if curindexed different of AllIndexed
		delIndexes := utils.DiffKeyValuesInMap(m.CurIndexedLabels, m.AllIndexedLabels)
		newIndexes := utils.DiffKeyValuesInMap(m.AllIndexedLabels, m.CurIndexedLabels)

		if len(newIndexes) == 0 && len(delIndexes) == 0 {
			// no changes on the Filter
			m.Log.Infof("No changes found on the Index for this measurement")
			return false, nil
		}
		m.CurIndexedLabels = m.AllIndexedLabels

		m.Log.Debugf("NEW INDEXES: %+v", newIndexes)
		m.Log.Debugf("DELETED INDEXES: %+v", delIndexes)

		if len(delIndexes) > 0 {
			m.MetricTable.Pop(delIndexes)
		}
		if len(newIndexes) > 0 {
			m.MetricTable.Push(newIndexes)
		}
		return true, nil
	}
	//----------------
	m.Log.Infof("Applying filter : [ %s ] on measurement", m.FilterCfg.ID)

	err = m.Filter.Update()
	if err != nil {
		m.Log.Errorf("Error while trying to apply Filter : ERROR: %s", err)
		return false, err
	}
	// check if all values have been filtered to send a warnign message.
	if m.Filter.Count() == 0 {
		m.Log.Warnf("WARNING after applying filter no values on this measurement will be sent")
	}
	// check if newfilterlabels are different than previous.

	// now we have the 	m.Filter,m.ls array initialized with only those values which we will need
	// Loading final Values to query with snmp
	newIndexedLabels := m.Filter.MapLabels(m.AllIndexedLabels)

	delIndexes := utils.DiffKeyValuesInMap(m.CurIndexedLabels, newIndexedLabels)
	newIndexes := utils.DiffKeyValuesInMap(newIndexedLabels, m.CurIndexedLabels)

	if len(newIndexes) == 0 && len(delIndexes) == 0 {
		// no changes on the Filter
		m.Log.Infof("No changes on the filter %s ", m.FilterCfg.FType)
		return false, nil
	}

	m.Log.Debugf("NEW INDEXES: %+v", newIndexes)
	m.Log.Debugf("DELETED INDEXES: %+v", delIndexes)

	m.CurIndexedLabels = newIndexedLabels

	if len(delIndexes) > 0 {
		m.MetricTable.Pop(delIndexes)
	}
	if len(newIndexes) > 0 {
		m.MetricTable.Push(newIndexes)
	}

	return true, nil
}

// GetData read data from device using SNMP get (GetMode=value) or walk (default).
func (m *Measurement) GetData() (int64, int64, int64) {
	now := time.Now()
	var gathered int64
	var processed int64
	var errors int64

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.Log.Debugf("DEBUG pdu [%+v] || Value type %T [%x]", pdu, pdu.Value, pdu.Type)
		gathered++
		if pdu.Value == nil {
			m.Log.Warnf("no value retured by pdu :%+v", pdu)
			errors++
			return nil // if error return the bulk process will stop
		}
		if metr, ok := m.OidSnmpMap[pdu.Name]; ok {
			m.Log.Debugf("OK measurement %s SNMP RESULT OID %s MetricFound", pdu.Name, pdu.Value)
			processed++
			metr.SetRawData(pdu, now)
		} else {
			m.Log.Debugf("returned OID from device: %s  Not Found in measurement metric list: %+v, %v", pdu.Name, m.cfg.ID, m.OidSnmpMap)
		}
		return nil
	}

	if m.cfg.GetMode == "value" {
		// never will be error
		m.snmpClient.Get(m.snmpOids, setRawData)
	} else {
		for _, v := range m.cfg.FieldMetric {
			if err := m.snmpClient.Walk(v.BaseOID, setRawData); err != nil {
				m.Log.Errorf("SNMP WALK for OID (%s) get error: %s", v.BaseOID, err)
				errors += int64(m.MetricTable.Len())
			}
		}
	}

	return gathered, processed, errors
}

// ComputeOidConditionalMetrics take OID contitional metrics and computes true value
func (m *Measurement) ComputeOidConditionalMetrics() {
	if m.cfg.OidCondMetric == nil {
		m.Log.Infof("Not Oid CONDITIONEVAL metrics exist on this measurement")
		return
	}
	switch m.cfg.GetMode {
	case "value":
		// compute Evalutated metrics
		for _, v := range m.cfg.OidCondMetric {
			evalkey := m.cfg.ID + "." + v.ID
			if metr, ok := m.OidSnmpMap[evalkey]; ok {
				m.Log.Debugf("OK OidCondition  metric found %s Eval KEY", evalkey)
				metr.Compute(m.snmpClient.Walk, dbc)
			} else {
				m.Log.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
			}
		}
	default:
		m.Log.Warnf("Warning there is CONDITIONAL metrics on indexed measurements!!")
	}
}

// ComputeEvaluatedMetrics take evaluated metrics and computes them from the other values
func (m *Measurement) ComputeEvaluatedMetrics(catalog map[string]interface{}) {
	if m.cfg.EvalMetric == nil {
		m.Log.Infof("Not EVAL metrics exist on  this measurement")
		return
	}

	// copy the input
	switch m.cfg.GetMode {
	case "value":
		parameters := make(map[string]interface{})
		// copy of the catalog map
		for k, v := range catalog {
			parameters[k] = v
		}

		m.Log.Debugf("Building parrameters array for index measurement %s", m.cfg.ID)
		parameters["NFR"] = len(m.AllIndexedLabels)                          // Number of non filtered rows
		parameters["NR"] = len(m.CurIndexedLabels)                           // Number of current rows (like awk) --after filtered applied  --
		parameters["NF"] = len(m.cfg.FieldMetric) + len(m.cfg.OidCondMetric) // Number of fields ( like awk)
		// getting all values to the array
		for _, v := range m.cfg.FieldMetric {
			if metr, ok := m.OidSnmpMap[v.BaseOID]; ok {
				metr.GetEvaluableVariables(parameters)
			} else {
				m.Log.Debugf("Evaluated metric not Found for Eval key %s", v.BaseOID)
			}
		}
		for _, v := range m.cfg.OidCondMetric {
			RealOID := m.cfg.ID + "." + v.ID
			if metr, ok := m.OidSnmpMap[RealOID]; ok {
				metr.GetEvaluableVariables(parameters)
			} else {
				m.Log.Debugf("Evaluated OIDCondition metric not Found for Eval key %s", RealOID)
			}
		}
		m.Log.Debugf("PARAMETERS: %+v", parameters)
		// compute Evalutated metrics
		for _, v := range m.cfg.EvalMetric {
			evalkey := m.cfg.ID + "." + v.ID
			if metr, ok := m.OidSnmpMap[evalkey]; ok {
				m.Log.Debugf("OK Evaluated metric found %s Eval KEY", evalkey)
				metr.Compute(parameters)
				metr.GetEvaluableVariables(parameters)
			} else {
				m.Log.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
			}
		}
	case "indexed", "indexed_it", "indexed_mit", "indexed_multiple":
		for key, val := range m.CurIndexedLabels {
			parameters := make(map[string]interface{})
			// copy of the catalog map
			for k, v := range catalog {
				parameters[k] = v
			}
			// building parameters array
			m.Log.Debugf("Building parrameters array for index %s/%s", key, val)
			parameters["NFR"] = len(m.AllIndexedLabels) // Number of non filtered rows
			parameters["NR"] = len(m.CurIndexedLabels)  // Number of rows (like awk)
			parameters["NF"] = len(m.cfg.FieldMetric)   // Number of fields ( like awk)
			// TODO: add other common variables => Elapsed , etc
			// getting all values to the array
			for _, v := range m.cfg.FieldMetric {
				if metr, ok := m.OidSnmpMap[v.BaseOID+"."+key]; ok {
					m.Log.Debugf("OK Field metric found %s with FieldName %s", metr.GetID(), metr.GetFieldName())
					metr.GetEvaluableVariables(parameters)
				} else {
					m.Log.Debugf("Evaluated metric not Found for Eval key %s")
				}
			}
			m.Log.Debugf("PARAMETERS: %+v", parameters)
			// compute Evalutated metrics
			for _, v := range m.cfg.EvalMetric {
				evalkey := m.cfg.ID + "." + v.ID + "." + key
				if metr, ok := m.OidSnmpMap[evalkey]; ok {
					m.Log.Debugf("OK Evaluated metric found %s Eval KEY", evalkey)
					metr.Compute(parameters)
					metr.GetEvaluableVariables(parameters)
				} else {
					m.Log.Debugf("Evaluated metric not Found for Eval key %s", evalkey)
				}
			}
		}
	}
}

func (m *Measurement) loadIndexedLabels() (map[string]string, error) {
	m.Log.Debugf("Looking up column names %s ", m.cfg.IndexOID)

	allindex := make(map[string]string)

	setRawData := func(pdu gosnmp.SnmpPDU) error {
		m.Log.Debugf("received SNMP  pdu:%+v", pdu)
		if pdu.Value == nil {
			m.Log.Warnf("no value retured by pdu :%+v", pdu)
			return nil // if error return the bulk process will stop
		}
		if len(pdu.Name) < m.curIdxPos+1 {
			m.Log.Warnf("Received PDU OID smaller  than minimal index(%d) position returned by pdu :%+v", m.curIdxPos, pdu)
			return nil // if error return the bulk process will stop
		}
		// i := strings.LastIndex(pdu.Name, ".")
		suffix := pdu.Name[m.curIdxPos+1:]

		if m.cfg.IndexAsValue == true {
			allindex[suffix] = suffix
			return nil
		}
		name := "ErrorOnGetIdxValue"
		switch pdu.Type {
		case gosnmp.OctetString:
			name = string(pdu.Value.([]byte))
			m.Log.Debugf("Got the following OctetString index for [%s/%s]", suffix, name)
		case gosnmp.Counter32, gosnmp.Counter64, gosnmp.Gauge32, gosnmp.Uinteger32:
			name = strconv.FormatUint(snmp.PduVal2UInt64(pdu), 10)
			m.Log.Debugf("Got the following Numeric index for [%s/%s]", suffix, name)
		case gosnmp.Integer:
			name = strconv.FormatInt(snmp.PduVal2Int64(pdu), 10)
			m.Log.Debugf("Got the following Numeric index for [%s/%s]", suffix, name)
		case gosnmp.IPAddress:
			var err error
			name, err = snmp.PduVal2IPaddr(pdu)
			m.Log.Debugf("Got the following IPaddress index for [%s/%s]", suffix, name)
			if err != nil {
				m.Log.Errorf("Error on  IndexedLabel  IPAddress  to string conversion: %s", err)
			}
		default:
			m.Log.Errorf("Error in IndexedLabel  IndexLabel %s ERR: Not String or numeric or IPAddress Value", m.cfg.IndexOID)
		}
		allindex[suffix] = name
		return nil
	}
	// needed to get data for different indexes
	m.curIdxPos = m.idxPosInOID
	err := m.snmpClient.Walk(m.cfg.IndexOID, setRawData)
	if err != nil {
		m.Log.Errorf("LOADINDEXEDLABELS - SNMP WALK error: %s", err)
		return allindex, err
	}
	if m.cfg.GetMode == "indexed" {
		for k, v := range allindex {
			allindex[k] = formatTag(m.Log, m.cfg.IndexTagFormat, map[string]string{"IDX1": k, "VAL1": v}, "VAL1")
		}
		return allindex, nil
	}
	// INDIRECT INDEXED
	// backup old index
	allindexOrigin := make(map[string]string, len(allindex))
	for k, v := range allindex {
		allindexOrigin[k] = v
	}

	switch m.cfg.GetMode {
	case "indexed_it":
		// initialize allindex again
		allindex = make(map[string]string)
		m.curIdxPos = m.idx2PosInOID
		err = m.snmpClient.Walk(m.cfg.TagOID, setRawData)
		if err != nil {
			m.Log.Errorf("SNMP WALK over IndexOID error: %s", err)
			return allindex, err
		}

		// At this point we have Indirect indexes on allindex_origin and values on allindex
		// Example:
		// allindexOrigin["1"]="9008"
		//    key1="1"
		//    val1="9008"
		// allindex["9008"]="eth0"
		//    key2="9008"
		//    val2="eth0"
		m.Log.Debugf("ORIGINAL INDEX: %+v", allindexOrigin)
		m.Log.Debugf("INDIRECT  INDEX : %+v", allindex)

		allindexIt := make(map[string]string)
		for key1, val1 := range allindexOrigin {
			if val2, ok := allindex[val1]; ok {
				allindexIt[key1] = formatTag(m.Log, m.cfg.IndexTagFormat, map[string]string{"IDX1": key1, "VAL1": val1, "IDX2": val1, "VAL2": val2}, "VAL2")
			} else {
				m.Log.Warnf("There is not valid index : %s on TagOID : %s", val1, m.cfg.TagOID)
			}
		}

		if len(allindexOrigin) != len(allindexIt) {
			m.Log.Warnf("Not all indexes have been indirected\n First Idx [%+v]\n Tagged Idx [ %+v]", allindexOrigin, allindexIt)
		}
		return allindexIt, nil

	case "indexed_mit":
		// Make another copy of origin, we need to mantain always a base origin index
		allindexRes := make(map[string]string, len(allindexOrigin))
		for k, v := range allindexOrigin {
			allindexRes[k] = v
		}

		// Go over all defined multipletagoid
		for k, tagcfg := range m.cfg.MultiTagOID {
			// initialize all index again
			allindex = make(map[string]string)
			// Store the last position to use it on allindex
			m.curIdxPos = len(tagcfg.TagOID)
			err = m.snmpClient.Walk(tagcfg.TagOID, setRawData)
			if err != nil {
				m.Log.Errorf("SNMP WALK over IndexOID error: %s", err)
				return allindex, err
			}
			// Create a base map, we need it to mantain the key1 as the first index, so we can reference it back to as the core index
			allindexIt := make(map[string]string)
			// Start logic of match
			for key1, val1 := range allindexRes {
				// As we only need the value to keep going on the different tables, we can create a custom check field if the index is not the same as original
				// It is used on qos to get parent cfg oids
				check := formatTag(m.Log, tagcfg.IndexFormat, map[string]string{"IDX1": key1, "VAL1": val1}, "VAL1")
				if val2, ok := allindex[check]; ok {
					// Only apply formatTag based on the last index...
					if k == len(m.cfg.MultiTagOID)-1 {
						allindexIt[key1] = formatTag(m.Log, m.cfg.IndexTagFormat, map[string]string{"IDX1": key1, "VAL1": val1, "IDX2": val1, "VAL2": val2}, "VAL2")
						continue
					}
					allindexIt[key1] = val2
				} else {
					// Set debug due to large logs generated. This case is generic, so a not match can be normal
					m.Log.Debugf("[%d] - There is not valid index : %s on TagOID : %s", k, val1, tagcfg.TagOID)
				}
			}
			allindexRes = allindexIt
		}

		if len(allindexOrigin) != len(allindexRes) {
			m.Log.Warnf("Not all indexes have been indirected\n First Idx [%+v]\n Tagged Idx [ %+v]", allindexOrigin, allindexRes)
		}

		return allindexRes, nil

	default:
		return allindex, fmt.Errorf("Uknown provided getmode %s on measurement %s", m.cfg.GetMode, m.ID)
	}
}

// SetupSnmpCli for external use
// set the SNMP client for this measurement
// perform SNMP Connection
// Initialize the measurement
// update statistics
func (m *Measurement) SetupSnmpCli(cli snmp.Client, systemOIDs []string) error {
	m.snmpClient = &cli
	info, err := m.snmpClient.Connect(systemOIDs)
	if err != nil {
		m.Log.Errorf("Not able to connect at the start of the measurement: %v", err)
		return err
	}
	m.stats.SetSysInfo(*info)
	// If the connection is succesfull, initialize
	m.Connected = true
	errInit := m.Init()
	if errInit != nil {
		m.Log.Errorf("Not able to initialize at the start of the measurement: %v", errInit)
		return err
	}
	m.initialized = true
	return nil
}

// GatherLoop do all measurement processing, gathering metrics, handling filters and receiving messages from device
// deviceBus used by device to pass messages to the measurements.
// deviceFreq used if the measurement does not have frequency.
// deviceUpdateFilterFreq is the number of gather loops after a update filters will be done
// gatherLock will be used to achieve sequential gathering (only one measurement goroutine gathering data at the same
// time). It will be enforced if gatherLock is not nil.
func (m *Measurement) GatherLoop(
	busNode *bus.Node,
	snmpCli snmp.Client,
	deviceFreq int,
	deviceUpdateFilterFreq int,

	varMap map[string]interface{},
	tagMap map[string]string,
	systemOIDs []string,
	influxClient *output.InfluxDB,
	gatherLock *sync.Mutex,
) {
	m.Log.Info("MeasurementLoop Fist Check....")

	m.snmpClient = &snmpCli
	// Try to connect for the first time, init metrics and gather data if Enabled
	// if not enabled will be initialized on the main loop
	if m.Active {
		info, err := m.snmpClient.Connect(systemOIDs)
		if err != nil {
			m.Log.Errorf("Not able to connect at the start of the measurement: %v", err)
		} else {
			m.stats.SetSysInfo(*info)
			// If the connection is succesfull, initialize
			m.Connected = true
			errInit := m.Init()
			if errInit != nil {
				m.Log.Errorf("Not able to initialize at the start of the measurement: %v", errInit)
			} else {
				// If connection and initialization are correct, mark the measurement as initiliazed and gather data for the first time
				// Set influxClient as nil, the first time it shouldn't send metrics
				m.initialized = true
				err := m.GatherOnce(gatherLock, varMap, tagMap, nil)
				if err != nil {
					// if error is because of no response from any metric
					// we can suppose the connection has been dropped
					m.Connected = false
					m.stats.SetStatus(m.Active, m.Connected)
					m.Log.Error(err)
				}
			}
		}
	}

	m.Log.Info("MeasurementLoop Init Loop Align....")

	// Measurement Freq overrides Device Freq (creating ticker)
	gatherFreq := deviceFreq
	if m.cfg.Freq != 0 {
		gatherFreq = m.cfg.Freq
	}
	utils.WaitAlignForNextCycle(gatherFreq, m.Log)
	gatherTicker := time.NewTicker(time.Duration(gatherFreq) * time.Second)
	defer gatherTicker.Stop()

	// Measurement Filter Freq overrides Device Filter Freq (creating ticker)
	filterFreq := gatherFreq * deviceUpdateFilterFreq
	if m.cfg.UpdateFltFreq != 0 {
		filterFreq = gatherFreq * m.cfg.UpdateFltFreq
	}
	updateFilterTicker := time.NewTicker(time.Duration(filterFreq) * time.Second)
	defer updateFilterTicker.Stop()

	// updating stats info
	m.stats.GatherFreq = gatherFreq
	m.stats.FilterFreq = filterFreq
	m.stats.SetFilterNextTime(time.Now().Add(time.Duration(filterFreq) * time.Second).Unix())
	m.stats.SetGatherNextTime(time.Now().Add(time.Duration(gatherFreq) * time.Second).Unix())

	for {
		m.Log.Info("MeasurementLoop new Iteration")
		// In each iteration, if measurement is enabled and not connected, try again to connect
		if m.Active && !m.Connected {
			// Connect
			info, err := m.snmpClient.Connect(systemOIDs)
			if err != nil {
				m.Log.Error("Not able to connect inside of the loop: %v", err)
			} else {
				m.Connected = true
				m.stats.SetSysInfo(*info)
			}
		}
		// If measurement is not initialized, do it. Could be because it returned an error while starting.
		if m.Active && m.Connected && !m.initialized {
			err := m.Init()
			if err != nil {
				m.Log.Errorf("Not able to initialize inside of the loop: %v", err)
				// If there is an error initializing, force reconnect
				m.Connected = false
			} else {
				// If connection and initialization are correct, mark the measurement as initialized and gather data for the first time
				m.initialized = true
				// Set influxClient as nil, the first time it shouldn't send metrics
				err := m.GatherOnce(gatherLock, varMap, tagMap, nil)
				if err != nil {
					// if error is because of no response from any metric
					// we can suppose the connection has been dropped
					m.Connected = false
					m.stats.SetStatus(m.Active, m.Connected)
					m.Log.Error(err)
				}
			}
		}
		m.stats.SetStatus(m.Active, m.Connected)
		select {
		case <-updateFilterTicker.C:
			// compute next filter time ( needed to show in the UI )
			m.stats.SetFilterNextTime(time.Now().Add(time.Duration(filterFreq) * time.Second).Unix())
			// filter
			m.filterUpdate()
		case <-gatherTicker.C:
			// compute next gather time ( needed to show in the UI )
			m.stats.SetGatherNextTime(time.Now().Add(time.Duration(gatherFreq) * time.Second).Unix())
			// Gather
			err := m.GatherOnce(gatherLock, varMap, tagMap, influxClient)
			if err != nil {
				// if error is because of no response from any metric
				// we can suppose the connection has been dropped
				m.Connected = false
				m.stats.SetStatus(m.Active, m.Connected)
				m.Log.Error(err)
			}
			// sending only once all statistics for this measurement
			// if not filtered the value should be 0 for filter counters
			m.stats.Send()
		case val, busChannelOpen := <-busNode.Read:
			if !busChannelOpen {
				m.Log.Infof("measurement channel bus closed: exiting")
				return
			}
			m.Log.Infof("measurement received message: %s (%+v)", val.Type, val.Data)
			switch val.Type {
			case bus.FilterUpdate:
				m.filterUpdate()
			case bus.ForceGather:
				m.GatherOnce(gatherLock, varMap, tagMap, influxClient)
			case bus.Enabled:
				active, ok := val.Data.(bool)
				if !ok {
					m.Log.Errorf("invalid value for active bus message: %v", val.Data)
					continue
				}
				m.rtData.Lock()
				m.Active = active
				if active {
					m.stats.SetActive(true)
				} else {
					m.stats.SetStatus(false, false)
					// Close the connection and mark as disconnected
					m.Connected = false
					err := m.snmpClient.Release()
					if err != nil {
						m.Log.Errorf("releasing snmp client in disabled: %v", err)
					}
				}
				m.Log.Infof("STATUS  ACTIVE  [%t] ", active)
				m.rtData.Unlock()
			case bus.SNMPReset:
				err := m.snmpClient.Release()
				if err != nil {
					m.Log.Errorf("releasing snmp client in snmpreset: %v", err)
				}
			case bus.SNMPResetHard:
				err := m.snmpClient.Release()
				if err != nil {
					m.Log.Errorf("releasing snmp client: %v", err)
				}
				m.initialized = false
			case bus.SNMPDebug:
				debug, ok := val.Data.(bool)
				if !ok {
					m.Log.Errorf("invalid value for debug bus message: %v", val.Data)
					continue
				}
				m.snmpClient.SetDebug(debug)
			case bus.SetSNMPMaxRep:
				maxrep, ok := val.Data.(uint8)
				if !ok {
					m.Log.Errorf("invalid value for maxrep bus message: %v", val.Data)
					continue
				}
				m.snmpClient.SetMaxRep(maxrep)
			case bus.Exit, bus.SyncExit:
				m.Log.Info("exit measurement")
				return
			default:
				m.Log.Errorf("unknown command: %v", val)
			}
		}
	}
}

// InitFilters look for filters and add to the measurement with this Filter after it initializes the runtime for the measurement
func (m *Measurement) InitFilters() {
	// check for filters associated with this measurement
	var mfilter *config.MeasFilterCfg
	var multi bool
	// If multi is found, all internal filters are initialized
	// multi must be marked as special...?
	for _, f := range m.measFilters {
		// we search if exist in the filter Database
		if filter, ok := m.mFilters[f]; ok {
			// check and init filters in measurement, applies also in multi
			if ex, mi := m.CheckInitFilter(filter); ex {
				mfilter = filter
				// as filters can be defined without specific order, multi must be persisted
				multi = mi || multi
			}
		}
	}
	// If multi, filters need to be propagated into the internal array and reload all
	if mfilter != nil || multi {
		m.Log.Debugf("filters %s found for device  and measurement %s ", mfilter.ID, m.ID)
		err := m.AddFilter(mfilter, multi)
		if err != nil {
			m.Log.Errorf("Error on initialize Filter for Measurement %s , Error:%s no data will be gathered for this measurement", m.ID, err)
		}
	} else {
		m.Log.Debugf("no filters found for device on measurement %s", m.ID)
	}
	// m.ApplyFilterts...
	// Initialize internal structs after
	m.InitBuildRuntime()
}

// MarshalJSON output a JSON representation of the Measurement ready to be consumed by the API.
// This function will be called by the ToJSON of the device.
// A read lock is taken to avoid trying to read at the same time that the gather process is running.
func (m *Measurement) MarshalJSON() ([]byte, error) {
	m.rtData.RLock()
	m.statsData.Lock()
	defer m.rtData.RUnlock()
	defer m.statsData.Unlock()

	return json.Marshal(&struct {
		ID               string
		MName            string
		TagName          []string
		MetricTable      *metric.MetricTable
		AllIndexedLabels map[string]string
		CurIndexedLabels map[string]string
		FilterCfg        *config.MeasFilterCfg
		Filter           filter.Filter
		MultiIndexMeas   []*Measurement
		Stats            *stats.GatherStats
	}{
		ID:               m.ID,
		MName:            m.MName,
		TagName:          m.TagName,
		MetricTable:      m.MetricTable,
		AllIndexedLabels: m.AllIndexedLabels,
		CurIndexedLabels: m.CurIndexedLabels,
		FilterCfg:        m.FilterCfg,
		Filter:           m.Filter,
		MultiIndexMeas:   m.MultiIndexMeas,
		Stats:            m.Stats,
	})
}

// GatherOnce metrics from device, process them and send values to the backend.
// It also checks if it should run based on the measurement state.
// At the end of the function it could change the status of the connection (to not
// connected) if it was unable to get data
func (m *Measurement) GatherOnce(
	gatherLock *sync.Mutex,
	varMap map[string]interface{},
	tagMap map[string]string,
	influxClient *output.InfluxDB,
) error {
	start := time.Now()
	// Do not gather data if measurement is disabled or it doesn't have a connection or measurement is not initialized
	if !m.Active || !m.Connected || !m.initialized {
		m.Log.Infof("Skip measurement Gathering process Active[%t],Connected[%t],Initialized[%t]", m.Active, m.Connected, m.initialized)
		m.stats.ResetCounters()
		m.statsData.Lock()
		m.Stats = m.getBasicStats()
		m.statsData.Unlock()
		return nil
	}

	// Do not allow the UI to get data from the measurement while it is gathering new data.
	// To have a complete picture, lock until we have all metrics.
	m.rtData.Lock()
	defer m.rtData.Unlock()

	m.Log.Infof("Init gather cycle mode")
	// Mark previous values as old so we can know if new metrics
	// have been gathered
	m.InvalidateMetrics()
	m.stats.ResetCounters()

	m.Log.Debugf("-------Processing measurement : %s", m.ID)

	// If gatherLock its defined it means we should gather data sequentially.
	// We get this lock to avoid another goroutines to gather data at the same time.
	if gatherLock != nil {
		m.Log.Debug("get lock to avoid concurrent gathering")
		gatherLock.Lock()
	}
	// Get data from device and set the values to the snmp metrics structs

	nGets, nProcs, nErrs := m.GetData()
	m.stats.UpdateSnmpGetStats(nGets, nProcs, nErrs)

	m.ComputeOidConditionalMetrics()
	if gatherLock != nil {
		m.Log.Debug("release lock to avoid concurrent gathering")
		gatherLock.Unlock()
	}

	m.ComputeEvaluatedMetrics(varMap)

	// prepare batchpoint
	metSent, metError, measSent, measError, points := m.GetInfluxPoint(tagMap)
	m.stats.AddMeasStats(metSent, metError, measSent, measError)

	sentStats := time.Now()
	// check if the influxClient is nil and skip the send process
	if influxClient != nil {
		bpts, _ := influxClient.BP()
		if bpts != nil {
			(*bpts).AddPoints(points)
			// send data
			influxClient.Send(bpts)
		} else {
			m.Log.Warnf("Can not send data to the output DB because of batchpoint creation error")
		}
	}
	elapsedSentStats := time.Since(sentStats)
	m.stats.AddSentDuration(sentStats, elapsedSentStats)

	end := time.Since(start)
	m.stats.SetGatherDuration(start, end)
	// updating public query stats
	m.statsData.Lock()
	m.Stats = m.getBasicStats()
	m.statsData.Unlock()
	if nProcs == 0 {
		return fmt.Errorf("Device not connected")
	}
	return nil
}

// filterUpdate does ... TODO
func (m *Measurement) filterUpdate() {
	m.Log.Debug("filterUpdate")
	// Do not try to update filters if measurement is disabled or it doesn't have a connection or measurement is not initialized
	if !m.Active || !m.Connected || !m.initialized {
		return
	}
	// Update filters
	if m.cfg.GetMode == "value" {
		return
	}
	start := time.Now()
	changed, err := m.UpdateFilter()
	if err != nil {
		m.Log.Errorf("Error on update Indexes/filter : ERR: %s", err)
		return
	}
	if changed {
		m.InitBuildRuntime()
	}
	duration := time.Since(start)
	m.stats.SetFltUpdateStats(start, duration) // TODO
	m.Log.Infof("snmp INIT runtime measurements/filters took [%s] ", duration)
}

// SetSNMPClient used by unit tests to insert the snmp client
func (m *Measurement) SetSNMPClient(snmpCli snmp.Client) {
	m.snmpClient = &snmpCli
}
