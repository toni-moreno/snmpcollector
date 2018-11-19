package config

import (
	"fmt"
	"strings"
)

// VarCatalogCfg is the main configuration for any InfluxDB TSDB
type VarCatalogCfg struct {
	// Variable Name unique
	ID          string `xorm:"'id' unique" binding:"Required"`
	Type        string `xorm:"type" binding:"Required"`
	Value       string `xorm:"value"`
	Description string `xorm:"description"`
}

/***************************
Global Var
	-GetVarCatalogCfgCfgByID(struct)
	-GetVarCatalogCfgMap (map - for interna config use
	-GetVarCatalogCfgArray(Array - for web ui use )
	-AddVarCatalogCfg
	-DelVarCatalogCfg
	-UpdateVarCatalogCfg
  -GetVarCatalogCfgAffectOnDel
***********************************/

/*GetVarCatalogCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetVarCatalogCfgByID(id string) (VarCatalogCfg, error) {
	cfgarray, err := dbc.GetVarCatalogCfgArray("id='" + id + "'")
	if err != nil {
		return VarCatalogCfg{}, err
	}
	if len(cfgarray) > 1 {
		return VarCatalogCfg{}, fmt.Errorf("Error %d results on get VarCatalogCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return VarCatalogCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the Global Var config table", id)
	}
	return *cfgarray[0], nil
}

/*GetVarCatalogCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetVarCatalogCfgMap(filter string) (map[string]*VarCatalogCfg, error) {
	cfgarray, err := dbc.GetVarCatalogCfgArray(filter)
	cfgmap := make(map[string]*VarCatalogCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetVarCatalogCfgArray generate an array of metrics with all its information */
func (dbc *DatabaseCfg) GetVarCatalogCfgArray(filter string) ([]*VarCatalogCfg, error) {
	var err error
	var devices []*VarCatalogCfg
	//Get Only data for selected metrics
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get VarCatalogCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get VarCatalogCfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddVarCatalogCfg for adding new Global Variable*/
func (dbc *DatabaseCfg) AddVarCatalogCfg(dev VarCatalogCfg) (int64, error) {
	var err error
	var affected int64

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
	log.Infof("Added new Global Variable Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelVarCatalogCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelVarCatalogCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in Measurements

	affected, err = session.Where("id='" + id + "'").Delete(&VarCatalogCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Global Var with ID %s [ %d Measurements Affected  ]", id, affecteddev)
	dbc.addChanges(affecteddev)
	return affected, nil
}

/*UpdateVarCatalogCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateVarCatalogCfg(id string, dev VarCatalogCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	// create VarCatalogCfg to check if any configuration issue found before persist to database.

	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		var metrics []*SnmpMetricCfg
		session.Where("datasrctype = 'STRINGEVAL' and extradata like '%" + id + "%'").Find(&metrics)
		for _, v := range metrics {
			v.ExtraData = strings.Replace(v.ExtraData, id, dev.ID, -1)
			_, err = session.Where("id='" + v.ID + "'").UseBool().AllCols().Update(v)
			if err != nil {
				session.Rollback()
				return 0, err
			}
			log.Infof("Updated STRING EVAL Metric %s devices old variable name %s new %s", v.ID, dev.ID, id)
		}
		log.Infof("Updated VarCatalogiableConfig to %d devices ", affecteddev)
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

	log.Infof("Updated VarCatalogiableConfig Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetVarCatalogCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetVarCatalogCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	//var devices []*MeasurementFieldCfg
	var obj []*DbObjAction
	/*
		if err := dbc.x.Where("id_metric_cfg='" + id + "'").Find(&devices); err != nil {
			log.Warnf("Error on Get Snmp Metric Cfg id %d for devices , error: %s", id, err)
			return nil, err
		}

		for _, val := range devices {
			obj = append(obj, &DbObjAction{
				Type:     "measurementcfg",
				TypeDesc: "Measurements",
				ObID:     val.IDMeasurementCfg,
				Action:   "Delete VarCatalogiablefield from Measurement relation",
			})

		}*/
	return obj, nil
}
