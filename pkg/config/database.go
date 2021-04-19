package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	// _ needed to mysql
	_ "github.com/go-sql-driver/mysql"

	// needed to posgress

	_ "github.com/lib/pq"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"

	// _ needed to sqlite3
	_ "github.com/mattn/go-sqlite3"

	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
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
func (dbc *DatabaseCfg) InitDB(cfg *GeneralConfig) error {
	// Create ORM engine and database
	var err error
	var dbtype string
	var datasource string

	log.Debugf("Database config: %+v", dbc)

	switch dbc.Type {
	case "sqlite3":
		dbtype = "sqlite3"
		datasource = dataDir + "/" + dbc.Name + ".db"

	case "postgres", "postgresql":
		dbtype = "postgres"
		addr, err := utils.SplitHostPortDefault(dbc.Host, "127.0.0.1", "5432")
		if err != nil {
			log.Errorf("Invalid host specifier '%s': Err: %s", dbc.Host, err)
			return err
		}

		if dbc.Password == "" {
			dbc.Password = "''"
		}
		if dbc.User == "" {
			dbc.User = "''"
		}
		if dbc.SslMode == "" {
			dbc.SslMode = "disable"
		}
		datasource = fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", dbc.User, dbc.Password, addr.Host, addr.Port, dbc.Name, dbc.SslMode, dbc.ClientCertPath, dbc.ClientKeyPath, dbc.CaCertPath)

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
		return nil
	}

	dbc.x, err = xorm.NewEngine(dbtype, datasource)
	if err != nil {
		log.Fatalf("Fail to create engine: %v\n", err)
	}

	if len(dbc.LogMode) > 0 && dbc.LogMode != "none" {
		log.Infof("Enabled SQL logging into: %s", dbc.LogMode)
		dbc.x.ShowSQL(true)
		if dbc.LogMode == "console" {

			dbc.x.SetLogger(xorm.NewSimpleLogger(os.Stdout))
		} else {
			filename := "sql.log"
			if len(dbc.SQLLogFile) > 0 {
				filename = dbc.SQLLogFile
			}
			f, error := os.Create(logDir + "/" + filename)
			if err != nil {
				log.Errorln("Fail to create log file  ", error)
			}
			dbc.x.SetLogger(xorm.NewSimpleLogger(f))
		}

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
	if err = dbc.x.Sync(new(PollerLocationCfg)); err != nil {
		log.Fatalf("Fail to sync database PollerLocationCfg: %v\n", err)
	}
	//Lookup if PollerLocation if got Instance_ID else create one
	var locationsFound []*PollerLocationCfg
	if err = dbc.x.Where("Instance_ID = '" + cfg.InstanceID + "'").Find(&locationsFound); err != nil {
		log.Fatalf("There were an error when looking for %s in PollerLocationCfg. Error: %+v", cfg.InstanceID, err)
	} else {
		if len(locationsFound) == 0 {
			//Create Location

			//Get Hostname
			hostname, err := os.Hostname()
			if err != nil {
				hostname = cfg.InstanceID
			}
			//Get external IP
			ip := getExternalIp()
			//Get Location or assign InstanceID
			loc := cfg.Location
			if len(loc) == 0 {
				loc = cfg.InstanceID
			}

			var location = PollerLocationCfg{
				ID:          cfg.InstanceID,
				Location:    loc,
				Instance_ID: cfg.InstanceID,
				Active:      true,
				Hostname:    hostname,
				IP:          ip.String(),
				Description: cfg.Description,
			}

			var affected int64
			session := dbc.x.NewSession()
			defer session.Close()
			affected, err = session.Insert(&location)
			if err != nil {
				log.Fatalf("There were an error when creating default PollerLocationCfg Error: %+v", err)
				session.Rollback()
			} else {
				log.Infof("PollerLocationCfg created succefully. %+v", affected)
			}
		} else {
			log.Debug("Instance_ID founded on PollerLocationCfg. Checking if any param was changed")
			for _, instance := range locationsFound {
				log.Debugf("--> %+v\n", instance)
			}
			//Check if instance params was changed
			instance := locationsFound[0]
			hasChanges := false
			//IP
			if instance.IP != getExternalIp().String() {
				instance.IP = getExternalIp().String()
				hasChanges = true
			}
			//Description
			if instance.Description != cfg.Description {
				instance.Description = cfg.Description
				hasChanges = true
			}
			//Hostname
			hostname, err := os.Hostname()
			if err != nil {
				hostname = cfg.InstanceID
			}
			if instance.Hostname != hostname {
				instance.Description = hostname
				hasChanges = true
			}
			//Location
			loc := cfg.Location
			if len(loc) == 0 {
				loc = cfg.InstanceID
			}
			if instance.Location != loc {
				instance.Location = loc
				hasChanges = true
			}

			if hasChanges {
				log.Debug("Updating with newer values...")
				var affected int64
				session := dbc.x.NewSession()
				defer session.Close()
				//affected, err := session.ID(instance.ID).AllCols().Update()
				affected, err = session.Where("Instance_ID ='" + instance.Instance_ID + "'").AllCols().Update(instance)
				if err != nil {
					log.Fatalf("There were an error when updating PollerLocationCfg Error: %+v", err)
					session.Rollback()
				} else {
					log.Infof("PollerLocationCfg updated succefully. %+v", affected)
				}
			}
		}
	}
	return nil
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
func (dbc *DatabaseCfg) LoadDbConfig(cfg *DBConfig, Location string) {
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
	//Parameters inside GetSnmpDeviceCfgMap are filter to get devices. All devices who match with location will be returned
	cfg.SnmpDevice, err = dbc.GetSnmpDeviceCfgMap("location = '" + Location + "'")
	if err != nil {
		log.Warningf("Some errors on get SnmpDeviceConf :%v", err)
	}
	dbc.resetChanges()
}

//Utils
func getExternalIp() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
