package config

import (
	"fmt"
	// _ needed to mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	// _ needed to sqlite3
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync/atomic"
)

func (dbc *DatabaseCfg) resetChanges() {
	atomic.StoreInt64(&dbc.numChanges, 0)
}

func (dbc *DatabaseCfg) addChanges(n int64) {
	atomic.AddInt64(&dbc.numChanges, n)
}
func (dbc *DatabaseCfg) getChanges() int64 {
	return atomic.LoadInt64(&dbc.numChanges)
}

//DbObjAction measurement groups to asign to devices
type DbObjAction struct {
	Type     string
	TypeDesc string
	ObID     string
	Action   string
}

//InitDB initialize de BD configuration
func (dbc *DatabaseCfg) InitDB() {
	// Create ORM engine and database
	var err error
	var dbtype string
	var datasource string

	log.Debugf("Database config: %+v", dbc)

	switch dbc.Type {
	case "sqlite3":
		dbtype = "sqlite3"
		datasource = dataDir + "/" + dbc.Name + ".db"
	case "mysql":
		dbtype = "mysql"
		datasource = dbc.User + ":" + dbc.Pass + "@" + dbc.Host + "/" + dbc.Name + "?charset=utf8"
	default:
		log.Errorf("unknown db  type %s", dbc.Type)
		return
	}

	dbc.x, err = xorm.NewEngine(dbtype, datasource)
	if err != nil {
		log.Fatalf("Fail to create engine: %v\n", err)
	}

	if len(dbc.SQLLogFile) != 0 {
		dbc.x.ShowSQL(true)
		f, error := os.Create(logDir + "/" + dbc.SQLLogFile)
		if err != nil {
			log.Errorln("Fail to create log file  ", error)
		}
		dbc.x.SetLogger(xorm.NewSimpleLogger(f))
	}
	if dbc.Debug == "true" {
		dbc.x.Logger().SetLevel(core.LOG_DEBUG)
	}

	// Sync tables
	if err = dbc.x.Sync(new(InfluxCfg)); err != nil {
		log.Fatalf("Fail to sync database InfluxCfg: %v\n", err)
	}
	if err = dbc.x.Sync(new(SnmpDeviceCfg)); err != nil {
		log.Fatalf("Fail to sync database SnmpDeviceCfg: %v\n", err)
	}
	if err = dbc.x.Sync(new(SnmpMetricCfg)); err != nil {
		log.Fatalf("Fail to sync database SnmpMetricCfg: %v\n", err)
	}
	if err = dbc.x.Sync(new(MeasurementCfg)); err != nil {
		log.Fatalf("Fail to sync database MeasurementCfg: %v\n", err)
	}
	if err = dbc.x.Sync(new(MeasFilterCfg)); err != nil {
		log.Fatalf("Fail to sync database MeasurementFilterCfg : %v\n", err)
	}
	if err = dbc.x.Sync(new(MeasurementFieldCfg)); err != nil {
		log.Fatalf("Fail to sync database MeasurementFieldCfg: %v\n", err)
	}
	if err = dbc.x.Sync(new(MGroupsCfg)); err != nil {
		log.Fatalf("Fail to sync database MGroupCfg: %v\n", err)
	}
	if err = dbc.x.Sync(new(MGroupsMeasurements)); err != nil {
		log.Fatalf("Fail to sync database MGroupsMeasurements: %v\n", err)
	}
	if err = dbc.x.Sync(new(SnmpDevMGroups)); err != nil {
		log.Fatalf("Fail to sync database SnmpDevMGroups: %v\n", err)
	}
	if err = dbc.x.Sync(new(SnmpDevFilters)); err != nil {
		log.Fatalf("Fail to sync database SnmpDevFilters: %v\n", err)
	}
	if err = dbc.x.Sync(new(CustomFilterCfg)); err != nil {
		log.Fatalf("Fail to sync database CustomFilterCfg: %v\n", err)
	}
	if err = dbc.x.Sync(new(CustomFilterItems)); err != nil {
		log.Fatalf("Fail to sync database CustomFilterItems: %v\n", err)
	}
	if err = dbc.x.Sync(new(OidConditionCfg)); err != nil {
		log.Fatalf("Fail to sync database OidConditionCfg: %v\n", err)
	}
}

