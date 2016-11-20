package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// GeneralConfig has miscelaneous configuration options
type GeneralConfig struct {
	InstanceID string `toml:"instanceID"`
	LogDir     string `toml:"logdir"`
	LogLevel   string `toml:"loglevel"`
}

var (
	version    string
	commit     string
	branch     string
	buildstamp string
)

var (
	log        = logrus.New()
	quit       = make(chan struct{})
	startTime  = time.Now()
	showConfig bool
	getversion bool
	httpPort   = 8080
	appdir     = os.Getenv("PWD")
	logDir     = filepath.Join(appdir, "log")
	confDir    = filepath.Join(appdir, "conf")
	configFile = filepath.Join(confDir, "config.toml")

	cfg = struct {
		General      GeneralConfig
		Database     DatabaseCfg
		Selfmon      SelfMonConfig
		Metrics      map[string]*SnmpMetricCfg
		Measurements map[string]*InfluxMeasurementCfg
		MFilters     map[string]*MeasFilterCfg
		GetGroups    map[string]*MGroupsCfg
		SnmpDevice   map[string]*SnmpDeviceCfg
		Influxdb     map[string]*InfluxCfg
		HTTP         HTTPConfig
	}{}
	//runtime devices
	devices map[string]*SnmpDevice
	//runtime output db's
	influxdb map[string]*InfluxDB
)

func flags() *flag.FlagSet {
	var f flag.FlagSet
	f.BoolVar(&getversion, "version", getversion, "display de version")
	f.BoolVar(&showConfig, "showconf", showConfig, "show all devices config and exit")
	f.StringVar(&configFile, "config", configFile, "config file")
	f.IntVar(&httpPort, "http", httpPort, "http port")
	f.StringVar(&logDir, "logs", logDir, "log directory")
	f.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		f.VisitAll(func(flag *flag.Flag) {
			format := "%10s: %s\n"
			fmt.Fprintf(os.Stderr, format, "-"+flag.Name, flag.Usage)
		})
		fmt.Fprintf(os.Stderr, "\nAll settings can be set in config file: %s\n", configFile)
		os.Exit(1)

	}
	return &f
}

/*
initMetricsCfg this function does 2 things
1.- Initialice id from key of maps for all SnmpMetricCfg and InfluxMeasurementCfg objects
2.- Initialice references between InfluxMeasurementCfg and SnmpMetricGfg objects
*/

func initMetricsCfg() error {
	//TODO:
	// - check duplicates OID's => warning messages
	//Initialize references to SnmpMetricGfg into InfluxMeasurementCfg
	log.Debug("--------------------Initializing Config metrics-------------------")
	log.Debug("Initializing SNMPMetricconfig...")
	for mKey, mVal := range cfg.Metrics {
		err := mVal.Init(mKey)
		if err != nil {
			log.Warnln("Error in Metric config:", err)
			//if some error int the format the metric is deleted from the config
			delete(cfg.Metrics, mKey)
		}
	}
	log.Debug("Initializing MEASSUREMENTSconfig...")
	for mKey, mVal := range cfg.Measurements {
		err := mVal.Init(mKey, &cfg.Metrics)
		if err != nil {
			log.Warnln("Error in Metric config:", err)
			//if some error int the format the metric is deleted from the config
			delete(cfg.Metrics, mKey)
		}

		log.Debugf("FIELDMETRICS: %+v", mVal.fieldMetric)
	}
	log.Debug("-----------------------END Config metrics----------------------")
	return nil
}

//PrepareInfluxDBs review all configured db's in the SQL database
// and check if exist at least a "default", if not creates a dummy db which does nothing
func PrepareInfluxDBs() map[string]*InfluxDB {
	idb := make(map[string]*InfluxDB)

	defFound := false
	for k, c := range cfg.Influxdb {
		//Inticialize each SNMP device
		if k == "default" {
			defFound = true
		}
		dev := InfluxDB{
			cfg:     c,
			dummy:   false,
			started: false,
			Sent:    0,
			Errors:  0,
		}
		idb[k] = &dev
	}
	if defFound == false {
		//no devices configured  as default device we need to set some device as itcan send data transparant to snmpdevices goroutines
		log.Warn("No Output default found influxdb devices found !!")
		idb["default"] = influxdbDummy
	}
	return idb
}

func init() {
	//Log format
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.Formatter = customFormatter
	customFormatter.FullTimestamp = true
	//----

	// parse first time to see if config file is being specified
	f := flags()
	f.Parse(os.Args[1:])

	if getversion {
		t, _ := strconv.ParseInt(buildstamp, 10, 64)
		fmt.Printf("snmpcollector v%s (git: %s ) built at [%s]\n", version, commit, time.Unix(t, 0).Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}

	// now load up config settings
	if _, err := os.Stat(configFile); err == nil {
		viper.SetConfigFile(configFile)
		confDir = filepath.Dir(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/opt/snmpcollector/conf/")
		viper.AddConfigPath("./conf/")
		viper.AddConfigPath(".")
	}
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}

	if len(cfg.General.LogDir) > 0 {
		logDir = cfg.General.LogDir
		os.Mkdir(logDir, 0755)
	}
	if len(cfg.General.LogLevel) > 0 {
		l, _ := logrus.ParseLevel(cfg.General.LogLevel)
		log.Level = l
	}
	//
	log.Infof("Set Default directories : \n   - Exec: %s\n   - Config: %s\n   -Logs: %s\n", appdir, confDir, logDir)

	//Init BD config

	InitDB(&cfg.Database)
	cfg.Database.LoadConfig()

	//Prepare the InfluxDataBases Configuration

	influxdb := PrepareInfluxDBs()

	// beginning self monitoring process if needed.( before each other gorotines)

	if cfg.Selfmon.Enabled {
		if val, ok := influxdb["default"]; ok {
			//only executed if a "*" influxdb exist
			cfg.Selfmon.Init()
			cfg.Selfmon.Influx = val
			cfg.Selfmon.Influx.Init()
			log.Printf("SELFMON enabled %+v", cfg.Selfmon)
		} else {
			cfg.Selfmon.Enabled = false
			log.Errorf("SELFMON disabled becaouse of no default db found !!! SELFMON[ %+v ]  INFLUXLIST[ %+v]\n", cfg.Selfmon, influxdb)
		}
	} else {
		log.Printf("SELFMON disabled %+v\n", cfg.Selfmon)
	}
	//Initialize Device Metrics CFG

	initMetricsCfg()

	//Initialize Device Runtime map

	devices = make(map[string]*SnmpDevice)

	for k, c := range cfg.SnmpDevice {
		//Inticialize each SNMP device and put pointer to the global map devices
		dev := NewSnmpDevice(c)
		//send db's map to initialize each one its own db if needed and not yet initialized
		dev.AttachOutDBMap(influxdb)
		devices[k] = dev
	}

	// only run when one needs to see the interface names of the device
	if showConfig {
		for _, c := range devices {
			fmt.Println("\nSNMP host:", c.cfg.ID)
			fmt.Println("=========================================")
			c.printConfig()
		}
		os.Exit(0)
	}

}

func main() {
	var wg sync.WaitGroup
	defer func() {
		//errorLog.Close()
	}()
	if cfg.Selfmon.Enabled {
		cfg.Selfmon.ReportStats(&wg)
	}

	for _, c := range devices {
		wg.Add(1)
		go c.Gather(&wg)
	}

	var port int
	if cfg.HTTP.Port > 0 {
		port = cfg.HTTP.Port
	} else {
		port = httpPort
	}

	if port > 0 {
		webServer(port)
	} else {
		<-quit
	}
}
