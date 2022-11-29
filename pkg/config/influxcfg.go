package config

import "fmt"

/***************************
	Influx DB backends
	-GetInfluxCfgCfgByID(struct)
	-GetInfluxCfgMap (map - for interna config use
	-GetInfluxCfgArray(Array - for web ui use )
	-AddInfluxCfg
	-DelInfluxCfg
	-UpdateInfluxCfg
  -GetInfluxCfgAffectOnDel
***********************************/

/*GetInfluxCfgByID get device data by id*/
func (dbc *DatabaseCfg) GetInfluxCfgByID(id string) (InfluxCfg, error) {
	cfgarray, err := dbc.GetInfluxCfgArray("id='" + id + "'")
	if err != nil {
		return InfluxCfg{}, err
	}
	if len(cfgarray) > 1 {
		return InfluxCfg{}, fmt.Errorf("Error %d results on get InfluxCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return InfluxCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the influx config table", id)
	}
	return *cfgarray[0], nil
}

/*GetInfluxCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetInfluxCfgMap(filter string) (map[string]*InfluxCfg, error) {
	cfgarray, err := dbc.GetInfluxCfgArray(filter)
	cfgmap := make(map[string]*InfluxCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetInfluxCfgArray generate an array of devices with all its information */
func (dbc *DatabaseCfg) GetInfluxCfgArray(filter string) ([]*InfluxCfg, error) {
	var err error
	var devices []*InfluxCfg
	// Get Only data for selected devices
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get InfluxCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get influxcfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddInfluxCfg for adding new devices*/
func (dbc *DatabaseCfg) AddInfluxCfg(dev InfluxCfg) (int64, error) {
	var err error
	var affected int64
	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return 0, err
	}
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	// no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Added new influx backend Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelInfluxCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelInfluxCfg(id string) (int64, error) {
	var affectedouts, affected int64
	var err error

	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return 0, err
	}
	defer session.Close()

	affectedouts, err = session.Where("id_backend='" + id + "' and backend_type = 'influxdb'").Cols("outdb").Delete(&OutputBackends{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete SnmpDevCfg with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&InfluxCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully influx db with ID %s [ %d Outputs Affected  ]", id, affectedouts)
	dbc.addChanges(affected + affectedouts)
	return affected, nil
}

/*UpdateInfluxCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateInfluxCfg(id string, dev InfluxCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return 0, err
	}
	defer session.Close()

	if id != dev.ID { // ID has been changed
		affecteddev, err = session.Where("id_backend='" + id + "' and backend_type='influxdb'").Cols("id_backend").Update(&OutputBackends{IDBackend: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Update InfluxConfig on update id(old) %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Influx Config to %d outputs ", affecteddev)
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

	log.Infof("Updated Influx Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetInfluxCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetInfluxCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var outputs []*OutputBackends
	var obj []*DbObjAction
	if err := dbc.x.Where("id_backend='" + id + "' and backend_type = 'influxdb'").Find(&outputs); err != nil {
		log.Warnf("Error on Get Output id %s for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range outputs {
		obj = append(obj, &DbObjAction{
			Type:     "outputcfg",
			TypeDesc: "Outputs",
			ObID:     val.IDOutput,
			Action:   "Remove InfluxDB Server from Outputs",
		})
	}
	return obj, nil
}
