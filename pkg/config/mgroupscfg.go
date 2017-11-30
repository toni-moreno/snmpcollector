package config

import "fmt"

/***************************
MEASUREMENT GROUPS
	-GetMGroupsCfgCfgByID(struct)
	-GetMGroupsCfgMap (map - for interna config use
	-GetMGroupsCfgArray(Array - for web ui use )
	-AddMGroupsCfg
	-DelMGroupsCfg
	-UpdateMGroupsCfg
  -GetMGroupsCfgAffectOnDel
***********************************/

/*GetMGroupsCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetMGroupsCfgByID(id string) (MGroupsCfg, error) {
	cfgarray, err := dbc.GetMGroupsCfgArray("id='" + id + "'")
	if err != nil {
		return MGroupsCfg{}, err
	}
	if len(cfgarray) > 1 {
		return MGroupsCfg{}, fmt.Errorf("Error %d results on get MGroupsCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return MGroupsCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the meas groups config table", id)
	}
	return *cfgarray[0], nil
}

/*GetMGroupsCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetMGroupsCfgMap(filter string) (map[string]*MGroupsCfg, error) {
	cfgarray, err := dbc.GetMGroupsCfgArray(filter)
	cfgmap := make(map[string]*MGroupsCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetMGroupsCfgArray generate an array of metrics with all its information */
func (dbc *DatabaseCfg) GetMGroupsCfgArray(filter string) ([]*MGroupsCfg, error) {
	var err error
	var devices []*MGroupsCfg
	//Get Only data for selected metrics
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get MGroupsCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get MGroupsCfg   data: %v\n", err)
			return nil, err
		}
	}

	//Load measurement for each groups
	var mgroupsmeas []*MGroupsMeasurements
	if err = dbc.x.Find(&mgroupsmeas); err != nil {
		log.Warnf("Fail to get MGroup Measurements relationship  data: %v\n", err)
	}

	for _, mVal := range devices {
		for _, mgm := range mgroupsmeas {
			if mgm.IDMGroupCfg == mVal.ID {
				mVal.Measurements = append(mVal.Measurements, mgm.IDMeasurementCfg)
			}
		}
	}
	return devices, nil
}

/*AddMGroupsCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddMGroupsCfg(dev MGroupsCfg) (int64, error) {
	var err error
	var affected, newmf int64
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//Measurement Fields
	for _, meas := range dev.Measurements {

		mstruct := MGroupsMeasurements{
			IDMGroupCfg:      dev.ID,
			IDMeasurementCfg: meas,
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
	log.Infof("Added new Measurement Group Successfully with id %s  [%d Measurements]", dev.ID, newmf)
	dbc.addChanges(affected + newmf)
	return affected, nil
}

/*DelMGroupsCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelMGroupsCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in Measurements tables
	affecteddev, err = session.Where("id_mgroup_cfg='" + id + "'").Delete(&MGroupsMeasurements{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Metric with id on delete MeasurementFieldCfg with id: %s, error: %s", id, err)
	}

	//deleting all references in devices (snmpdevfilters)
	affecteddev, err = session.Where("id_mgroup_cfg='" + id + "'").Delete(&SnmpDevMGroups{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Filter on SnmpDeviceFilter table with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&MGroupsCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Measurment Group with ID %s [ %d Devices Affected  ]", id, affecteddev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*UpdateMGroupsCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateMGroupsCfg(id string, dev MGroupsCfg) (int64, error) {
	var affecteddev, newmg, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("id_mgroup_cfg='" + id + "'").Cols("id_mgroup_cfg").Update(&SnmpDevMGroups{IDMGroupCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Metric id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Measurement Group Config to %s devices ", affecteddev)
	}
	//Remove all measurements in group.
	_, err = session.Where("id_mgroup_cfg='" + id + "'").Delete(&MGroupsMeasurements{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Metric with id on delete MeasurementFieldCfg with id: %s, error: %s", id, err)
	}
	//adding again
	for _, meas := range dev.Measurements {

		mstruct := MGroupsMeasurements{
			IDMGroupCfg:      dev.ID,
			IDMeasurementCfg: meas,
		}
		newmg, err = session.Insert(&mstruct)
		if err != nil {
			session.Rollback()
			return 0, err
		}
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

	log.Infof("Updated Measurement Group Successfully with id %s [%d measurements], affected", dev.ID, newmg)
	dbc.addChanges(affected + newmg)
	return affected, nil
}

/*GetMGroupsCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetMGroupsCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var devices []*SnmpDevMGroups
	var obj []*DbObjAction
	if err := dbc.x.Where("id_mgroup_cfg='" + id + "'").Find(&devices); err != nil {
		log.Warnf("Error on Get Measrument groups id %d for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range devices {
		obj = append(obj, &DbObjAction{
			Type:     "snmpdevicecfg",
			TypeDesc: "SNMP Devices",
			ObID:     val.IDSnmpDev,
			Action:   "Delete SNMPDevice from Measurement Group relation",
		})

	}
	return obj, nil
}
