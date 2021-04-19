package config

import (
	"fmt"
	"strings"
)

/***************************
Global Var
	-GetPollerLocationCfgCfgByLocation(struct)
	-GetPollerLocationCfgMap (map - for interna config use
	-GetPollerLocationCfgArray(Array - for web ui use )
	-AddPollerLocationCfg
	-DelPollerLocationCfg
	-UpdatePollerLocationCfg
  -GetPollerLocationCfgAffectOnDel
***********************************/

/*GetPollerLocationCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetPollerLocationCfgByID(id string) (PollerLocationCfg, error) {
	cfgarray, err := dbc.GetPollerLocationCfgArray("id='" + id + "'")
	if err != nil {
		return PollerLocationCfg{}, err
	}
	if len(cfgarray) > 1 {
		return PollerLocationCfg{}, fmt.Errorf("Error %d results on get PollerLocationCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return PollerLocationCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the Global Var config table", id)
	}
	return *cfgarray[0], nil
}

/*GetPollerLocationCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetPollerLocationCfgMap(filter string) (map[string]*PollerLocationCfg, error) {
	cfgarray, err := dbc.GetPollerLocationCfgArray(filter)
	cfgmap := make(map[string]*PollerLocationCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		//log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetPollerLocationCfgArray generate an array of metrics with all its information */
func (dbc *DatabaseCfg) GetPollerLocationCfgArray(filter string) ([]*PollerLocationCfg, error) {
	var err error
	var devices []*PollerLocationCfg
	//Get Only data for selected metrics
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get PollerLocationCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get PollerLocationCfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddPollerLocationCfg for adding new Global Variable*/
func (dbc *DatabaseCfg) AddPollerLocationCfg(dev PollerLocationCfg) (int64, error) {
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

/*DelPollerLocationCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelPollerLocationCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in Measurements

	affected, err = session.Where("id='" + id + "'").Delete(&PollerLocationCfg{})
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

/*UpdatePollerLocationCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdatePollerLocationCfg(id string, dev PollerLocationCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	// create PollerLocationCfg to check if any configuration issue found before persist to database.

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
		log.Infof("Updated PollerLocationiableConfig to %d devices ", affecteddev)
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

	log.Infof("Updated PollerLocationiableConfig Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetPollerLocationCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetPollerLocationCfgAffectOnDel(id string) ([]*DbObjAction, error) {
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
				Action:   "Delete PollerLocationiablefield from Measurement relation",
			})

		}*/
	return obj, nil
}
