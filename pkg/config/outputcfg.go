package config

import "fmt"

/***************************
	Kafka DB backends
	-GetOutputCfgCfgByID(struct)
	-GetOutputCfgMap (map - for interna config use
	-GetOutputCfgArray(Array - for web ui use )
	-AddOutputCfg
	-DelOutputCfg
	-UpdateOutputCfg
	-GetOutputCfgAffectOnDel
***********************************/

/*GetOutputCfgByID get device data by id*/
func (dbc *DatabaseCfg) GetOutputCfgByID(id string) (OutputCfg, error) {
	cfgarray, err := dbc.GetOutputCfgArray("id='" + id + "'")
	if err != nil {
		return OutputCfg{}, err
	}
	if len(cfgarray) > 1 {
		return OutputCfg{}, fmt.Errorf("Error %d results on get SnmpDeviceCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return OutputCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the Kafka config table", id)
	}
	return *cfgarray[0], nil
}

/*GetOutputCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetOutputCfgMap(filter string) (map[string]*OutputCfg, error) {
	cfgarray, err := dbc.GetOutputCfgArray(filter)
	cfgmap := make(map[string]*OutputCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetOutputCfgArray generate an array of devices with all its information */
func (dbc *DatabaseCfg) GetOutputCfgArray(filter string) ([]*OutputCfg, error) {
	var err error
	var devices []*OutputCfg
	// Get Only data for selected devices
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get OutputCfg data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get OutputCfg data: %v\n", err)
			return nil, err
		}
	}

	// Asign Groups to devices.
	var outputkafkabackends []*OutputBackends
	if err = dbc.x.Find(&outputkafkabackends); err != nil {
		log.Warnf("Fail to get Output and Backends relationship data: %v\n", err)
		return devices, err
	}

	for _, kVal := range devices {
		for _, bk := range outputkafkabackends {
			if bk.IDOutput == kVal.ID {
				kVal.Backend = bk.IDBackend
				kVal.BackendType = bk.BackendType
			}
		}
	}

	// Load backends based on type

	return devices, nil
}

/*AddOutputCfg for adding new devices*/
func (dbc *DatabaseCfg) AddOutputCfg(dev OutputCfg) (int64, error) {
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

	// Backends
	obktruct := OutputBackends{
		IDOutput:    dev.ID,
		IDBackend:   dev.Backend,
		BackendType: dev.BackendType,
	}
	newback, err := session.Insert(&obktruct)
	if err != nil {
		session.Rollback()
		return 0, err
	}

	// no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Infof("Added new Output backend successfully with id %s ", dev.ID)
	dbc.addChanges(affected + newback)
	return affected, nil
}

/*DelOutputCfg for deleting Kafka databases from ID*/
func (dbc *DatabaseCfg) DelOutputCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return 0, err
	}
	defer session.Close()
	// deleting references in SnmpDevCfg

	affecteddev, err = session.Where("outdb='" + id + "'").Cols("outdb").Update(&SnmpDeviceCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("error on delete Device with id on delete SnmpDevCfg with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&OutputCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Kafka db with ID %s [ %d Devices Affected  ]", id, affecteddev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*UpdateOutputCfg for adding new Kafka*/
func (dbc *DatabaseCfg) UpdateOutputCfg(id string, dev OutputCfg) (int64, error) {
	var affectedouts, affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return 0, err
	}
	defer session.Close()

	if id != dev.ID { // ID has been changed
		affecteddev, err = session.Where("outdb='" + id + "'").Cols("outdb").Update(&SnmpDeviceCfg{OutDB: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("error Update OutputBackend id(old) %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Outputs to %d SNMPDevices ", affecteddev)
	}

	// Remove all outputs in group.
	_, err = session.Where("id_output='" + id + "'").Delete(&OutputBackends{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("error on delete old output with id: %s, error: %s", id, err)
	}
	// adding again
	// Backends
	obktruct := OutputBackends{
		IDOutput:    dev.ID,
		IDBackend:   dev.Backend,
		BackendType: dev.BackendType,
	}
	newback, err := session.Insert(&obktruct)
	if err != nil {
		session.Rollback()
		return 0, err
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

	log.Infof("Updated Output Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affectedouts + affecteddev + newback)
	return affected, nil
}

/*GetOutputCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetOutputCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var devices []*SnmpDeviceCfg
	var obj []*DbObjAction
	if err := dbc.x.Where("outdb='" + id + "'").Find(&devices); err != nil {
		log.Warnf("Error on Get Outout db id %s for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range devices {
		obj = append(obj, &DbObjAction{
			Type:     "snmpdevicecfg",
			TypeDesc: "SNMP Devices",
			ObID:     val.ID,
			Action:   "Remove Output from SNMPDevice",
		})
	}
	return obj, nil
}
