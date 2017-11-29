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

//CheckComputedMetric2 check for computed metrics based on check if variable definition exist
func (mc *MeasurementCfg) CheckComputedMetric2(parameters map[string]interface{}) error {
	var extvars []string
	var allusablevars []string

	//external defined vars ( from Catalog database)
	for k := range parameters {
		extvars = append(extvars, k)
	}
	allusablevars = append(allusablevars, extvars...)
	//internall defined vars( FieldNames and FR/NR)
	intvars, _ := mc.GetInternalVarNames()
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

/*/CheckComputedMetric check for computed metrics based on  Evalutation Execution
func (mc *MeasurementCfg) CheckComputedMetric(parameters map[string]interface{}) error {

	log.Debugf("Building check parrameters array for index measurement %s", mc.ID)
	parameters["NR"] = 1                   //Number of rows (like awk)
	parameters["NF"] = len(mc.FieldMetric) //Number of fields ( like awk)
	//getting all values to the array
	for _, v := range mc.FieldMetric {
		parameters[v.FieldName] = float64(1)
	}
	log.Debugf("PARAMETERS: %+v", parameters)
	//compute Evalutated metrics
	var err error
	var errstr []string
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
}*/

// GetInternalVarNames returns an string array with all posible internal variable Names
func (mc *MeasurementCfg) GetInternalVarNames() ([]string, error) {
	var intvars []string // internal ( from metric ID's)
	//Get InternalVariables (those from field names and "NF" and "NR")
	for _, val := range mc.FieldMetric {
		intvars = append(intvars, val.FieldName)
	}
	for _, val := range mc.EvalMetric {
		intvars = append(intvars, val.FieldName)
	}
	for _, val := range mc.OidCondMetric {
		intvars = append(intvars, val.FieldName)
	}
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

	intvars, err := mc.GetInternalVarNames()
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
	//Check if all evaluated metrics has well defined its parameters as FieldNames
	if varmap == nil {
		varmap = make(map[string]interface{})
	}
	err := mc.CheckComputedMetric2(varmap)
	if err != nil {
		log.Warnf(" This computed Metric has some errors!! please review its definition", err)
		return err
	}

	return nil
}
