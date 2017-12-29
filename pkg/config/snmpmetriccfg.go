package config

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
)

// MetricMultiMap Value
type MetricMultiMap struct {
	IType string
	IName string
	IConv string
	Value interface{}
}

//SnmpMetricCfg Metric config
type SnmpMetricCfg struct {
	ID          string         `xorm:"'id' unique" binding:"Required"` //name of the key in the config array
	FieldName   string         `xorm:"field_name" binding:"Required"`
	Description string         `xorm:"description"`
	BaseOID     string         `xorm:"baseoid"`
	DataSrcType string         `xorm:"datasrctype"`
	GetRate     bool           `xorm:"getrate"` //ony Valid with COUNTERS
	Scale       float64        `xorm:"scale"`
	Shift       float64        `xorm:"shift"`
	IsTag       bool           `xorm:"'istag' default 0"` //Not Valid on  MULTISTRINGPARSER
	ExtraData   string         `xorm:"extradata"`         //Only Valid with STRINGPARSER, MULTISTRINGPARSER, STRINGEVAL , BITS , BITSCHK
	Names       map[int]string `xorm:"-" json:"-"`        //BitString Name array
}

/*
Init Initializes Metric
3.- Check minimal data is set  (pending)
name, BaseOID BaseOID beginning with "."
fieldname != null
*/
// Init initialize metrics
func (m *SnmpMetricCfg) Init() error {
	//valIDate config values
	if len(m.FieldName) == 0 {
		return errors.New("FieldName not set in metric Config " + m.ID)
	}
	if len(m.BaseOID) == 0 && m.DataSrcType != "STRINGEVAL" && m.DataSrcType != "CONDITIONEVAL" {
		return fmt.Errorf("BaseOid not set in metric Config %s type  %s"+m.ID, m.DataSrcType)
	}
	//https://tools.ietf.org/html/rfc2578 (SMIv2)
	//https://tools.ietf.org/html/rfc2579 (Textual Conventions for SMIv2)
	//https://tools.ietf.org/html/rfc2851 (Textual Conventions for Internet Network Address)
	switch m.DataSrcType {
	case "INTEGER", "Integer32":
	case "Gauge32":
	case "UInteger32", "Unsigned32":
	case "Counter32", "COUNTER32": //raw and cooked increment of Counter32
	case "Counter64", "COUNTER64": //raw and Cooked increment of Counter64
	case "COUNTERXX": //raw and Coocked increment with non_negative behaviour of Counters
	case "TimeTicks", "TIMETICKS": //raw and cooked to second of timeticks
	case "BITS", "BITSCHK":
	case "OCTETSTRING":
	case "OID":
	case "HWADDR":
	case "IpAddress":
	case "STRINGPARSER":
	case "MULTISTRINGPARSER":
	case "STRINGEVAL":
	case "CONDITIONEVAL":
	default:
		return errors.New("UnkNown DataSourceType:" + m.DataSrcType + " in metric Config " + m.ID)
	}
	if m.DataSrcType == "BITSCHK" {
		if len(m.ExtraData) == 0 {
			return errors.New("BITSCHK type requires extradata to work " + m.ID)
		}
		_, err := strconv.Atoi(m.ExtraData)
		if err != nil {
			return errors.New("BITSCHK type requires extradata to be a positive integer to work: ERROR " + err.Error())
		}
	}
	if m.DataSrcType == "BITS" {
		if len(m.ExtraData) == 0 {
			return errors.New("BITS type requires extradata to work " + m.ID)
		}
		//named bits array construction for this Config
		re := regexp.MustCompile("([a-zA-Z0-9]+)\\(([0-9]+)\\)")
		m.Names = make(map[int]string)
		str := re.FindAllStringSubmatch(m.ExtraData, -1)
		for _, x := range str {
			i, _ := strconv.Atoi(x[2])
			m.Names[i] = x[1]
		}
	}
	if m.DataSrcType != "STRINGEVAL" && m.DataSrcType != "CONDITIONEVAL" && !strings.HasPrefix(m.BaseOID, ".") {
		return errors.New("Bad BaseOid format:" + m.BaseOID + " in metric Config " + m.ID)
	}
	if m.DataSrcType == "STRINGPARSER" && len(m.ExtraData) == 0 {
		return errors.New("STRINGPARSER type requires extradata to work " + m.ID)
	}

	if m.DataSrcType == "MULTISTRINGPARSER" {
		if len(m.ExtraData) == 0 {
			return errors.New("MULTISTRINGPARSER type requires extradata to work " + m.ID)
		}
		//Check Field Syntax.
		ma, err := m.GetMultiStringTagFieldMap()
		if err != nil {
			return fmt.Errorf("MULTISTRINGPARSER ITEM definition Format Error %s type  %s", m.ID, err)
		}
		//Check regex syntax
		re, err := regexp.Compile(m.ExtraData)
		if err != nil {
			return fmt.Errorf("MULTISTRINGPARSER Regexp Format Error %s type  %s", m.ID, err)
		}
		//Check number of capturing groups are the same as ITEM definitions
		ncg := len(re.SubexpNames())
		if ncg != len(ma)+1 {
			return fmt.Errorf("MULTISTRINGPARSER  %s Format Error %d ITEMS defined different than %d Capturing groups", m.ID, len(ma), ncg-1)
		}

	}
	if m.DataSrcType == "STRINGEVAL" && len(m.ExtraData) == 0 {
		return fmt.Errorf("ExtraData not set in metric Config %s type  %s", m.ID, m.DataSrcType)
	}
	if m.DataSrcType == "CONDITIONEVAL" && len(m.ExtraData) == 0 {
		return fmt.Errorf("ExtraData not set in metric Config %s type  %s", m.ID, m.DataSrcType)
	}
	return nil
}

