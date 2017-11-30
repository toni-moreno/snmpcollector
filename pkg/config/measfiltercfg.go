package config

import "fmt"

/***************************
	MEASUREMENT FILTERS
	-GetMeasFilterCfgByID(struct)
	-GetMeasFilterCfgMap (map - for interna config use
	-GetMeasFilterCfgArray(Array - for web ui use )
	-AddMeasFilterCfg
	-DelMeasFilterCfg
	-UpdateMeasFilterCfg
  -GetMeasFilterCfgAffectOnDel
***********************************/

/*GetMeasFilterCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetMeasFilterCfgByID(id string) (MeasFilterCfg, error) {
	cfgarray, err := dbc.GetMeasFilterCfgArray("id='" + id + "'")
	if err != nil {
		return MeasFilterCfg{}, err
	}
	if len(cfgarray) > 1 {
		return MeasFilterCfg{}, fmt.Errorf("Error %d results on get MeasurementFilter by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return MeasFilterCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the meas filter config table", id)
	}
	return *cfgarray[0], nil
}

/*GetMeasFilterCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetMeasFilterCfgMap(filter string) (map[string]*MeasFilterCfg, error) {
	cfgarray, err := dbc.GetMeasFilterCfgArray(filter)
	cfgmap := make(map[string]*MeasFilterCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetMeasFilterCfgArray generate an array of measurements with all its information */
func (dbc *DatabaseCfg) GetMeasFilterCfgArray(filter string) ([]*MeasFilterCfg, error) {
	var err error
	var devices []*MeasFilterCfg
	//Get Only data for selected measurements
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get MeasFilterCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get MeasFilterCfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddMeasFilterCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddMeasFilterCfg(dev MeasFilterCfg) (int64, error) {
	var err error
	var affected int64
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//here we should also add file if this is a file filter

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Added new Measurement Filter Successfully with id %s  ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelMeasFilterCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelMeasFilterCfg(id string) (int64, error) {
	var affectedfl, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in SnmpDeviceCfg
	affectedfl, err = session.Where("id_filter='" + id + "'").Delete(&SnmpDevFilters{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Filter on SnmpDeviceFilter table with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&MeasFilterCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Measurement Filter with ID %s [ %d Devices Affected  ]", id, affectedfl)
	dbc.addChanges(affected + affectedfl)
	return affected, nil
}

/*UpdateMeasFilterCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateMeasFilterCfg(id string, dev MeasFilterCfg) (int64, error) {
	var affecteddev, newmf, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed only need change id's in snsmpdev
		affecteddev, err = session.Where("id_filter='" + id + "'").Cols("id_filter").Update(&SnmpDevFilters{IDFilter: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Filter id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Measurement Filter Config to %s devices ", affecteddev)
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

	log.Infof("Updated Measurement Filter Config Successfully with id %s and  (%d previous / %d new Fields), affected", id, affecteddev, newmf)
	dbc.addChanges(affected + affecteddev + newmf)
	return affected, nil
}

/*GetMeasFilterCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetMeasFilterCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var mf []*SnmpDevFilters
	var obj []*DbObjAction
	var err error
	err = dbc.x.Where("id_filter='" + id + "'").Find(&mf)
	if err != nil {
		return nil, fmt.Errorf("Error on Delete Measurement filter with id: %s, error: %s", id, err)
	}
	for _, val := range mf {
		obj = append(obj, &DbObjAction{
			Type:     "snmpdevicecfg",
			TypeDesc: "SNMP Devices",
			ObID:     val.IDSnmpDev,
			Action:   "Delete Measurement Filter in SNMPDevices relation",
		})
	}
	return obj, nil
}
