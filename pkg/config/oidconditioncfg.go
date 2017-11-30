package config

import "fmt"

/***************************
Oid Condition Cfg
	-GetOidConditionCfgByID(struct)
	-GetOidConditionCfgMap (map - for interna config use
	-GetOidConditionCfgArray(Array - for web ui use )
	-AddOidConditionCfg
	-DelOidConditionCfg
	-UpdateOidConditionCfg
  -GetOidConditionCfgAffectOnDel
***********************************/

/*GetOidConditionCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetOidConditionCfgByID(id string) (OidConditionCfg, error) {
	cfgarray, err := dbc.GetOidConditionCfgArray("id='" + id + "'")
	if err != nil {
		return OidConditionCfg{}, err
	}
	if len(cfgarray) > 1 {
		return OidConditionCfg{}, fmt.Errorf("Error %d results on get CustomfilterCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return OidConditionCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the filter config table", id)
	}
	return *cfgarray[0], nil
}

/*GetOidConditionCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetOidConditionCfgMap(filter string) (map[string]*OidConditionCfg, error) {
	cfgarray, err := dbc.GetOidConditionCfgArray(filter)
	cfgmap := make(map[string]*OidConditionCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetOidConditionCfgArray generate an array of metrics with all its information */
func (dbc *DatabaseCfg) GetOidConditionCfgArray(filter string) ([]*OidConditionCfg, error) {
	var err error
	var filters []*OidConditionCfg
	//Get Only data for selected metrics
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&filters); err != nil {
			log.Warnf("Fail to get OidConditionCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&filters); err != nil {
			log.Warnf("Fail to get OidConditionCfg   data: %v\n", err)
			return nil, err
		}
	}
	return filters, nil
}

/*AddOidConditionCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddOidConditionCfg(dev OidConditionCfg) (int64, error) {
	var err error
	var affected int64
	// create OidConditionCfg to check if any configuration issue found before persist to database.
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
	log.Infof("Added new OidConditionCfg Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelOidConditionCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelOidConditionCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references filter_name on Measurement Filters
	affecteddev, err = session.Where("filter_name='" + id + "'").Cols("filter_name").Update(&MeasFilterCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete OIDCondition on Measurement Filter table with id:  %s , error: %s", id, err)
	}

	// deleting references extrada on SNMP Metric on related ConditionEval
	affecteddev, err = session.Where("extradata='" + id + "' and datasrctype = 'CONDITIONEVAL'").Cols("extradata").Update(&SnmpMetricCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete OIDCondition on Metric table with id:  %s , error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&OidConditionCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Custom Filter with ID %s [ %d Items Affected  ]", id, affecteddev)
	dbc.addChanges(affecteddev)
	return affected, nil
}

/*UpdateOidConditionCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateOidConditionCfg(id string, dev OidConditionCfg) (int64, error) {
	var err error
	var affected, affecteddev int64
	// create OidConditionCfg to check if any configuration issue found before persist to database.
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		//SnmpMetricCfg
		affecteddev, err = session.Where("extradata='" + id + "' and datasrctype = 'CONDITIONEVAL'").Cols("extradata").Update(&SnmpMetricCfg{ExtraData: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Update SnmpMetricCfg on update OID Condition id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		//MeasFilterCfg
		affecteddev, err = session.Where("filter_name='" + id + "'").Cols("filter_name").Update(&MeasFilterCfg{FilterName: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Update Custom Filter on update id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
	}

	affected, err = session.Where("id='" + id + "'").UseBool().AllCols().Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Updated new OidConditionCfg Successfully with id %s [ %s id changed]", dev.ID, affecteddev)
	dbc.addChanges(affected)
	return affected, nil
}

/*GetOidConditionCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetOidConditionCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var metrics []*SnmpMetricCfg
	var measf []*MeasFilterCfg
	var obj []*DbObjAction
	var err error
	if err = dbc.x.Where("extradata='" + id + "' and datasrctype = 'CONDITIONEVAL'").Find(&metrics); err != nil {
		log.Warnf("Error on Get CustomID  id %d for Measurement Filters , error: %s", id, err)
		return nil, err
	}

	for _, val := range metrics {
		obj = append(obj, &DbObjAction{
			Type:     "snmpmetriccfg",
			TypeDesc: "Metrics",
			ObID:     val.ID,
			Action:   "Change Measurement filter to other custom or delete them",
		})
	}

	if err = dbc.x.Where("filter_name='" + id + "'").Find(&measf); err != nil {
		log.Warnf("Error on Get CustomID  id %d for Measurement Filters , error: %s", id, err)
		return nil, err
	}

	for _, val := range measf {
		obj = append(obj, &DbObjAction{
			Type:     "measfiltercfg",
			TypeDesc: "Meas. Filters",
			ObID:     val.ID,
			Action:   "Change Measurement filter to other custom or delete them",
		})
	}

	return obj, nil
}
