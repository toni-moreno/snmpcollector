package main

import (
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
	MetricGroups []string `xorm:"-"`
	MeasFilters  []string `xorm:"-"`
}

//SnmpMetricCfg Metric config
type SnmpMetricCfg struct {
	ID          string  `xorm:"'id' unique"` //name of the key in the config array
	FieldName   string  `xorm:"field_name"`
	Description string  `xorm:"description"`
	BaseOID     string  `xorm:"baseoid"`
	DataSrcType string  `xorm:"datasrctype"`
	GetRate     bool    `xorm:"getrate"` //ony Valid with COUNTERS/ABSOLUTE
	Scale       float64 `xorm:"scale"`
	Shift       float64 `xorm:"shift"`
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
		datasource = appdir + "/conf/" + dbc.Name + ".db"
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

//LoadConfig get data from database
func (dbc *DatabaseCfg) LoadConfig() {
	var err error

	//Load Influxdb databases

	var influxdb []*InfluxCfg
	if err = dbc.x.Find(&influxdb); err != nil {
		log.Warnf("Fail to get Influxdb config data: %v\n", err)
	}
	cfg.Influxdb = make(map[string]*InfluxCfg)
	for _, val := range influxdb {
		cfg.Influxdb[val.ID] = val
	}
	//Load metrics
	var metrics []*SnmpMetricCfg
	if err = dbc.x.Find(&metrics); err != nil {
		log.Warnf("Fail to get Metrics data: %v\n", err)
	}

	cfg.Metrics = make(map[string]*SnmpMetricCfg)
	for _, val := range metrics {
		cfg.Metrics[val.ID] = val
	}

	//Load Measurements

	var measurements []*InfluxMeasurementCfg
	if err = dbc.x.Find(&measurements); err != nil {
		log.Warnf("Fail to get Measurements data: %v\n", err)
	}

	cfg.Measurements = make(map[string]*InfluxMeasurementCfg)
	for _, val := range measurements {
		cfg.Measurements[val.ID] = val
	}

	var MeasureMetric []*MeasurementFieldCfg
	if err = dbc.x.Find(&MeasureMetric); err != nil {
		log.Warnf("Fail to get Measurements Metric relationship data: %v\n", err)
	}

	//Load Measurements and metrics relationship
	//We asign field metric ID to each measurement
	for _, mVal := range cfg.Measurements {
		for _, mm := range MeasureMetric {
			if mm.IDMeasurementCfg == mVal.ID {
				mVal.Fields = append(mVal.Fields, mm.IDMetricCfg)
			}
		}
	}

	//Load Measurement Filters

	var mfilters []*MeasFilterCfg
	if err = dbc.x.Find(&mfilters); err != nil {
		log.Warnf("Fail to get Measurement Filters  data: %v\n", err)
	}

	cfg.MFilters = make(map[string]*MeasFilterCfg)
	for _, val := range mfilters {
		cfg.MFilters[val.ID] = val
	}

	//Load measourement Groups

	var mgroups []*MGroupsCfg
	if err = dbc.x.Find(&mgroups); err != nil {
		log.Warnf("Fail to get Measurement Groups  data: %v\n", err)
	}

	cfg.GetGroups = make(map[string]*MGroupsCfg)
	for _, val := range mgroups {
		cfg.GetGroups[val.ID] = val
	}

	//Load measurement for each groups
	var mgroupsmeas []*MGroupsMeasurements
	if err = dbc.x.Find(&mgroupsmeas); err != nil {
		log.Warnf("Fail to get MGroup Measurements relationship  data: %v\n", err)
	}

	for _, mVal := range cfg.GetGroups {
		for _, mgm := range mgroupsmeas {
			if mgm.IDMGroupCfg == mVal.ID {
				mVal.Measurements = append(mVal.Measurements, mgm.IDMeasurementCfg)
			}
		}
	}

	//Load Device Configurations

	var devices []*SnmpDeviceCfg
	if err = dbc.x.Find(&devices); err != nil {
		log.Warnf("Fail to get SnmpDevices Groups  data: %v\n", err)
	}

	cfg.SnmpDevice = make(map[string]*SnmpDeviceCfg)
	for _, val := range devices {
		cfg.SnmpDevice[val.ID] = val
		log.Debugf("%+v", *val)
	}

	//Asign Groups to devices.
	var snmpdevmgroups []*SnmpDevMGroups
	if err = dbc.x.Find(&snmpdevmgroups); err != nil {
		log.Warnf("Fail to get SnmpDevices and Measurement groups relationship data: %v\n", err)
	}

	//Load Measurements and metrics relationship
	//We asign field metric ID to each measurement
	for _, mVal := range cfg.SnmpDevice {
		for _, mg := range snmpdevmgroups {
			if mg.IDSnmpDev == mVal.ID {
				mVal.MetricGroups = append(mVal.MetricGroups, mg.IDMGroupCfg)
			}
		}
	}

	//Asign Filters to devices.
	var snmpdevfilters []*SnmpDevFilters
	if err = dbc.x.Find(&snmpdevfilters); err != nil {
		log.Warnf("Fail to get SnmpDevices and Filter relationship data: %v\n", err)
	}

	//Load Measurements and metrics relationship
	//We asign field metric ID to each measurement
	for _, mVal := range cfg.SnmpDevice {
		for _, mf := range snmpdevfilters {
			if mf.IDSnmpDev == mVal.ID {
				mVal.MeasFilters = append(mVal.MeasFilters, mf.IDFilter)
			}
		}
	}

}