//LoadDbConfig get data from database
func (dbc *DatabaseCfg) LoadDbConfig(cfg *SQLConfig) {
	var err error

	//Load Influxdb databases
	cfg.Influxdb, err = dbc.GetInfluxCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Influx db's :%v", err)
	}

	//Load metrics
	cfg.Metrics, err = dbc.GetSnmpMetricCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Metrics  :%v", err)
	}

	//Load Measurements
	cfg.Measurements, err = dbc.GetMeasurementCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Measurements  :%v", err)
	}

	//Load Measurement Filters
	cfg.MFilters, err = dbc.GetMeasFilterCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Measurement Filters  :%v", err)
	}

	//Load measourement Groups

	cfg.GetGroups, err = dbc.GetMGroupsCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Measurements Groups  :%v", err)
	}

	//Device

	cfg.SnmpDevice, err = dbc.GetSnmpDeviceCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get SnmpDeviceConf :%v", err)
	}
	dbc.resetChanges()
}

/***************************
SNMP Metric
	-GetSnmpMetricCfgCfgByID(struct)
	-GetSnmpMetricCfgMap (map - for interna config use
	-GetSnmpMetricCfgArray(Array - for web ui use )
	-AddSnmpMetricCfg
	-DelSnmpMetricCfg
	-UpdateSnmpMetricCfg
  -GetSnmpMetricCfgAffectOnDel
***********************************/

/*GetSnmpMetricCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetSnmpMetricCfgByID(id string) (SnmpMetricCfg, error) {
	cfgarray, err := dbc.GetSnmpMetricCfgArray("id='" + id + "'")
	if err != nil {
		return SnmpMetricCfg{}, err
	}
	if len(cfgarray) > 1 {
		return SnmpMetricCfg{}, fmt.Errorf("Error %d results on get SnmpMetricCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return SnmpMetricCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the metric config table", id)
	}
	return *cfgarray[0], nil
}

/*GetSnmpMetricCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetSnmpMetricCfgMap(filter string) (map[string]*SnmpMetricCfg, error) {
	cfgarray, err := dbc.GetSnmpMetricCfgArray(filter)
	cfgmap := make(map[string]*SnmpMetricCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetSnmpMetricCfgArray generate an array of metrics with all its information */
func (dbc *DatabaseCfg) GetSnmpMetricCfgArray(filter string) ([]*SnmpMetricCfg, error) {
	var err error
	var devices []*SnmpMetricCfg
	//Get Only data for selected metrics
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get SnmpMetricCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get SnmpMetricCfg   data: %v\n", err)
			return nil, err
		}
	}
	return devices, nil
}

/*AddSnmpMetricCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddSnmpMetricCfg(dev SnmpMetricCfg) (int64, error) {
	var err error
	var affected int64
	// create SnmpMetricCfg to check if any configuration issue found before persist to database.
	err = dev.Init()
	if err != nil {
		return 0, err
	}
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
	log.Infof("Added new Snmp Metric Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelSnmpMetricCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelSnmpMetricCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in Measurements

	affecteddev, err = session.Where("id_metric_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Metric with id on delete MeasurementFieldCfg with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&SnmpMetricCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Metricdb with ID %s [ %d Measurements Affected  ]", id, affecteddev)
	dbc.addChanges(affecteddev)
	return affected, nil
}

/*UpdateSnmpMetricCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateSnmpMetricCfg(id string, dev SnmpMetricCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	// create SnmpMetricCfg to check if any configuration issue found before persist to database.
	err = dev.Init()
	if err != nil {
		return 0, err
	}
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("id_metric_cfg='" + id + "'").Cols("id_metric_cfg").Update(&MeasurementFieldCfg{IDMetricCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Metric id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated SnmpMetric Config to %s devices ", affecteddev)
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

	log.Infof("Updated SnmpMetric Config Successfully with id %s and data:%+v, affected", id, dev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*GetSnmpMetricCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetSnmpMetricCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var devices []*MeasurementFieldCfg
	var obj []*DbObjAction
	if err := dbc.x.Where("id_metric_cfg='" + id + "'").Find(&devices); err != nil {
		log.Warnf("Error on Get Snmp Metric Cfg id %d for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range devices {
		obj = append(obj, &DbObjAction{
			Type:     "measurementcfg",
			TypeDesc: "Measurements",
			ObID:     val.IDMeasurementCfg,
			Action:   "Delete SNMPMetric field from Measurement relation",
		})

	}
	return obj, nil
}

/***************************
	MEASUREMENTS
	-GetMeasurementCfgByID(struct)
	-GetMeasurementCfgMap (map - for interna config use
	-GetMeasurementCfgArray(Array - for web ui use )
	-AddMeasurementCfg
	-DelMeasurementCfg
	-UpdateMeasurementCfg
  -GetMeasurementCfgAffectOnDel
***********************************/

