package config

import (
	"fmt"
	"strconv"
)

/***************************
	SNMPDevice Services
	-GetSnmpDeviceCfgByID(struct)
	-GetSnmpDeviceCfgMap (map - for interna config use
	-GetSnmpDeviceCfgArray(Array - for web ui use )
	-AddSnmpDeviceCfg
	-DelSnmpDeviceCfg
	-UpdateSnmpDeviceCfg
	-GeSnmpDeviceCfgAffectOnDel
***********************************/

/*GetSnmpDeviceCfgByID get device data by id*/
func (dbc *DatabaseCfg) GetSnmpDeviceCfgByID(id string) (SnmpDeviceCfg, error) {
	devcfgarray, err := dbc.GetSnmpDeviceCfgArray("id='" + id + "'")
	if err != nil {
		return SnmpDeviceCfg{}, err
	}
	if len(devcfgarray) > 1 {
		return SnmpDeviceCfg{}, fmt.Errorf("Error %d results on get SnmpDeviceCfg by id %s", len(devcfgarray), id)
	}
	if len(devcfgarray) == 0 {
		return SnmpDeviceCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the snmp device config table", id)
	}
	return *devcfgarray[0], nil
}

/*GetSnmpDeviceCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetSnmpDeviceCfgMap(filter string) (map[string]*SnmpDeviceCfg, error) {
	devcfgarray, err := dbc.GetSnmpDeviceCfgArray(filter)
	devcfgmap := make(map[string]*SnmpDeviceCfg)
	for _, val := range devcfgarray {
		devcfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return devcfgmap, err
}

func (dbc *DatabaseCfg) GetNumberOfSnmpDevices() (int, error) {
	var err error
	// var devices []*SnmpDeviceCfg
	// counts, err := dbc.x.Count(&devices)
	// counts, err := dbc.x.SQL("SELECT count(*) AS total from snmp_device_cfg").Count(&devices)
	result, err := dbc.x.QueryString("SELECT count(*) AS total from snmp_device_cfg")
	counts, _ := strconv.Atoi(result[0]["total"])
	return counts, err
}


/*GetSnmpDeviceCfgArray generate an array of devices with all its information */
func (dbc *DatabaseCfg) GetSnmpDeviceCfgArray(filter string) ([]*SnmpDeviceCfg, error) {
	var err error
	var devices []*SnmpDeviceCfg
	//Get Only data for selected devices
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get SnmpDevicesCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get SnmpDevicesCfg   data: %v\n", err)
			return nil, err
		}
	}

	//Asign Groups to devices.
	var snmpdevmgroups []*SnmpDevMGroups
	if err = dbc.x.Find(&snmpdevmgroups); err != nil {
		log.Warnf("Fail to get SnmpDevices and Measurement groups relationship data: %v\n", err)
		return devices, err
	}

	//Load Measurements and metrics relationship
	//We assign field metric ID to each measurement
	for _, mVal := range devices {
		for _, mg := range snmpdevmgroups {
			if mg.IDSnmpDev == mVal.ID {
				mVal.MeasurementGroups = append(mVal.MeasurementGroups, mg.IDMGroupCfg)
			}
		}
	}

	//Asign Filters to devices.
	var snmpdevfilters []*SnmpDevFilters
	if err = dbc.x.Find(&snmpdevfilters); err != nil {
		log.Warnf("Fail to get SnmpDevices and Filter relationship data: %v\n", err)
		return devices, err
	}

	//Load Measurements and metrics relationship
	//We assign field metric ID to each measurement
	for _, mVal := range devices {
		for _, mf := range snmpdevfilters {
			if mf.IDSnmpDev == mVal.ID {
				mVal.MeasFilters = append(mVal.MeasFilters, mf.IDFilter)
			}
		}
	}
	return devices, nil
}

