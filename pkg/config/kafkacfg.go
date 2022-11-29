package config

import "fmt"

/***************************
	Kafka DB backends
	-GetKafkaCfgCfgByID(struct)
	-GetKafkaCfgMap (map - for interna config use
	-GetKafkaCfgArray(Array - for web ui use )
	-AddKafkaCfg
	-DelKafkaCfg
	-UpdateKafkaCfg
	-GetKafkaCfgAffectOnDel
***********************************/

/*GetKafkaCfgByID get device data by id*/
func (dbc *DatabaseCfg) GetKafkaCfgByID(id string) (KafkaCfg, error) {
	cfgarray, err := dbc.GetKafkaCfgArray("id='" + id + "'")
	if err != nil {
		return KafkaCfg{}, err
	}
	if len(cfgarray) > 1 {
		return KafkaCfg{}, fmt.Errorf("Error %d results on get KafkaCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return KafkaCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the Kafka config table", id)
	}
	return *cfgarray[0], nil
}

/*GetKafkaCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetKafkaCfgMap(filter string) (map[string]*KafkaCfg, error) {
	cfgarray, err := dbc.GetKafkaCfgArray(filter)
	cfgmap := make(map[string]*KafkaCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetKafkaCfgArray generate an array of devices with all its information */
func (dbc *DatabaseCfg) GetKafkaCfgArray(filter string) ([]*KafkaCfg, error) {
	var err error
	var devices []*KafkaCfg
	// Get Only data for selected devices
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get KafkaCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get Kafkacfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddKafkaCfg for adding new devices*/
func (dbc *DatabaseCfg) AddKafkaCfg(dev KafkaCfg) (int64, error) {
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
	log.Infof("Added new Kafka backend Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelKafkaCfg for deleting Kafka databases from ID*/
func (dbc *DatabaseCfg) DelKafkaCfg(id string) (int64, error) {
	var affectedouts, affected int64
	var err error

	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return 0, err
	}
	defer session.Close()
	// deleting references in SnmpDevCfg

	affectedouts, err = session.Where("id_backend='" + id + "' and backend_type = 'kafka'").Cols("outdb").Delete(&OutputBackends{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete SnmpDevCfg with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&KafkaCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Kafka db with ID %s [ %d Outputs Affected  ]", id, affectedouts)
	dbc.addChanges(affected + affectedouts)
	return affected, nil
}

/*UpdateKafkaCfg for adding new Kafka*/
func (dbc *DatabaseCfg) UpdateKafkaCfg(id string, dev KafkaCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	if err := session.Begin(); err != nil {
		// if returned then will rollback automatically
		return 0, err
	}
	defer session.Close()
	if id != dev.ID { // ID has been changed
		affecteddev, err = session.Where("id_backend='" + id + "' and backend_type='kafka'").Cols("id_backend").Update(&OutputBackends{IDBackend: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Update KafkaConfig on update id(old) %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Kafka Config to %d outputs ", affecteddev)
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

	log.Infof("Updated Kafka Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetKafkaCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetKafkaCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var outputs []*OutputBackends
	var obj []*DbObjAction
	if err := dbc.x.Where("id_backend='" + id + "' and backend_type = 'kafka'").Find(&outputs); err != nil {
		log.Warnf("Error on Get Output id %s for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range outputs {
		obj = append(obj, &DbObjAction{
			Type:     "outputcfg",
			TypeDesc: "Outputs",
			ObID:     val.IDOutput,
			Action:   "Remove Kafka Server from Outputs",
		})
	}
	return obj, nil
}