/*GetMeasurementCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetMeasurementCfgByID(id string) (MeasurementCfg, error) {
	cfgarray, err := dbc.GetMeasurementCfgArray("id='" + id + "'")
	if err != nil {
		return MeasurementCfg{}, err
	}
	if len(cfgarray) > 1 {
		return MeasurementCfg{}, fmt.Errorf("Error %d results on get MeasurementCfg by id %s", len(cfgarray), id)
	}
	if len(cfgarray) == 0 {
		return MeasurementCfg{}, fmt.Errorf("Error no values have been returned with this id %s in the measurement config table", id)
	}
	return *cfgarray[0], nil
}

/*GetMeasurementCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetMeasurementCfgMap(filter string) (map[string]*MeasurementCfg, error) {
	cfgarray, err := dbc.GetMeasurementCfgArray(filter)
	cfgmap := make(map[string]*MeasurementCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetMeasurementCfgArray generate an array of measurements with all its information */
func (dbc *DatabaseCfg) GetMeasurementCfgArray(filter string) ([]*MeasurementCfg, error) {
	var err error
	var devices []*MeasurementCfg
	//Get Only data for selected measurements
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get MeasurementCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get MeasurementCfg   data: %v\n", err)
			return nil, err
		}
	}

	var MeasureMetric []*MeasurementFieldCfg
	if err = dbc.x.Find(&MeasureMetric); err != nil {
		log.Warnf("Fail to get Measurements Metric relationship data: %v\n", err)
	}

	//Load Measurements and metrics relationship
	//We asign field metric ID to each measurement
	for _, mVal := range devices {
		for _, mm := range MeasureMetric {
			if mm.IDMeasurementCfg == mVal.ID {
				data := struct {
					ID     string
					Report int
				}{
					mm.IDMetricCfg,
					mm.Report,
				}
				mVal.Fields = append(mVal.Fields, data)
			}
		}
	}
	return devices, nil
}

/*AddMeasurementCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddMeasurementCfg(dev MeasurementCfg) (int64, error) {
	var err error
	var affected, newmf int64

	// create SnmpMetricCfg to check if any configuration issue found before persist to database
	// We need to get data from database
	cfg, _ := dbc.GetSnmpMetricCfgMap("")
	err = dev.Init(&cfg)
	if err != nil {
		return 0, err
	}
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	affected, err = session.Insert(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	//Measurement Fields
	for _, metric := range dev.Fields {

		mstruct := MeasurementFieldCfg{
			IDMeasurementCfg: dev.ID,
			IDMetricCfg:      metric.ID,
			Report:           metric.Report,
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
	log.Infof("Added new Measurement Successfully with id %s and [%d Fields] ", dev.ID, newmf)
	dbc.addChanges(affected + newmf)
	return affected, nil
}

/*DelMeasurementCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelMeasurementCfg(id string) (int64, error) {
	var affectedfl, affectedmg, affectedft, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in MeasurementFieldCfg
	affectedfl, err = session.Where("id_measurement_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement on MeasurementFieldCfg with id: %s, error: %s", id, err)
	}

	affectedmg, err = session.Where("id_measurement_cfg='" + id + "'").Delete(&MGroupsMeasurements{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	affectedft, err = session.Where("id_measurement_cfg='" + id + "'").Cols("id_measurement_cfg").Update(&MeasFilterCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Update FilterMeasurement on with id: %s, error: %s", id, err)
	}

	affected, err = session.Where("id='" + id + "'").Delete(&MeasurementCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Measurement with ID %s [ %d Measurements Groups Affected / %d Fields Affected / %d Filters Afected ]", id, affectedmg, affectedfl, affectedft)
	dbc.addChanges(affected + affectedmg + affectedfl + affectedft)
	return affected, nil
}

/*UpdateMeasurementCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateMeasurementCfg(id string, dev MeasurementCfg) (int64, error) {
	var affecteddev, newmf, affected int64
	var err error
	// create SnmpMetricCfg to check if any configuration issue found before persist to database.
	// config should be got from database
	// TODO: filter only metrics in Measurement to test if measurement was well defined
	cfg, _ := dbc.GetSnmpMetricCfgMap("")
	err = dev.Init(&cfg)
	if err != nil {
		return 0, err
	}
	// initialize data persistence
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		log.Infof("Updated Measurement Config to %s devices ", affecteddev)

		affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Cols("id_measurement_cfg").Update(&MGroupsMeasurements{IDMeasurementCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Measurement id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Cols("id_measurement_cfg").Update(&MeasFilterCfg{IDMeasurementCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Measurement id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Measurement config to %s devices ", affecteddev)
	}
	//delete all previous values
	affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	//Creating nuew Measurement Fields
	for _, metric := range dev.Fields {

		mstruct := MeasurementFieldCfg{
			IDMeasurementCfg: dev.ID,
			IDMetricCfg:      metric.ID,
			Report:           metric.Report,
		}
		newmf, err = session.Insert(&mstruct)
		if err != nil {
			session.Rollback()
			return 0, err
		}
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

	log.Infof("Updated Influx Measurement Config Successfully with id %s and  (%d previous / %d new Fields), affected", id, affecteddev, newmf)
	dbc.addChanges(affecteddev + newmf)
	return affected, nil
}

/*GetMeasurementCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetMeasurementCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var mf []*MeasurementFieldCfg
	var mg []*MGroupsMeasurements
	var obj []*DbObjAction
	var err error
	err = dbc.x.Where("id_measurement_cfg='" + id + "'").Find(&mf)
	if err != nil {
		return nil, fmt.Errorf("Error on Delete Measurement on MeasurementFieldCfg with id: %s, error: %s", id, err)
	}
	for _, val := range mf {
		obj = append(obj, &DbObjAction{
			Type:     "snmpmetriccfg",
			TypeDesc: "Metrics",
			ObID:     val.IDMetricCfg,
			Action:   "Delete SNMPMetric field from Measurement relation",
		})
	}

	err = dbc.x.Where("id_measurement_cfg='" + id + "'").Find(&mg)
	if err != nil {
		return nil, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	for _, val := range mg {
		obj = append(obj, &DbObjAction{
			Type:     "measgroupscfg",
			TypeDesc: "Meas. Groups",
			ObID:     val.IDMGroupCfg,
			Action:   "Delete Measurement from Measurement Group relation",
		})
	}
	return obj, nil
}

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
	//We asign field metric ID to each measurement
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
	//We asign field metric ID to each measurement
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
	var affectedmg, affectedft, affected int64
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

	affected, err = session.Where("id='" + id + "'").Delete(&SnmpDeviceCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully device with ID %s [] %d Measurement Groups Affected , %d Filters Afected ]", id, affectedmg, affectedft)
	dbc.addChanges(affected + affectedmg + affectedft)
	return affected, nil
}

/*UpdateSnmpDeviceCfg for adding new devices*/
func (dbc *DatabaseCfg) UpdateSnmpDeviceCfg(id string, dev SnmpDeviceCfg) (int64, error) {
	var deletemg, newmg, deleteft, newft, affected int64
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
	dbc.addChanges(affected + deletemg + newmg + deleteft + newft)
	return affected, nil
}