/*AddSnmpDeviceCfg for adding new devices*/
func (dbc *DatabaseCfg) AddSnmpDeviceCfg(dev SnmpDeviceCfg) (int64, error) {

	var err error
	var affected, newmg, newft int64
	session := dbc.x.NewSession()
	defer session.Close()

	_, err = dbc.GetSnmpDeviceCfgByID(dev.ID)
	if err == nil { //device exist already in the database
		return dbc.UpdateSnmpDeviceCfg(dev.ID, dev)	
	}

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//Measurement Groups
	for _, mg := range dev.MeasurementGroups {

		mgstruct := SnmpDevMGroups{
			IDSnmpDev:   dev.ID,
			IDMGroupCfg: mg,
		}
		newmg, err = session.Insert(&mgstruct)
		if err != nil {
			session.Rollback()
			return 0, err
		}
	}
	//Filters
	for _, mf := range dev.MeasFilters {
		mfstruct := SnmpDevFilters{
			IDSnmpDev: dev.ID,
			IDFilter:  mf,
		}
		newft, err = session.Insert(&mfstruct)
		if err != nil {
			session.Rollback()
			return 0, err
		}
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Added new Device Successfully with id %s [%d Measurment Groups | %d filters]", dev.ID, newmg, newft)
	dbc.addChanges(affected + newmg + newft)
	return affected, nil
}

/*DelSnmpDeviceCfg for deleting devices from ID*/
func (dbc *DatabaseCfg) DelSnmpDeviceCfg(id string) (int64, error) {
	var affectedmg, affectedft, affectedcf, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	//first deleting references in SnmpDevMGroups SnmpDevFilters
	// Measurement Groups
	affectedmg, err = session.Where("id_snmpdev='" + id + "'").Delete(&SnmpDevMGroups{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete SnmpDevMGroups with id: %s, error: %s", id, err)
	}
	//Filters{}
	affectedft, err = session.Where("id_snmpdev='" + id + "'").Delete(&SnmpDevFilters{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete SnmpDevFilters with id: %s, error: %s", id, err)
	}
	//CustomFilter Reladed Dev
	affectedcf, err = session.Where("related_dev='" + id + "'").Cols("related_dev").Update(&CustomFilterCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete SnmpDevCfg with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&SnmpDeviceCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully device with ID %s [] %d Measurement Groups affected , %d Filters affected ,%d Custom filter  affected]", id, affectedmg, affectedft, affectedcf)
	dbc.addChanges(affected + affectedmg + affectedft + affectedcf)
	return affected, nil
}

/*UpdateSnmpDeviceCfg for adding new devices*/
func (dbc *DatabaseCfg) UpdateSnmpDeviceCfg(id string, dev SnmpDeviceCfg) (int64, error) {
	var deletemg, newmg, deleteft, newft, affectedcf, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()
	//Deleting first all relations
	deletemg, err = session.Where("id_snmpdev='" + id + "'").Delete(&SnmpDevMGroups{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete SnmpDevMGroups with id: %s, error: %s", id, err)
	}
	//Filters{}
	deleteft, err = session.Where("id_snmpdev='" + id + "'").Delete(&SnmpDevFilters{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Device with id on delete SnmpDevFilters with id: %s, error: %s", id, err)
	}

	affectedcf, err = session.Where("related_dev='" + id + "'").Cols("related_dev").Update(&CustomFilterCfg{RelatedDev: dev.ID})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error Update SnmpDevice id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
	}

	//Measurement Groups
	for _, mg := range dev.MeasurementGroups {
		mgstruct := SnmpDevMGroups{
			IDSnmpDev:   dev.ID,
			IDMGroupCfg: mg,
		}
		newmg, err = session.Insert(&mgstruct)
	}
	//Filters
	for _, mf := range dev.MeasFilters {
		mfstruct := SnmpDevFilters{
			IDSnmpDev: dev.ID,
			IDFilter:  mf,
		}
		newft, err = session.Insert(&mfstruct)
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
	log.Infof("Updated device constrains (old %d / new %d ) Measurement Groups", deletemg, newmg)
	log.Infof("Updated device constrains (old %d / new %d ) MFilters", deleteft, newft)
	log.Infof("Updated new Device Successfully with id %s and data:%+v", id, dev)
	dbc.addChanges(affected + deletemg + newmg + deleteft + newft + affectedcf)
	return affected, nil
}

/*GeSnmpDeviceCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GeSnmpDeviceCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var devices []*CustomFilterCfg
	var obj []*DbObjAction
	if err := dbc.x.Where("related_dev='" + id + "'").Find(&devices); err != nil {
		log.Warnf("Error on Get Custotm Filter id %s for devices , error: %s", id, err)
		return nil, err
	}
	for _, val := range devices {
		obj = append(obj, &DbObjAction{
			Type:     "customfiltercfg",
			TypeDesc: "Custom Filters",
			ObID:     val.ID,
			Action:   "Delete related Device from CustomFilter",
		})
	}
	return obj, nil
}
