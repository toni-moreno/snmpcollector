package config

import "fmt"

/***************************
Custom Filter Cfg
	-GetCustomFilterCfgCfgByID(struct)
	-GetCustomFilterCfgMap (map - for interna config use
	-GetCustomFilterCfgArray(Array - for web ui use )
	-AddCustomFilterCfg
	-DelCustomFilterCfg
	-UpdateCustomFilterCfg
  -GetCustomFilterCfgAffectOnDel
***********************************/

/*GetCustomFilterCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetCustomFilterCfgByID(id string) (CustomFilterCfg, error) {
	cfgarray, err := dbc.GetCustomFilterCfgArray("id='" + id + "'")
	if err != nil {
		return CustomFilterCfg{}, err
	}
	if len(cfgarray) > 1 {
		return CustomFilterCfg{}, fmt.Errorf("Error %d results on get CustomfilterCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return CustomFilterCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the filter config table", id)
	}
	return *cfgarray[0], nil
}

/*GetCustomFilterCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetCustomFilterCfgMap(filter string) (map[string]*CustomFilterCfg, error) {
	cfgarray, err := dbc.GetCustomFilterCfgArray(filter)
	cfgmap := make(map[string]*CustomFilterCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetCustomFilterCfgArray generate an array of metrics with all its information */
func (dbc *DatabaseCfg) GetCustomFilterCfgArray(filter string) ([]*CustomFilterCfg, error) {
	var err error
	var filters []*CustomFilterCfg
	//Get Only data for selected metrics
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&filters); err != nil {
			log.Warnf("Fail to get CustomFilterCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&filters); err != nil {
			log.Warnf("Fail to get CustomFilterCfg   data: %v\n", err)
			return nil, err
		}
	}
	for k, vf := range filters {
		var item []*CustomFilterItems
		if err = dbc.x.Where("customid = '" + vf.ID + "'").Find(&item); err != nil {
			log.Warnf("Fail to get CustomFilterItems  data filtered with ID %s : %v\n", vf.ID, err)
			continue
		}
		//log.Debugf("ITEM ( %s ) %+v", vf.ID, item)
		for _, vi := range item {
			i := struct {
				TagID string
				Alias string
			}{
				TagID: vi.TagID,
				Alias: vi.Alias,
			}
			filters[k].Items = append(filters[k].Items, i)
		}
	}
	return filters, nil
}

/*AddCustomFilterCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddCustomFilterCfg(dev CustomFilterCfg) (int64, error) {
	var err error
	var affected int64
	// create CustomFilterCfg to check if any configuration issue found before persist to database.
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	// first we will remove all previous entries
	affected, err = session.Where("customid='" + dev.ID + "'").Delete(&CustomFilterItems{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Addig new filter config inputs with id on add MeasurementFieldCfg with id: %s, error: %s", dev.ID, err)
	}
	//inserting new ones
	for _, v := range dev.Items {
		affected, err = session.Insert(&CustomFilterItems{CustomID: dev.ID, TagID: v.TagID, Alias: v.Alias})
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
	log.Infof("Added new CustomFilterCfg Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelCustomFilterCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelCustomFilterCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in Measurements

	affecteddev, err = session.Where("filter_name='" + id + "'").Cols("filter_name").Update(&MeasFilterCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Custom Filter on Measurement Filter table with id:  %s , error: %s", id, err)
	}

	affecteddev, err = session.Where("customid='" + id + "'").Delete(&CustomFilterItems{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	affected, err = session.Where("id='" + id + "'").Delete(&CustomFilterCfg{})
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

/*UpdateCustomFilterCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateCustomFilterCfg(id string, dev CustomFilterCfg) (int64, error) {
	var err error
	var affected, affecteddev int64
	// create CustomFilterCfg to check if any configuration issue found before persist to database.
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed so we need to update Related MeasurementCfg
		affecteddev, err = session.Where("filter_name='" + id + "'").Cols("filter_name").Update(&MeasFilterCfg{FilterName: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Update Custom Filter on update id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
	}
	// first we will remove all previous entries
	affected, err = session.Where("customid='" + dev.ID + "'").Delete(&CustomFilterItems{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Addig new filter config inputs with id on add MeasurementFieldCfg with id: %s, error: %s", dev.ID, err)
	}
	//inserting new ones
	for _, v := range dev.Items {
		affected, err = session.Insert(&CustomFilterItems{CustomID: dev.ID, TagID: v.TagID, Alias: v.Alias})
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
	//no other relation
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Updated new CustomFilterCfg Successfully with id %s [ %s id changed]", dev.ID, affecteddev)
	dbc.addChanges(affected)
	return affected, nil
}

/*GetCustomFilterCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetCustomFilterCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var filters []*MeasFilterCfg
	var obj []*DbObjAction
	if err := dbc.x.Where("filter_name='" + id + "'").Find(&filters); err != nil {
		log.Warnf("Error on Get CustomID  id %d for Measurement Filters , error: %s", id, err)
		return nil, err
	}

	for _, val := range filters {
		obj = append(obj, &DbObjAction{
			Type:     "measfiltercfg",
			TypeDesc: "Meas. Filters",
			ObID:     val.ID,
			Action:   "Change Measurement filter to other custom or delete them",
		})

	}
	return obj, nil
}
