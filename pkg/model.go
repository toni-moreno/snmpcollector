package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

// InfluxCfg is the main configuration for any InfluxDB TSDB
type InfluxCfg struct {
	ID        string `xorm:"'id' unique"`
	Host      string `xorm:"host"`
	Port      int    `xorm:"port"`
	DB        string `xorm:"db"`
	User      string `xorm:"user"`
	Password  string `xorm:"password"`
	Retention string `xorm:"retention"`
}

// SnmpDeviceCfg contains all snmp related device definitions
type SnmpDeviceCfg struct {
	ID string `xorm:"'id' unique"`
	//snmp connection config
	Host    string `xorm:"host"`
	Port    int    `xorm:"port"`
	Retries int    `xorm:"retries"`
	Timeout int    `xorm:"timeout"`
	Repeat  int    `xorm:"repeat"`
	//snmp auth  config
	SnmpVersion string `xorm:"snmpversion"`
	Community   string `xorm:"community"`
	V3SecLevel  string `xorm:"v3seclevel"`
	V3AuthUser  string `xorm:"v3authuser"`
	V3AuthPass  string `xorm:"v3authpass"`
	V3AuthProt  string `xorm:"v3authprot"`
	V3PrivPass  string `xorm:"v3privpass"`
	V3PrivProt  string `xorm:"v3privprot"`
	//snmp runtime config
	Freq int `xorm:"freq"`

	OutDB    string `xorm:"outdb"`
	LogLevel string `xorm:"loglevel"`
	LogFile  string `xorm:"logfile"`

	SnmpDebug bool `xorm:"snmpdebug"`
	//influx tags
	DeviceTagName  string   `xorm:"devicetagname"`
	DeviceTagValue string   `xorm:"devicetagvalue"`
	ExtraTags      []string `xorm:"extra-tags"`

	//Filters for measurements
	MeasurementGroups []string `xorm:"-"`
	MeasFilters       []string `xorm:"-"`
}

//SnmpMetricCfg Metric config
type SnmpMetricCfg struct {
	ID          string  `xorm:"'id' unique"` //name of the key in the config array
	FieldName   string  `xorm:"field_name"`
	Description string  `xorm:"description"`
	BaseOID     string  `xorm:"baseoid"`
	DataSrcType string  `xorm:"datasrctype"`
	GetRate     bool    `xorm:"getrate"` //ony Valid with COUNTERS
	Scale       float64 `xorm:"scale"`   //only valid with gauge/integer
	Shift       float64 `xorm:"shift"`
	IsTag       bool    `xorm:"istag"`
}

//InfluxMeasurementCfg the measurement configuration
type InfluxMeasurementCfg struct {
	ID   string `xorm:"'id' unique"`
	Name string `xorm:"name"`

	GetMode     string           `xorm:"getmode"` //0=value 1=indexed
	IndexOID    string           `xorm:"indexoid"`
	IndexTag    string           `xorm:"indextag"`
	Fields      []string         `xorm:"-"` //Got from MeasurementFieldCfg table
	fieldMetric []*SnmpMetricCfg `xorm:"-"`
}

//MeasurementFieldCfg the metrics contained on each measurement (to initialize on the fieldMetric array)
type MeasurementFieldCfg struct {
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
	IDMetricCfg      string `xorm:"id_metric_cfg"`
}

//MeasFilterCfg the filter configuration
type MeasFilterCfg struct {
	ID               string `xorm:"'id' unique"`
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
	FType            string `xorm:"filter_type"`  //file/OIDCondition
	FileName         string `xorm:"file_name"`    //only vaid if file
	EnableAlias      bool   `xorm:"enable_alias"` //only valid if file
	OIDCond          string `xorm:"cond_oid"`
	CondType         string `xorm:"cond_type"`
	CondValue        string `xorm:"cond_value"`
}

//SnmpDevFilters filters to use with indexed measurement
type SnmpDevFilters struct {
	IDSnmpDev string `xorm:"id_snmpdev"`
	IDFilter  string `xorm:"id_filter"`
}

//MGroupsCfg measurement groups to asign to devices
type MGroupsCfg struct {
	ID           string   `xorm:"'id' unique"`
	Measurements []string `xorm:"-"`
}

//MGroupsMeasurements measurements contained on each Measurement Group
type MGroupsMeasurements struct {
	IDMGroupCfg      string `xorm:"id_mgroup_cfg"`
	IDMeasurementCfg string `xorm:"id_measurement_cfg"`
}

