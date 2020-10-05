package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	// _ needed to mysql
	_ "github.com/go-sql-driver/mysql"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"

	// _ needed to sqlite3
	_ "github.com/mattn/go-sqlite3"
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

//DbObjAction measurement groups to assign to devices
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
		protocol := "tcp"
		if strings.HasPrefix(dbc.Host, "/") {
			protocol = "unix"
		}
		datasource = fmt.Sprintf("%s:%s@%s(%s)/%s?charset=utf8", dbc.User, dbc.Password, protocol, dbc.Host, dbc.Name)
		//datasource = dbc.User + ":" + dbc.Pass + "@" + dbc.Host + "/" + dbc.Name + "?charset=utf8"
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
	if err = dbc.x.Sync(new(VarCatalogCfg)); err != nil {
		log.Fatalf("Fail to sync database VarCatalogCfg: %v\n", err)
	}
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

	SetupDefaultDbData(dbc)
}

// SetupDefaultDbData initialize our database with mandatory default values
func SetupDefaultDbData(dbc *DatabaseCfg) {

	// create default monitoring InfluxDB

	var id string
	hastInfluxDb, existsError := dbc.x.Table("influx_cfg").Where("ID = ?", "influx").Cols("ID").Get(&id)

	if existsError != nil {
		log.Fatalf("Error on determine wether monitioring InfluxDB exists, error: %s", existsError)
	}


	var influxHost = os.Getenv("INFLUXDB_HOST")
	if influxHost == "" {
		influxHost = "localhost" // default
	}

	var influxDB = os.Getenv("INFLUXDB_DB")
	if influxDB == "" {
		influxDB = "monitoring" // default
	}

	var influxUser = os.Getenv("INFLUXDB_USER")
	if influxUser == "" {
		influxUser = "admin" // default
	}

	var influxPassword = os.Getenv("INFLUXDB_USER_PASSWORD")
	if influxPassword == "" {
		influxPassword = "admin" // default
	}

	if !hastInfluxDb {
		// create default influx db

		monitoringInfluxDB := InfluxCfg{
			ID:                 "influx",
			Host:               influxHost,
			Port:               8086,
			DB:                 influxDB,
			User:               influxUser,
			Password:           influxPassword,
			Retention:          "autogen",
			Precision:          "s",
			Timeout:            30,
			UserAgent:          "snmpcollector",
			EnableSSL:          false,
			InsecureSkipVerify: true,
			BufferSize:         1000,
		}

		affected, err := dbc.AddInfluxCfg(monitoringInfluxDB)
		if err != nil {
			log.Fatalf("Error on insert monitoring InfluxDB, affected : %+v , error: %s", affected, err)
		}
	} else {
		// update default influx db

		var defaultInfluxDb, err = dbc.GetInfluxCfgByID("influx")
		if err != nil {
			log.Fatalf("Error on update default InfluxDB with env variables, error: %s", err)
		} else {

			defaultInfluxDb.Host = influxHost
			defaultInfluxDb.DB = influxDB
			defaultInfluxDb.User = influxUser
			defaultInfluxDb.Password = influxPassword

			var _, err = dbc.UpdateInfluxCfg(defaultInfluxDb.ID, defaultInfluxDb)
			if err != nil {
				log.Fatalf("Error on update default InfluxDB with env variables, error: %s", err)
			}
		}

	}
}

// CatalogVar2Map return interface map from variable table
func CatalogVar2Map(cv map[string]*VarCatalogCfg) map[string]interface{} {
	m := make(map[string]interface{})
	var err error
	for k, v := range cv {
		log.Infof("KEY %s Type %s %value ", v.ID, v.Type, v.Value)
		switch v.Type {
		case "string":
			m[k] = v.Value
		case "integer":
			m[k], err = strconv.ParseInt(v.Value, 10, 64)
			if err != nil {
				log.Errorf("Error in Integer convesrion %s value %s: %s", v.Type, v.Value, err)
			}
		case "float":
			m[k], err = strconv.ParseFloat(v.Value, 64)
			if err != nil {
				log.Errorf("Error in Float convesrion %s value %s: %s", v.Type, v.Value, err)
			}
		default:
			log.Warnf("unknown type %s", v.Type)
		}
	}
	return m
}

//LoadDbConfig get data from database
func (dbc *DatabaseCfg) LoadDbConfig(cfg *DBConfig) {
	var err error
	//Load Global Variables
	VarCatalog := make(map[string]*VarCatalogCfg)
	VarCatalog, err = dbc.GetVarCatalogCfgMap("")
	if err != nil {
		log.Warningf("Some errors on get Global variables :%v", err)
	}
	cfg.VarCatalog = make(map[string]interface{}, len(VarCatalog))
	cfg.VarCatalog = CatalogVar2Map(VarCatalog)

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
