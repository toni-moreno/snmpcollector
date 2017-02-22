package impexp

import (
	"encoding/json"
	"fmt"
	"github.com/toni-moreno/snmpcollector/pkg/config"
)

func (e *ExportData) ImportCheck() ([]*ExportObject, error) {

	var duplicated []*ExportObject

	for i := 0; i < len(e.Objects); i++ {
		o := e.Objects[i]
		switch o.ObjectTypeID {
		case "snmpdevicecfg":
			_, err := dbc.GetSnmpDeviceCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		case "influxcfg":
			_, err := dbc.GetInfluxCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		case "measfiltercfg":
			_, err := dbc.GetMeasFilterCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		case "customfiltercfg":
			_, err := dbc.GetCustomFilterCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		case "oidconditioncfg":
			_, err := dbc.GetOidConditionCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		case "measurementcfg":
			_, err := dbc.GetMeasurementCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		case "snmpmetriccfg":
			_, err := dbc.GetSnmpMetricCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		case "measgroupcfg":
			_, err := dbc.GetMGroupsCfgByID(o.ObjectID)
			if err == nil {
				duplicated = append(duplicated, o)
			}
		default:
			return duplicated, fmt.Errorf("Unknown type object type %s ", o.ObjectTypeID)
		}
	}
	if len(duplicated) > 0 {
		return duplicated, fmt.Errorf("There is %d objects duplicated in the config database ", len(duplicated))
	}
	return duplicated, nil
}

func (e *ExportData) Import() error {

	a, err := e.ImportCheck()
	if len(a) > 0 {
		return fmt.Errorf("Error there is  %d objects already in database: %s", len(a), err)
	}

	for i := 0; i < len(e.Objects); i++ {

		o := e.Objects[i]
		log.Debugf("Importing object %+v", o)
		raw, err := json.Marshal(o.ObjectCfg)
		if err != nil {
			return fmt.Errorf("error on reformating object %s: error: %s ", o.ObjectID, err)
		}
		switch o.ObjectTypeID {
		case "snmpdevicecfg":
			log.Debugf("Importing snmpdevicecfg : %+v", o.ObjectCfg)
			data := config.SnmpDeviceCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddSnmpDeviceCfg(data)
			if err != nil {
				return err
			}
		case "influxcfg":
			log.Debugf("Importing influxcfg : %+v", o.ObjectCfg)
			data := config.InfluxCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddInfluxCfg(data)
			if err != nil {
				return err
			}
		case "measfiltercfg":
			log.Debugf("Importing measfiltercfg : %+v", o.ObjectCfg)
			data := config.MeasFilterCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddMeasFilterCfg(data)
			if err != nil {
				return err
			}
		case "customfiltercfg":
			log.Debugf("Importing customfiltercfg : %+v", o.ObjectCfg)
			data := config.CustomFilterCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddCustomFilterCfg(data)
			if err != nil {
				return err
			}
		case "oidconditioncfg":
			log.Debugf("Importing oidconfitioncfg : %+v", o.ObjectCfg)
			data := config.OidConditionCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddOidConditionCfg(data)
			if err != nil {
				return err
			}
		case "measurementcfg":
			data := config.MeasurementCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddMeasurementCfg(data)
			if err != nil {
				return err
			}
		case "snmpmetriccfg":
			log.Debugf("Importing snmpmetric : %+v", o.ObjectCfg)
			data := config.SnmpMetricCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddSnmpMetricCfg(data)
			if err != nil {
				return err
			}
		case "measgroupcfg":
			log.Debugf("Importing measgroupcfg : %+v", o.ObjectCfg)
			data := config.MGroupsCfg{}
			json.Unmarshal(raw, &data)
			_, err := dbc.AddMGroupsCfg(data)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown type object type %s ", o.ObjectTypeID)
		}
	}
	return nil
}
