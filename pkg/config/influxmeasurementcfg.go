package config

import (
	"errors"
	"fmt"
	"strings"
)

//InfluxMeasurementCfg the measurement configuration
type InfluxMeasurementCfg struct {
	ID   string `xorm:"'id' unique"`
	Name string `xorm:"name"`

	GetMode        string `xorm:"getmode"`  //value ,indexed  (direct tag), indexed_it ( indirect_tag)
	IndexOID       string `xorm:"indexoid"` //only valid if Indexed (direct or indirect)
	TagOID         string `xorm:"tagoid"`   //only valid if inderecta TAG indexeded
	IndexTag       string `xorm:"indextag"`
	IndexTagFormat string `xorm:"indextagformat"`
	IndexAsValue   bool   `xorm:"'indexasvalue' default 0"`
	Fields         []struct {
		ID     string
		Report bool
	} `xorm:"-"` //Got from MeasurementFieldCfg table
	FieldMetric []*SnmpMetricCfg `xorm:"-"`
	EvalMetric  []*SnmpMetricCfg `xorm:"-"`
	Description string           `xorm:"description"`
}

func (mc *InfluxMeasurementCfg) CheckComputedMetric() error {
	parameters := make(map[string]interface{})
	log.Debugf("Building check parrameters array for index measurement %s", mc.ID)
	parameters["NR"] = 1                   //Number of rows (like awk)
	parameters["NF"] = len(mc.FieldMetric) //Number of fields ( like awk)
	//getting all values to the array
	for _, v := range mc.FieldMetric {
		parameters[v.FieldName] = float64(1)
	}
	log.Debugf("PARAMETERS: %+v", parameters)
	//compute Evalutated metrics
	for _, v := range mc.EvalMetric {
		err := v.CheckEvalCfg(parameters)
		if err != nil {
			return fmt.Errorf("Error on metric %s evaluation ERROR : %s", v.ID, err)
		}
		parameters[v.FieldName] = float64(1)
	}
	return nil
}

//Init initialize the measurement configuration
func (mc *InfluxMeasurementCfg) Init(MetricCfg *map[string]*SnmpMetricCfg) error {
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

	for _, f_val := range mc.Fields {
		log.Debugf("looking for measurement %s : fields: %s : Report %t", mc.Name, f_val.ID, f_val.Report)
		if val, ok := (*MetricCfg)[f_val.ID]; ok {
			if val.DataSrcType == "STRINGEVAL" {
				mc.EvalMetric = append(mc.EvalMetric, val)
				log.Debugf("EVAL metric found measurement %s : fields: %s ", mc.Name, f_val.ID)
			} else {

				log.Debugf("found Metric configuration: %s/ %s", f_val.ID, val.BaseOID)
				mc.FieldMetric = append(mc.FieldMetric, val)
			}
		} else {
			log.Warnf("measurement field  %s NOT FOUND in Metrics Database !", f_val.ID)
		}
	}
	//check if there is any field ( should be at least one!!)
	if len(mc.FieldMetric) == 0 {
		return fmt.Errorf("There is no any Field metrics in measurement Config  %s (should be at least one)", mc.ID)
	}
	//Check if duplicated oids
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
	//Check if duplicated fieldNames in any of field/eval Metrics
	fieldnamecheckarray := make(map[string]string)
	for _, v := range mc.FieldMetric {
		//check if the OID has already used as metric in the same measurement
		log.Debugf("VALIDATE MEASUREMENT: %s/%s", v.FieldName, v.ID)
		if v2, ok := fieldnamecheckarray[v.FieldName]; ok {
			//oid has already inserted
			return fmt.Errorf("This measurement has duplicated FieldName[%s] in metric [%s/%s] ", v.FieldName, v.ID, v2)
		}
		oidcheckarray[v.FieldName] = v.ID
	}
	for _, v := range mc.EvalMetric {
		//check if the OID has already used as metric in the same measurement
		log.Debugf("VALIDATE MEASUREMENT: %s/%s", v.FieldName, v.ID)
		if v2, ok := fieldnamecheckarray[v.FieldName]; ok {
			//oid has already inserted
			return fmt.Errorf("This measurement has duplicated FieldName[%s] in metric [%s/%s] ", v.FieldName, v.ID, v2)
		}
		oidcheckarray[v.FieldName] = v.ID
	}
	//Check if all evaluated metrics has well defined its parameters as FieldNames
	err := mc.CheckComputedMetric()
	if err != nil {
		return err
	}

	return nil
}
