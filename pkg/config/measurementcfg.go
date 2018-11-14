package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

//MeasurementCfg the measurement configuration
type MeasurementCfg struct {
	ID             string `xorm:"'id' unique" binding:"Required"`
	Name           string `xorm:"name" binding:"Required"`
	GetMode        string `xorm:"getmode" binding:"In(value,indexed,indexed_it)"` //value ,indexed  (direct tag), indexed_it ( indirect_tag)
	IndexOID       string `xorm:"indexoid"`                                       //only valid if Indexed (direct or indirect)
	TagOID         string `xorm:"tagoid"`                                         //only valid if inderecta TAG indexeded
	IndexTag       string `xorm:"indextag"`
	IndexTagFormat string `xorm:"indextagformat"`
	IndexAsValue   bool   `xorm:"'indexasvalue' default 0"`
	Fields         []struct {
		ID     string
		Report int
	} `xorm:"-"` //Got from MeasurementFieldCfg table
	FieldMetric   []*SnmpMetricCfg `xorm:"-" json:"-"`
	EvalMetric    []*SnmpMetricCfg `xorm:"-" json:"-"`
	OidCondMetric []*SnmpMetricCfg `xorm:"-" json:"-"`
	Description   string           `xorm:"description"`
}

//CheckComputedMetricVars check for computed metrics based on check if variable definition exist
func (mc *MeasurementCfg) CheckComputedMetricVars(parameters map[string]interface{}) error {
	var extvars []string
	var allusablevars []string

	//external defined vars ( from Catalog database)
	for k := range parameters {
		extvars = append(extvars, k)
	}
	allusablevars = append(allusablevars, extvars...)
	//internall defined vars( FieldNames and NFR/FR/NR)
	intvars, _ := mc.GetEvaluableVarNames()
	allusablevars = append(allusablevars, intvars...)

	for _, val := range mc.EvalMetric {
		varinmetric, err := val.GetUsedVarNames()
		if err != nil {
			return fmt.Errorf("Error on Get Variables on Metric %s ERROR: %s", val.ID, err)
		}
		log.Debugf("checking if existing var in measurement: %s Metric: %s", mc.ID, val.ID)

		for _, varin := range varinmetric {
			//check if exist on usable vars
			found := false
			for _, varusable := range allusablevars {
				if varin == varusable {
					found = true
					break
				}
			}
			log.Debugf("variable %s : FOUND: %t ", varin, found)
			if !found {
				return fmt.Errorf("variable %s defined in Metric %s not Found  (in any of other metric FieldNames of the measurement, NF, RF and Variable Catalog) ", varin, val.ID)
			}
		}

	}

	return nil
}

// CheckComputedMetricEval check for computed metrics based on  Evalutation Execution
func (mc *MeasurementCfg) CheckComputedMetricEval(parameters map[string]interface{}) error {
	var err error
	var errstr []string
	//get all the eval parameters
	ep, _ := mc.GetEvaluableVarNames()
	for _, t := range ep {
		parameters[t] = float64(1)
	}
	parameters["NF"] = len(mc.FieldMetric) + len(mc.OidCondMetric) //Number of fields ( like awk)
	log.Debugf("PARAMETERS: %+v", parameters)
	for _, v := range mc.EvalMetric {
		err = v.CheckEvalCfg(parameters)
		if err != nil {
			str := fmt.Sprintf("Error on metric %s evaluation ERROR : %s", v.ID, err)
			log.Warnf("%s", str)
			errstr = append(errstr, str)
			//return fmt.Errorf("Error on metric %s evaluation ERROR : %s", v.ID, err)
		}
		parameters[v.FieldName] = float64(1)
	}
	if err != nil {
		return fmt.Errorf("%s", strings.Join(errstr, "\n"))
	}
	return nil
}