// CheckEvalCfg : check evaluated expresion based in govaluate
func (m *SnmpMetricCfg) CheckEvalCfg(parameters map[string]interface{}) error {
	if m.DataSrcType != "STRINGEVAL" {
		return nil
	}
	expression, err := govaluate.NewEvaluableExpression(m.ExtraData)
	if err != nil {
		//log.Errorf("Error on initialice STRINGEVAL on metric %s evaluation : %s : ERROR : %s", m.ID, m.ExtraData, err)
		return err
	}
	_, err = expression.Evaluate(parameters)
	if err != nil {
		//log.Errorf("Error in metric %s On EVAL string: %s : ERROR : %s", m.ID, m.ExtraData, err)
		return err
	}
	return nil
}

// GetMultiStringTagFieldMap get tag/field description map
func (m *SnmpMetricCfg) GetMultiStringTagFieldMap() ([]*MetricMultiMap, error) {

	var retval []*MetricMultiMap

	items := strings.Split(m.FieldName, ",")

	for _, v := range items {
		itcfg := strings.Split(v, "|")
		//checklength
		iType := itcfg[0] //T/F
		if iType != "T" && iType != "F" {
			str := fmt.Sprintf("MultiString Parse Config error on Metric %s  Type %s is not of type (T=Tag ) or (F=Field)", m.ID, itcfg)
			log.Warnf(str)
			return nil, errors.New(str)
		}
		// Name
		iName := itcfg[1] //name
		//Default Conversions
		var iConv string
		if iType == "T" {
			iConv = "STR"
		}
		if iType == "F" {
			iConv = "INT"
		}
		if len(itcfg) > 2 {
			switch itcfg[2] {
			case "STR":
			case "BL":
			case "INT":
			case "FP":
			default:
				str := fmt.Sprintf("MultiString Parse Config error on Metric %s Conversion Type (%s) should be of type STR|INT|FP|BL", m.ID, itcfg[2])
				log.Errorf(str)
				return nil, errors.New(str)
			}
			iConv = itcfg[2]
		}
		retval = append(retval, &MetricMultiMap{IName: iName, IType: iType, IConv: iConv})
		// Could be null on Tag type
	}
	return retval, nil
}

// GetUsedVarNames Get Needed External Variables on this Metric ( only vaid in STRINGEVAL)
func (m *SnmpMetricCfg) GetUsedVarNames() ([]string, error) {
	if m.DataSrcType != "STRINGEVAL" {
		return nil, nil
	}
	expression, err := govaluate.NewEvaluableExpression(m.ExtraData)
	if err != nil {

		return nil, err
	}
	return expression.Vars(), nil
}

// GetEvaluableVarNames get a string array with all posible metric values to use in metric evaluatator
func (m *SnmpMetricCfg) GetEvaluableVarNames() ([]string, error) {
	if m.DataSrcType == "MULTISTRINGPARSER" {
		mmap, err := m.GetMultiStringTagFieldMap()
		if err != nil {
			return nil, err
		}
		var retarray []string
		for _, mm := range mmap {
			retarray = append(retarray, mm.IName)
		}
		return retarray, nil
	}
	return []string{m.FieldName}, nil
}

