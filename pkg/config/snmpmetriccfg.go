package config

import (
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"

	"strings"
)

//SnmpMetricCfg Metric config
type SnmpMetricCfg struct {
	ID          string  `xorm:"'id' unique"` //name of the key in the config array
	FieldName   string  `xorm:"field_name"`
	Description string  `xorm:"description"`
	BaseOID     string  `xorm:"baseoid"`
	DataSrcType string  `xorm:"datasrctype"`
	GetRate     bool    `xorm:"getrate"` //ony Valid with COUNTERS
	Scale       float64 `xorm:"scale"`
	Shift       float64 `xorm:"shift"`
	IsTag       bool    `xorm:"'istag' default 0"`
	ExtraData   string  `xorm:"extradata"` //Only Valid with STRINGPARSER and STRINGEVAL
}

/*
3.- Check minimal data is set  (pending)
name, BaseOID BaseOID begining with "."
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
	case "COUNTERXX": //raw and Coocked increment with non_negative behaivour of Counters
	case "TimeTicks", "TIMETICKS": //raw and cooked to second of timeticks
	case "OCTETSTRING":
	case "HWADDR":
	case "IpAddress":
	case "STRINGPARSER":
	case "STRINGEVAL":
	case "CONDITIONEVAL":
	default:
		return errors.New("UnkNown DataSourceType:" + m.DataSrcType + " in metric Config " + m.ID)
	}
	if m.DataSrcType != "STRINGEVAL" && m.DataSrcType != "CONDITIONEVAL" && !strings.HasPrefix(m.BaseOID, ".") {
		return errors.New("Bad BaseOid format:" + m.BaseOID + " in metric Config " + m.ID)
	}
	if m.DataSrcType == "STRINGPARSER" && len(m.ExtraData) == 0 {
		return errors.New("STRINGPARSER type requires extradata to work " + m.ID)
	}
	if m.DataSrcType == "STRINGEVAL" && len(m.ExtraData) == 0 {
		return fmt.Errorf("ExtraData not set in metric Config %s type  %s"+m.ID, m.DataSrcType)
	}
	if m.DataSrcType == "CONDITIONEVAL" && len(m.ExtraData) == 0 {
		return fmt.Errorf("ExtraData not set in metric Config %s type  %s"+m.ID, m.DataSrcType)
	}
	return nil
}

func (m SnmpMetricCfg) CheckEvalCfg(parameters map[string]interface{}) error {
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