// GetEvaluableVarNames returns an string array with all posible internal variable Names
func (mc *MeasurementCfg) GetEvaluableVarNames() ([]string, error) {
	var intvars []string // internal ( from metric ID's)
	//Get InternalVariables (those from field names and "NF" and "NR")
	for _, val := range mc.FieldMetric {
		vars, err := val.GetEvaluableVarNames()
		if err != nil {
			return intvars, err
		}
		intvars = append(intvars, vars...)
	}
	for _, val := range mc.EvalMetric {
		intvars = append(intvars, val.FieldName)
	}
	for _, val := range mc.OidCondMetric {
		intvars = append(intvars, val.FieldName)
	}
	intvars = append(intvars, "NFR")
	intvars = append(intvars, "NF")
	intvars = append(intvars, "NR")
	//log.Debugf("INTVARS %s : %#+v ", mc.ID, intvars)
	return utils.RemoveDuplicatesUnordered(intvars), nil
}

// GetAllUsedVarNamesInMetrics return all used Var names in Metrics
func (mc *MeasurementCfg) GetAllUsedVarNamesInMetrics() ([]string, error) {
	var allvars []string // all
	// Get All used Variables in Formulas
	for _, val := range mc.EvalMetric {
		vars, err := val.GetUsedVarNames()
		if err != nil {
			return nil, err
		}
		log.Debugf("checking if existing var sin measurment %s : %#+v ", mc.ID, vars)
		allvars = append(allvars, vars...)
	}

	log.Debugf("ALLVARS %s : %#+v ", mc.ID, allvars)
	return utils.RemoveDuplicatesUnordered(allvars), nil
}

// GetExternalVars Get Needed External Variables in this Measurement
func (mc *MeasurementCfg) GetExternalVars() ([]string, error) {

	intvars, err := mc.GetEvaluableVarNames()
	if err != nil {
		return nil, err
	}

	allvars, err := mc.GetAllUsedVarNamesInMetrics()
	if err != nil {
		return nil, err
	}

	//get difference
	extvars := utils.DiffSlice(intvars, allvars)

	log.Debugf("EXTVARS %s : %#+v ", mc.ID, extvars)

	return extvars, nil
}

