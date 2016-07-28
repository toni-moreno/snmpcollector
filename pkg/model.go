package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

// InfluxCfg is the main configuration for any InfluxDB TSDB
type InfluxCfg struct {
	ID        string `xorm:"id"`
	Host      string `xorm:"host"`
	Port      int    `xorm:"port"`
	DB        string `xorm:"db"`
	User      string `xorm:"user"`
	Password  string `xorm:"password"`
	Retention string `xorm:"retention"`
}

// SnmpDeviceCfg contains all snmp related device definitions
type SnmpDeviceCfg struct {
	ID string `xorm:"id"`
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
	MetricGroups []string   `xorm:"-"`
	MeasFilters  [][]string `xorm:"-"`
}

//SnmpMetricCfg Metric config
type SnmpMetricCfg struct {
	ID          string  `xorm:"id"` //name of the key in the config array
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
	ID   string `xorm:"id"`
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
	ID               string `xorm:"id"`
	idMeasurementCfg string `xorm:"id_measurement_cfg"`
	fType            string `xorm:"filter_type"`  //file/OIDCondition
	FileName         string `xorm:"file_name"`    //only vaid if file
	enableAlias      bool   `xorm:"enable_alias"` //only valid if file
	OIDCond          string `xorm:"oid_name"`
	condType         string `xorm:"condition"`
	condValue        string `xorm:"value"`
}

//MGroupsCfg measurement groups to asign to devices
type MGroupsCfg struct {
	ID           string   `xorm:"id"`
	Measurements []string `xorm:"-"`
}

//MGroupsMeasurements measurements contained on each Measurement Group
type MGroupsMeasurements struct {
	idMGroupCfg      string `xorm:"id_mgroup_cfg"`
	idMeasurementCfg string `xorm:"id_measurement_cfg"`
}

//SnmpDevFilters filters to use with indexed measurement
type SnmpDevFilters struct {
	idSnmpDev string `xorm:"id_snmpdev"`
	idFilter  string `xorm:"id_filter"`
}

//SnmpDevMGroups Mgroups defined on each SnmpDevice
type SnmpDevMGroups struct {
	idSnmpDev   string `xorm:"id_snmpdev"`
	idMGroupCfg string `xorm:"id_mgroup_cfg"`
}

// ORM engine
//var x *xorm.Engine

//DatabaseCfg de configuration for the database
type DatabaseCfg struct {
	Type string `toml:"type"`
	Host string `toml:"host"`
	Name string `toml:"name"`
	User string `toml:"user"`
	Pass string `toml:"password"`
	x    *xorm.Engine
}

//InitDB initialize de BD configuration
func InitDB(dbc DatabaseCfg) {
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

	dbc.x.ShowSQL(true)

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

/*
Metrics      map[string]*SnmpMetricCfg
Measurements map[string]*InfluxMeasurementCfg
GetGroups    map[string]*MGroupsCfg
SnmpDevice   map[string]*SnmpDeviceCfg
Influxdb     map[string]*InfluxCfg
*/

//LoadConfig get data from database
func (dbc *DatabaseCfg) LoadConfig() {
	var err error

	var MeasureMetric []*MeasurementFieldCfg
	if err = dbc.x.Find(&MeasureMetric); err != nil {
		log.Warnf("Fail to get Measurements Metric relationship data: %v\n", err)
	}
	//Load metrics

	//cfg.Metrics = make(map[string]*SnmpMetricCfg)
	var kk []SnmpMetricCfg
	if err = dbc.x.Find(&kk); err != nil {
		log.Warnf("Fail to get Metrics data: %v\n", err)
	}
	/*
		for _, val := range kk {
			cfg.Metrics[val.ID] = &val
		}

		cfg.Measurements = make(map[string]*InfluxMeasurementCfg)
		if err = dbc.x.Find(cfg.Measurements); err != nil {
			log.Warnf("Fail to get Measurements data: %v\n", err)
		}

		var MeasureMetric []*MeasurementFieldCfg
		if err = dbc.x.Find(&MeasureMetric); err != nil {
			log.Warnf("Fail to get Measurements Metric relationship data: %v\n", err)
		}

		// We asign field names to
		for _, mVal := range cfg.Measurements {
			for _, mm := range MeasureMetric {
				if mm.idMeasurementCfg == mVal.ID {
					mVal.Fields = append(mVal.Fields, mm.idMetricCfg)
				}
			}
		}
	*/
	//Assinging data to the Measurement.fieldMetric array

}
