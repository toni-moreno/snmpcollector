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
	LogDir   string `toml:"logdir"`
	LogLevel string `toml:"loglevel"`
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
	verbose    bool
	startTime  = time.Now()
	showConfig bool
	getversion bool
	repeat     = 0
	freq       = 30
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
	//runtme array
	devices  map[string]*SnmpDevice
	influxdb map[string]*InfluxDB
)

func flags() *flag.FlagSet {
	var f flag.FlagSet
	f.BoolVar(&getversion, "version", getversion, "display de version")
	f.BoolVar(&showConfig, "showconf", showConfig, "show all devices config and exit")
	f.StringVar(&configFile, "config", configFile, "config file")
	f.BoolVar(&verbose, "verbose", verbose, "verbose mode")
	f.IntVar(&freq, "freq", freq, "delay (in seconds)")
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

	log.Printf("set Default directories : \n   - Exec: %s\n   - Config: %s\n   -Logs: %s\n", appdir, confDir, logDir)

	// now load up config settings
	if _, err := os.Stat(configFile); err == nil {
		viper.SetConfigFile(configFile)
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
	}
	if len(cfg.General.LogLevel) > 0 {
		l, _ := logrus.ParseLevel(cfg.General.LogLevel)
		log.Level = l

	}
	//Init BD config
	log.Debugf("%+v", cfg)
	InitDB(&cfg.Database)
	cfg.Database.LoadConfig()
	log.Debugf("%+v", cfg)
	//Init Metrics CFG

	initMetricsCfg()
	log.Debugf("%+v", cfg)

	//Init InfluxDataBases

	influxdb = make(map[string]*InfluxDB)

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
		//dev.cfg.ID = k
		//dev.Init(k) D'ont initialize if there is not any device really sending data to the output backend
		influxdb[k] = &dev
	}
	if defFound == false {
		//no devices configured  as default device we need to set some device as itcan send data transparant to snmpdevices goroutines
		log.Warn("No Output default found influxdb devices found !!")
		influxdb["default"] = influxdbDummy
	}

	//Init Devices

	devices = make(map[string]*SnmpDevice)

	var ok bool
	for k, c := range cfg.SnmpDevice {
		//Inticialize each SNMP device
		dev := SnmpDevice{}
		dev.cfg = c
		dev.Init(k)
		if dev.cfg.Freq == 0 {
			dev.cfg.Freq = freq
		}
		devices[k] = &dev
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

	// re-read cmd line args to override as indicated
	f = flags()
	f.Parse(os.Args[1:])
	os.Mkdir(logDir, 0755)

	// now make sure each snmp device has a db

	//for name, c := range cfg.SnmpDevice {
	for name, c := range devices {
		// default is to use name of snmp config, but it can be overridden
		if len(c.cfg.OutDB) > 0 {
			name = c.cfg.OutDB
		}
		if c.Influx, ok = influxdb[name]; !ok {
			if c.Influx, ok = influxdb["default"]; !ok {
				log.Errorf("No influx config for snmp device: %s", name)
			}
		}
		c.Influx.Init()

	}

	//make sure the selfmon has a deb

	if cfg.Selfmon.Enabled {
		if val, ok := influxdb["*"]; ok {
			//only executed if a "*" influxdb exist
			cfg.Selfmon.Init()
			cfg.Selfmon.Influx = val
			cfg.Selfmon.Influx.Init()
			fmt.Printf("SELFMON enabled %+vn\n", cfg.Selfmon)
		} else {
			cfg.Selfmon.Enabled = false
		}
	} else {
		fmt.Printf("SELFMON disabled %+vn\n", cfg.Selfmon)
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