//Init initialize the measurement configuration
func (mc *MeasurementCfg) Init(MetricCfg *map[string]*SnmpMetricCfg, varmap map[string]interface{}) error {
	//mc.ID = name
	//validate config values
	if len(mc.Name) == 0 {
		return errors.New("Name not set in measurement Config " + mc.ID)
	}
	if len(mc.Fields) == 0 {
		return errors.New("No Fields added to measurement " + mc.ID)
	}

	switch mc.GetMode {
	case "indexed", "indexed_it":
		if len(mc.IndexOID) == 0 {
			return errors.New("Indexed measurement with no IndexOID in measurement Config " + mc.ID)
		}
		if len(mc.IndexTag) == 0 {
			return errors.New("Indexed measurement with no IndexTag configuredin measurement " + mc.ID)
		}
		if !strings.HasPrefix(mc.IndexOID, ".") {
			return errors.New("Bad BaseOid format:" + mc.IndexOID + " in metric Config " + mc.ID)
		}
		if mc.GetMode == "indexed_it" {
			if !strings.HasPrefix(mc.TagOID, ".") {
				return errors.New("Bad BaseOid format:" + mc.TagOID + "  for  indirect TAG OID in metric Config " + mc.ID)
			}
		}

	case "value":
	default:
		return errors.New("Unknown GetMode" + mc.GetMode + " in measurement Config " + mc.ID)
	}

	log.Infof("processing measurement key: %s ", mc.ID)
	log.Debugf("%+v", mc)

	for _, fVal := range mc.Fields {
		log.Debugf("looking for measurement %s : fields: %s : Report %d", mc.Name, fVal.ID, fVal.Report)
		if val, ok := (*MetricCfg)[fVal.ID]; ok {
			switch val.DataSrcType {
			case "STRINGEVAL":
				mc.EvalMetric = append(mc.EvalMetric, val)
				log.Debugf("STRING EVAL metric found measurement %s : fields: %s ", mc.Name, fVal.ID)
			case "CONDITIONEVAL":
				mc.OidCondMetric = append(mc.OidCondMetric, val)
				log.Debugf("OID CONDITION EVAL metric found measurement %s : fields: %s ", mc.Name, fVal.ID)
			default:
				log.Debugf("found Metric configuration: %s/ %s", fVal.ID, val.BaseOID)
				mc.FieldMetric = append(mc.FieldMetric, val)
			}
		} else {
			log.Warnf("measurement field  %s NOT FOUND in Metrics Database !", fVal.ID)
		}
	}
	//check for valid fields ( should be at least one!! Field in indexed measurements and at least one field or ) in
	switch mc.GetMode {
	case "indexed", "indexed_it":
		if len(mc.FieldMetric) == 0 {
			return fmt.Errorf("There is no any Field metrics in measurement type \"%s\" Config  %s (should be at least one)", mc.GetMode, mc.ID)
		}
	case "value":
		if (len(mc.FieldMetric) + len(mc.OidCondMetric)) == 0 {
			return fmt.Errorf("There is no any Field or OID conditional metrics in measurement type \"value\" Config  %s (should be at least one)", mc.ID)
		}
	}
	//Check if duplicated oids for Field metrics
	oidcheckarray := make(map[string]string)
	for _, v := range mc.FieldMetric {
		//check if the OID has already used as metric in the same measurement
		log.Debugf("VALIDATE MEASUREMENT: %s/%s", v.BaseOID, v.ID)
		if v2, ok := oidcheckarray[v.BaseOID]; ok {
			//oid has already inserted
			return fmt.Errorf("This measurement has duplicated OID[%s] in metric [%s/%s] ", v.BaseOID, v.ID, v2)
		}
		oidcheckarray[v.BaseOID] = v.ID
	}
	//Check if duplicated fieldNames in any of field/eval/oidCondition Metrics
	fieldnamecheckarray := make(map[string]string)
	for _, v := range mc.FieldMetric {
		log.Debugf("VALIDATE MEASUREMENT: %s/%s", v.FieldName, v.ID)
		if v2, ok := fieldnamecheckarray[v.FieldName]; ok {
			//field name has already inserted
			return fmt.Errorf("This measurement has duplicated FieldName[%s] in metric [%s/%s] ", v.FieldName, v.ID, v2)
		}
		fieldnamecheckarray[v.FieldName] = v.ID
	}
	for _, v := range mc.EvalMetric {
		log.Debugf("VALIDATE MEASUREMENT: %s/%s", v.FieldName, v.ID)
		if v2, ok := fieldnamecheckarray[v.FieldName]; ok {
			//field name has already inserted
			return fmt.Errorf("This measurement has duplicated FieldName[%s] in metric [%s/%s] ", v.FieldName, v.ID, v2)
		}
		fieldnamecheckarray[v.FieldName] = v.ID
	}
	for _, v := range mc.OidCondMetric {
		log.Debugf("VALIDATE MEASUREMENT: %s/%s", v.FieldName, v.ID)
		if v2, ok := fieldnamecheckarray[v.FieldName]; ok {
			//field name has already inserted
			return fmt.Errorf("This measurement has duplicated FieldName[%s] in metric [%s/%s] ", v.FieldName, v.ID, v2)
		}
		fieldnamecheckarray[v.FieldName] = v.ID
	}
	//Check if all evaluated metrics has well defined its parameters as FieldNames and evaluation syntax
	var err error
	if varmap == nil {
		varmap = make(map[string]interface{})
	}
	varmapcopy := make(map[string]interface{})
	for k, v := range varmap {
		varmapcopy[k] = v
	}
	err = mc.CheckComputedMetricVars(varmapcopy)
	if err != nil {
		log.Warnf(" This computed Metric has Variable errors!! please review its definition Error: %s", err)
		return err
	}

	err = mc.CheckComputedMetricEval(varmapcopy)
	if err != nil {
		log.Warnf(" This computed Metric has Evaluation errors!! please review its definition Error: %s", err)
		return err
	}

	return nil
}