//SnmpDevMGroups Mgroups defined on each SnmpDevice
type SnmpDevMGroups struct {
	IDSnmpDev   string `xorm:"id_snmpdev"`
	IDMGroupCfg string `xorm:"id_mgroup_cfg"`
}

//DatabaseCfg de configuration for the database
type DatabaseCfg struct {
	Type       string `toml:"type"`
	Host       string `toml:"host"`
	Name       string `toml:"name"`
	User       string `toml:"user"`
	Pass       string `toml:"password"`
	SQLLogFile string `toml:"sqllogfile"`
	Debug      string `toml:"debug"`
	x          *xorm.Engine
}

//DbObjAction measurement groups to asign to devices
type DbObjAction struct {
	Type   string
	obID   string
	action string
}

//InitDB initialize de BD configuration
func InitDB(dbc *DatabaseCfg) {
	// Create ORM engine and database
	var err error
	var dbtype string
	var datasource string

	log.Debugf("Database config: %+v", dbc)

	switch dbc.Type {
	case "sqlite3":
		dbtype = "sqlite3"
		datasource = confDir + "/" + dbc.Name + ".db"
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
		f, error := os.Create(appdir + "/log/" + dbc.SQLLogFile)
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
	if err = dbc.x.Sync(new(InfluxMeasurementCfg)); err != nil {
		log.Fatalf("Fail to sync database InfluxMeasurementCfg: %v\n", err)
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
	cfgarray, err := cfg.Database.GetSnmpMetricCfgArray("id='" + id + "'")
	if err != nil {
		return SnmpMetricCfg{}, err
	}
	if len(cfgarray) > 1 {
		return SnmpMetricCfg{}, fmt.Errorf("Error %d results on get SnmpMetricCfg by id %s", len(cfgarray), id)
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
	log.Info("Added new Snmp Metric Successfully with id %s ", dev.ID)
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
	return affected, nil
}

/*UpdateSnmpMetricCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateSnmpMetricCfg(id string, dev SnmpMetricCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("id_metric_cfg='" + id + "'").Update(&MeasurementFieldCfg{IDMetricCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Metric id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Info("Updated SnmpMetric Config to %s devices ", affecteddev)
	}

	affected, err = session.Where("id='" + id + "'").Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Info("Updated SnmpMetric Config Successfully with id %s and data:%+v, affected", id, dev)
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
			Type:   "measurements",
			obID:   val.IDMeasurementCfg,
			action: "delete field from measurement",
		})

	}
	return obj, nil
}

/***************************
	MEASUREMENTS
	-GetInfluxMeasurementCfgByID(struct)
	-GetInfluxMeasurementCfgMap (map - for interna config use
	-GetInfluxMeasurementCfgArray(Array - for web ui use )
	-AddInfluxMeasurementCfg
	-DelInfluxMeasurementCfg
	-UpdateInfluxMeasurementCfg
  -GetInfluxMeasurementCfgAffectOnDel
***********************************/
/*GetInfluxMeasurementCfgByID get metric data by id*/
func (dbc *DatabaseCfg) GetInfluxMeasurementCfgByID(id string) (InfluxMeasurementCfg, error) {
	cfgarray, err := cfg.Database.GetInfluxMeasurementCfgArray("id='" + id + "'")
	if err != nil {
		return InfluxMeasurementCfg{}, err
	}
	if len(cfgarray) > 1 {
		return InfluxMeasurementCfg{}, fmt.Errorf("Error %d results on get InfluxMeasurementCfg by id %s", len(cfgarray), id)
	}
	return *cfgarray[0], nil
}

/*GetInfluxMeasurementCfgMap  return data in map format*/
func (dbc *DatabaseCfg) GetInfluxMeasurementCfgMap(filter string) (map[string]*InfluxMeasurementCfg, error) {
	cfgarray, err := dbc.GetInfluxMeasurementCfgArray(filter)
	cfgmap := make(map[string]*InfluxMeasurementCfg)
	for _, val := range cfgarray {
		cfgmap[val.ID] = val
		log.Debugf("%+v", *val)
	}
	return cfgmap, err
}

/*GetInfluxMeasurementCfgArray generate an array of measurements with all its information */
func (dbc *DatabaseCfg) GetInfluxMeasurementCfgArray(filter string) ([]*InfluxMeasurementCfg, error) {
	var err error
	var devices []*InfluxMeasurementCfg
	//Get Only data for selected measurements
	if len(filter) > 0 {
		if err = dbc.x.Where(filter).Find(&devices); err != nil {
			log.Warnf("Fail to get InfluxMeasurementCfg  data filteter with %s : %v\n", filter, err)
			return nil, err
		}
	} else {
		if err = dbc.x.Find(&devices); err != nil {
			log.Warnf("Fail to get InfluxMeasurementCfg   data: %v\n", err)
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
				mVal.Fields = append(mVal.Fields, mm.IDMetricCfg)
			}
		}
	}
	return devices, nil
}

/*AddInfluxMeasurementCfg for adding new Metric*/
func (dbc *DatabaseCfg) AddInfluxMeasurementCfg(dev InfluxMeasurementCfg) (int64, error) {
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
	for _, metricID := range dev.Fields {

		mstruct := MeasurementFieldCfg{
			IDMeasurementCfg: dev.ID,
			IDMetricCfg:      metricID,
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
	return affected, nil
}

/*DelInfluxMeasurementCfg for deleting influx databases from ID*/
func (dbc *DatabaseCfg) DelInfluxMeasurementCfg(id string) (int64, error) {
	var affectedfl, affectedmg, affected int64
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

	affected, err = session.Where("id='" + id + "'").Delete(&InfluxMeasurementCfg{})
	if err != nil {
		session.Rollback()
		return 0, err
	}

	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Infof("Deleted Successfully Measurement with ID %s [ %d Measurements Groups Affected /%d Fields Affected ]", id, affectedmg, affectedfl)
	return affected, nil
}

/*UpdateInfluxMeasurementCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateInfluxMeasurementCfg(id string, dev InfluxMeasurementCfg) (int64, error) {
	var affecteddev, newmf, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		log.Info("Updated Measurement Config to %s devices ", affecteddev)
		affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Update(&MGroupsMeasurements{IDMeasurementCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Measurement id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Info("Updated Measurement config to %s devices ", affecteddev)
	}
	//delete all previous values
	affecteddev, err = session.Where("id_measurement_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		session.Rollback()
		return 0, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	//Creating nuew Measurement Fields
	for _, metricID := range dev.Fields {

		mstruct := MeasurementFieldCfg{
			IDMeasurementCfg: dev.ID,
			IDMetricCfg:      metricID,
		}
		newmf, err = session.Insert(&mstruct)
		if err != nil {
			session.Rollback()
			return 0, err
		}
	}
	//update data
	affected, err = session.Where("id='" + id + "'").Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Info("Updated Influx Measurement Config Successfully with id %s and  (%d previous / %d new Fields), affected", id, affecteddev, newmf)
	return affected, nil
}

/*GetInfluxMeasurementCfgAffectOnDel for deleting devices from ID*/
func (dbc *DatabaseCfg) GetInfluxMeasurementCfgAffectOnDel(id string) ([]*DbObjAction, error) {
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
			Type:   "fields",
			obID:   val.IDMeasurementCfg,
			action: "delete field from measurement",
		})
	}

	err = dbc.x.Where("id_measurement_cfg='" + id + "'").Find(&mg)
	if err != nil {
		return nil, fmt.Errorf("Error on Delete Measurement on MGroupsMeasurements with id: %s, error: %s", id, err)
	}

	for _, val := range mg {
		obj = append(obj, &DbObjAction{
			Type:   "measurements_group",
			obID:   val.IDMeasurementCfg,
			action: "delete measurements from  measurement group",
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
	cfgarray, err := cfg.Database.GetMeasFilterCfgArray("id='" + id + "'")
	if err != nil {
		return MeasFilterCfg{}, err
	}
	if len(cfgarray) > 1 {
		return MeasFilterCfg{}, fmt.Errorf("Error %d results on get MeasurementFilter by id %s", len(cfgarray), id)
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
	return affected, nil
}

/*UpdateMeasFilterCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateMeasFilterCfg(id string, dev MeasFilterCfg) (int64, error) {
	var affecteddev, newmf, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed only need change id's in snsmpdev
		affecteddev, err = session.Where("id_filter='" + id + "'").Update(&SnmpDevFilters{IDFilter: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Filter id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Info("Updated Measurement Filter Config to %s devices ", affecteddev)
	}

	//update data
	affected, err = session.Where("id='" + id + "'").Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Info("Updated Measurement Filter Config Successfully with id %s and  (%d previous / %d new Fields), affected", id, affecteddev, newmf)
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
			Type:   "snmpdevices",
			obID:   val.IDSnmpDev,
			action: "delete filter in device",
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
	cfgarray, err := cfg.Database.GetMGroupsCfgArray("id='" + id + "'")
	if err != nil {
		return MGroupsCfg{}, err
	}
	if len(cfgarray) > 1 {
		return MGroupsCfg{}, fmt.Errorf("Error %d results on get MGroupsCfg by id %s", len(cfgarray), id)
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
	log.Info("Added new Measurement Group Successfully with id %s  [%d Measurements]", dev.ID, newmf)
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
	return affected, nil
}

/*UpdateMGroupsCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateMGroupsCfg(id string, dev MGroupsCfg) (int64, error) {
	var affecteddev, newmg, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()

	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("id_mgroup_cfg='" + id + "'").Update(&SnmpDevMGroups{IDMGroupCfg: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error Update Metric id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Info("Updated Measurement Group Config to %s devices ", affecteddev)
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

	affected, err = session.Where("id='" + id + "'").Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Info("Updated Measurement Group Successfully with id %s [%d measurements], affected", dev.ID, newmg)
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
			Type:   "devices",
			obID:   val.IDSnmpDev,
			action: "delete field from measurement group",
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
***********************************/

/*GetSnmpDeviceCfgByID get device data by id*/
func (dbc *DatabaseCfg) GetSnmpDeviceCfgByID(id string) (SnmpDeviceCfg, error) {
	devcfgarray, err := cfg.Database.GetSnmpDeviceCfgArray("id='" + id + "'")
	if err != nil {
		return SnmpDeviceCfg{}, err
	}
	if len(devcfgarray) > 1 {
		return SnmpDeviceCfg{}, fmt.Errorf("Error %d results on get SnmpDeviceCfg by id %s", len(devcfgarray), id)
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
	log.Info("Added new Device Successfully with id %s [%d Measurment Groups | %d filters]", dev.ID, newmg, newft)
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
	affected, err = session.Where("id='" + id + "'").Update(dev)

	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}
	log.Info("Updated device constrains (old %d / new %d ) Measurement Groups", deletemg, newmg)
	log.Info("Updated device constrains (old %d / new %d ) MFilters", deleteft, newft)
	log.Info("Updated new Device Successfully with id %s and data:%+v", id, dev)
	return affected, nil
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
	cfgarray, err := cfg.Database.GetInfluxCfgArray("id='" + id + "'")
	if err != nil {
		return InfluxCfg{}, err
	}
	if len(cfgarray) > 1 {
		return InfluxCfg{}, fmt.Errorf("Error %d results on get SnmpDeviceCfg by id %s", len(cfgarray), id)
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
	log.Info("Added new influx backend Successfully with id %s ", dev.ID)
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
	return affected, nil
}

/*UpdateInfluxCfg for adding new influxdb*/
func (dbc *DatabaseCfg) UpdateInfluxCfg(id string, dev InfluxCfg) (int64, error) {
	var affecteddev, affected int64
	var err error
	session := dbc.x.NewSession()
	defer session.Close()
	if id != dev.ID { //ID has been changed
		affecteddev, err = session.Where("outdb='" + id + "'").Update(&SnmpDeviceCfg{OutDB: dev.ID})
		if err != nil {
			session.Rollback()
			return 0, fmt.Errorf("Error on Delete InfluxConfig on uopdate id(old)  %s with (new): %s, error: %s", id, dev.ID, err)
		}
		log.Info("Updated Influx Config to %s devices ", affecteddev)
	}

	affected, err = session.Where("id='" + id + "'").Update(dev)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	err = session.Commit()
	if err != nil {
		return 0, err
	}

	log.Info("Updated Influx Config Successfully with id %s and data:%+v, affected", id, dev)
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
			Type:   "device",
			obID:   val.ID,
			action: "reset OutDB from de device to",
		})

	}
	return obj, nil
}

//LoadConfig get data from database
func (dbc *DatabaseCfg) LoadConfig() {
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
	cfg.Measurements, err = dbc.GetInfluxMeasurementCfgMap("")
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

}