// GetMetricHeader Get Info
func (m *SnmpMetricCfg) GetMetricHeader(report int) interface{} {
	var retval interface{}
	switch m.DataSrcType {
	case "STRINGPARSER", "BITS", "BITSCHK", "CONDITIONEVAL", "STRINGEVAL":
		retval = &struct {
			FieldID     string
			Type        string
			BaseOID     string
			ExtraData   string
			IsTag       bool
			Report      int
			Description string
		}{
			FieldID:     m.ID,
			Type:        m.DataSrcType,
			BaseOID:     m.BaseOID,
			ExtraData:   m.ExtraData,
			IsTag:       m.IsTag,
			Report:      report,
			Description: m.Description,
		}
	case "MULTISTRINGPARSER":
		retval = &struct {
			Title       string
			FieldID     string
			Type        string
			BaseOID     string
			ExtraData   string
			Report      int
			Description string
		}{
			Title:       m.ID,
			FieldID:     m.FieldName,
			Type:        m.DataSrcType,
			BaseOID:     m.BaseOID,
			ExtraData:   m.ExtraData,
			Report:      report,
			Description: m.Description,
		}
	default:
		retval = &struct {
			FieldID     string
			Type        string
			BaseOID     string
			IsTag       bool
			Report      int
			Description string
		}{
			FieldID:     m.ID,
			Type:        m.DataSrcType,
			BaseOID:     m.BaseOID,
			IsTag:       m.IsTag,
			Report:      report,
			Description: m.Description,
		}
	}
	return retval
}

/***************************
SNMP Metric
	-GetSnmpMetricCfgCfgByID(struct)
	-GetSnmpMetricCfgMap (map - for interna config use
	-GetSnmpMetricCfgArray(Array - for web ui use )
	-AddSnmpMetricCfg
	-DelSnmpMetricCfg
	-UpdateSnmpMetricCfg
  -GetSnmpMetricCfgAffectOnDel
***********************************/

/*GetSnmpMetricCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetSnmpMetricCfgByID(id string) (SnmpMetricCfg, error) {
	cfgarray, err := dbc.GetSnmpMetricCfgArray("id='" + id + "'")
	if err != nil {
		return SnmpMetricCfg{}, err
	}
	if len(cfgarray) > 1 {
		return SnmpMetricCfg{}, fmt.Errorf("Error %d results on get SnmpMetricCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return SnmpMetricCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the metric config table", id)
	}
	return *cfgarray[0], nil
}

/*GetSnmpMetricCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetSnmpMetricCfgMap(filter string) (map[string]*SnmpMetricCfg, error) {
	cfgarray, err := dbc.GetSnmpMetricCfgArray(filter)
	cfgmap := make(map[string]*SnmpMetricCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetSnmpMetricCfgArray generate an array of metrics with all its information */
func (dbc *DatabaseCfg) GetSnmpMetricCfgArray(filter string) ([]*SnmpMetricCfg, error) {
	var err error
	var devices []*SnmpMetricCfg
	//Get Only data for selected metrics
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get SnmpMetricCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get SnmpMetricCfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddSnmpMetricCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddSnmpMetricCfg(dev SnmpMetricCfg) (int64, error) {
	var err error
	var affected int64
	// create SnmpMetricCfg to check if any configuration issue found before persist to database.
	err = dev.Init()
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
	//no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Added new Snmp Metric Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelSnmpMetricCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelSnmpMetricCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in Measurements

	affecteddev, err = session.Where("id_metric_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Metric with id on delete MeasurementFieldCfg with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&SnmpMetricCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Metricdb with ID %s [ %d Measurements Affected  ]", id, affecteddev)
	dbc.addChanges(affecteddev)
	return affected, nil
}

/*UpdateSnmpMetricCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateSnmpMetricCfg(id string, dev SnmpMetricCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	// create SnmpMetricCfg to check if any configuration issue found before persist to database.
	err = dev.Init()
	if err != nil {
		return 0, err
	}
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("id_metric_cfg='" + id + "'").Cols("id_metric_cfg").Update(&MeasurementFieldCfg{IDMetricCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Metric id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated SnmpMetric Config to %s devices ", affecteddev)
	}

	affected, err = session.Where("id='" + id + "'").UseBool().AllCols().Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Infof("Updated SnmpMetric Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetSnmpMetricCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetSnmpMetricCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var devices []*MeasurementFieldCfg
	var obj []*DbObjAction
	if err := dbc.x.Where("id_metric_cfg='" + id + "'").Find(&devices); err != nil {
		log.Warnf("Error on Get Snmp Metric Cfg id %d for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range devices {
		obj = append(obj, &DbObjAction{
			Type:     "measurementcfg",
			TypeDesc: "Measurements",
			ObID:     val.IDMeasurementCfg,
			Action:   "Delete SNMPMetric field from Measurement relation",
		})

	}
	return obj, nil
}