/***************************
	MEASUREMENTS
	-GetMeasurementCfgByID(struct)
	-GetMeasurementCfgMap (map - for interna config use
	-GetMeasurementCfgArray(Array - for web ui use )
	-AddMeasurementCfg
	-DelMeasurementCfg
	-UpdateMeasurementCfg
  -GetMeasurementCfgAffectOnDel
***********************************/

/*GetMeasurementCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetMeasurementCfgByID(id string) (MeasurementCfg, error) {
	cfgarray, err := dbc.GetMeasurementCfgArray("id='" + id + "'")
	if err != nil {
		return MeasurementCfg{}, err
	}
	if len(cfgarray) > 1 {
		return MeasurementCfg{}, fmt.Errorf("Error %d results on get MeasurementCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return MeasurementCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the measurement config table", id)
	}
	return *cfgarray[0], nil
}

/*GetMeasurementCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetMeasurementCfgMap(filter string) (map[string]*MeasurementCfg, error) {
	cfgarray, err := dbc.GetMeasurementCfgArray(filter)
	cfgmap := make(map[string]*MeasurementCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetMeasurementCfgArray generate an array of measurements with all its information */
func (dbc *DatabaseCfg) GetMeasurementCfgArray(filter string) ([]*MeasurementCfg, error) {
	var err error
	var devices []*MeasurementCfg
	//Get Only data for selected measurements
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get MeasurementCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get MeasurementCfg   data: %v\n", err)
			return nil, err
		}
	}

	var MeasureMetric []*MeasurementFieldCfg
	if err = dbc.x.Find(&MeasureMetric); err != nil {
		log.Warnf("Fail to get Measurements Metric relationship data: %v\n", err)
	}

	//Load Measurements and metrics relationship
	//We assign field metric ID to each measurement
	for _, mVal := range devices {
		for _, mm := range MeasureMetric {
			if mm.IDMeasurementCfg == mVal.ID {
				data := struct {
					ID     string
					Report int
				}{
					mm.IDMetricCfg,
					mm.Report,
				}
				mVal.Fields = append(mVal.Fields, data)
			}
		}
	}
	return devices, nil
}