/*GeSnmpDeviceCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GeSnmpDeviceCfgAffectOnDel(id string) ([]*DbObjAction, error) {
	var obj []*DbObjAction
	return obj, nil
}

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
		return InfluxCfg{}, fmt.Errorf("Error %d results on get SnmpDeviceCfg by id %s", len(cfgarray), id)
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
	//Get Only data for selected devices
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
	log.Infof("Added new influx backend Successfully with id %s ", dev.ID)
	dbc.addChanges(affected)
	return affected, nil
}

/*DelInfluxCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelInfluxCfg(id string) (int64, error) {
	var affecteddev, affected int64
	var err error

	session := dbc.x.NewSession()
	defer session.Close()
	// deleting references in SnmpDevCfg

	affecteddev, err = session.Where("outdb='" + id + "'").Cols("outdb").Update(&SnmpDeviceCfg{})
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
	log.Infof("Deleted Successfully influx db with ID %s [ %d Devices Affected  ]", id, affecteddev)
	dbc.addChanges(affected + affecteddev)
	return affected, nil
}

/*UpdateInfluxCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateInfluxCfg(id string, dev InfluxCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()
	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("outdb='" + id + "'").Cols("outdb").Update(&SnmpDeviceCfg{OutDB: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Update InfluxConfig on update id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Infof("Updated Influx Config to %s devices ", affecteddev)
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
	var devices []*SnmpDeviceCfg
	var obj []*DbObjAction
	if err := dbc.x.Where("outdb='" + id + "'").Find(&devices); err != nil {
		log.Warnf("Error on Get Outout db id %d for devices , error: %s", id, err)
		return nil, err
	}

	for _, val := range devices {
		obj = append(obj, &DbObjAction{
			Type:     "snmpdevicecfg",
			TypeDesc: "SNMP Devices",
			ObID:     val.ID,
			Action:   "Reset InfluxDB Server from SNMPDevice to 'default' InfluxDB Server",
		})

	}
	return obj, nil
}

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