/*AddMeasurementCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddMeasurementCfg(dev MeasurementCfg) (int64, error) {
	var err error
	var affected, newmf int64

	// create SnmpMetricCfg to check if any configuration issue found before persist to database
	// We need to get data from database
	cfg, _ := dbc.GetSnmpMetricCfgMap("")
	gv, _ := dbc.GetVarCatalogCfgMap("")

	err = dev.Init(&cfg, CatalogVar2Map(gv))
	if err != nil {
		return 0, err
	}
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//Measurement Fields
	for _, metric := range dev.Fields {

		mstruct := MeasurementFieldCfg{
			IDMeasurementCfg: dev.ID,
			IDMetricCfg:      metric.ID,
			Report:           metric.Report,
		}
		newmf, err = session.Insert(&mstruct)
		if err != nil {
			session.Rollback()
			return 0, err
		}
	}
	//no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Added new Measurement Successfully with id %s and [%d Fields] ", dev.ID, newmf)
	dbc.addChanges(affected + newmf)
	return affected, nil
}

/*DelMeasurementCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelMeasurementCfg(id string) (int64, error) {
	var affectedfl, affectedmg, affectedft, affectedcf, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in MeasurementFieldCfg
	affectedfl, err = session.Where("id_measurement_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement on MeasurementFieldCfg with id: %s, error: %s", id, err)
	}

	affectedmg, err = session.Where("id_measurement_cfg='" + id + "'").Delete(&MGroupsMeasurements{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	affectedft, err = session.Where("id_measurement_cfg='" + id + "'").Cols("id_measurement_cfg").Update(&MeasFilterCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Update FilterMeasurement on with id: %s, error: %s", id, err)
	}

	//CustomFilter Related Dev
	affectedcf, err = session.Where("related_meas='" + id + "'").Cols("related_meas").Update(&CustomFilterCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement with id on delete CustomFilter with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&MeasurementCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Measurement with ID %s [ %d Measurements Groups Affected / %d Fields Affected / %d Filters Afected / %d Custom Filters Afected ]", id, affectedmg, affectedfl, affectedft, affectedcf)
	dbc.addChanges(affected + affectedmg + affectedfl + affectedft + affectedcf)
	return affected, nil
}

/*UpdateMeasurementCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateMeasurementCfg(id string, dev MeasurementCfg) (int64, error) {
	var affecteddev, newmf, affected int64
	var err error
	// create SnmpMetricCfg to check if any configuration issue found before persist to database.
	// config should be got from database
	// TODO: filter only metrics in Measurement to test if measurement was well defined
	cfg, _ := dbc.GetSnmpMetricCfgMap("")
	gv, _ := dbc.GetVarCatalogCfgMap("")

	err = dev.Init(&cfg, CatalogVar2Map(gv))
	if err != nil {
		return 0, err
	}
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		log.Infof("Updated Measurement Config to %d devices ", affecteddev)

		affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Cols("id_measurement_cfg").Update(&MGroupsMeasurements{IDMeasurementCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Measurement id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Cols("id_measurement_cfg").Update(&MeasFilterCfg{IDMeasurementCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Measurement id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		affecteddev, err = session.Where("related_meas='" + id + "'").Cols("related_meas").Update(&CustomFilterCfg{RelatedMeas: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Measurement id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}

		log.Infof("Updated Measurement config to %d devices ", affecteddev)
	}
	//delete all previous values
	affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	//Creating nuew Measurement Fields
	for _, metric := range dev.Fields {

		mstruct := MeasurementFieldCfg{
			IDMeasurementCfg: dev.ID,
			IDMetricCfg:      metric.ID,
			Report:           metric.Report,
		}
		newmf, err = session.Insert(&mstruct)
		if err != nil {
			session.Rollback()
			return 0, err
		}
	}
	//update data
	affected, err = session.Where("id='" + id + "'").UseBool().AllCols().Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Infof("Updated Influx Measurement Config Successfully with id %s and  (%d previous / %d new Fields), affected", id, affecteddev, newmf)
	dbc.addChanges(affecteddev + newmf)
	return affected, nil
}

/*GetMeasurementCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetMeasurementCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var mf []*MeasurementFieldCfg
	var mg []*MGroupsMeasurements
	var cf []*CustomFilterCfg
	var obj []*DbObjAction
	var err error
	err = dbc.x.Where("id_measurement_cfg='" + id + "'").Find(&mf)
	if err != nil {
		return nil, fmt.Errorf("Error on Delete Measurement on MeasurementFieldCfg with id: %s, error: %s", id, err)
	}
	for _, val := range mf {
		obj = append(obj, &DbObjAction{
			Type:     "snmpmetriccfg",
			TypeDesc: "Metrics",
			ObID:     val.IDMetricCfg,
			Action:   "Delete SNMPMetric field from Measurement relation",
		})
	}

	err = dbc.x.Where("id_measurement_cfg='" + id + "'").Find(&mg)
	if err != nil {
		return nil, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	for _, val := range mg {
		obj = append(obj, &DbObjAction{
			Type:     "measgroupcfg",
			TypeDesc: "Meas. Groups",
			ObID:     val.IDMGroupCfg,
			Action:   "Delete Measurement from Measurement Group relation",
		})
	}

	err = dbc.x.Where("related_meas='" + id + "'").Find(&cf)
	if err != nil {
		return nil, fmt.Errorf("Error on Delete Measurement on MeasurementFieldCfg with id: %s, error: %s", id, err)
	}
	for _, val := range cf {
		obj = append(obj, &DbObjAction{
			Type:     "customfiltercfg",
			TypeDesc: "Custom Filter",
			ObID:     val.ID,
			Action:   "Delete related Measurement from CustomFilter",
		})
	}

	return obj, nil
}
