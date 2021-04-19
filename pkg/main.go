package main

import (
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/agent/bus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/device"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/impexp"
	"github.com/toni-moreno/snmpcollector/pkg/data/measurement"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/webui"
)

var (
	log        = logrus.New()
	quit       = make(chan struct{})
	startTime  = time.Now()
	getversion bool
	httpListen = ":8080"
	appdir     = os.Getenv("PWD")
	homeDir    string
	pidFile    string
	logDir     = filepath.Join(appdir, "log")
	confDir    = filepath.Join(appdir, "conf")
	dataDir    = confDir
	configFile = filepath.Join(confDir, "config.toml")
)

func writePIDFile() {
	if pidFile == "" {
		return
	}

	// Ensure the required directory structure exists.
	err := os.MkdirAll(filepath.Dir(pidFile), 0700)
	if err != nil {
		log.Fatal(3, "Failed to verify pid directory", err)
	}

	// Retrieve the PID and write it.
	pid := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(pidFile, []byte(pid), 0644); err != nil {
		log.Fatal(3, "Failed to write pidfile", err)
	}
}

func flags() *flag.FlagSet {
	var f flag.FlagSet
	f.BoolVar(&getversion, "version", getversion, "display the version")
	f.StringVar(&configFile, "config", configFile, "config file")
	f.StringVar(&httpListen, "http", httpListen, "http port")
	f.StringVar(&logDir, "logs", logDir, "log directory")
	f.StringVar(&homeDir, "home", homeDir, "home directory")
	f.StringVar(&dataDir, "data", dataDir, "Data directory")
	f.StringVar(&pidFile, "pidfile", pidFile, "path to pid file")
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

func init() {
	//Log format
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.Formatter = customFormatter
	customFormatter.FullTimestamp = true

	// parse first time to see if config file is being specified
	f := flags()
	f.Parse(os.Args[1:])

	if getversion {
		t, _ := strconv.ParseInt(agent.BuildStamp, 10, 64)
		fmt.Printf("snmpcollector v%s (git: %s ) built at [%s]\n", agent.Version, agent.Commit, time.Unix(t, 0).Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}

	// now load up config settings
	if _, err := os.Stat(configFile); err == nil {
		viper.SetConfigFile(configFile)
		confDir = filepath.Dir(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/snmpcollector/")
		viper.AddConfigPath("/opt/snmpcollector/conf/")
		viper.AddConfigPath("./conf/")
		viper.AddConfigPath(".")
	}
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}
	err = viper.Unmarshal(&agent.MainConfig)
	if err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}
	log.Debugf("CONFIG FROM FILE : %+v", &agent.MainConfig)

	err = envconfig.Process("SNMPCOL_", &agent.MainConfig)
	if err != nil {
		log.Warnf("Some error happened when trying to read config from env: %s", err)
	}

	log.Debugf("CONFIG AFTER MERGE : %+v", &agent.MainConfig)

	cfg := &agent.MainConfig

	log.Infof("Main agent Logging will be written to %s ", cfg.General.LogMode)
	if cfg.General.LogMode == "console" {
		//default if not set
		log.Out = os.Stdout

	} else {
		if len(cfg.General.LogDir) > 0 {
			logDir = cfg.General.LogDir
		}
		os.Mkdir(logDir, 0755)
		//Log output
		f, _ := os.OpenFile(logDir+"/snmpcollector.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		log.Out = f
	}

	if len(cfg.General.LogLevel) > 0 {
		l, _ := logrus.ParseLevel(cfg.General.LogLevel)
		log.Level = l
	}
	if len(cfg.General.DataDir) > 0 {
		dataDir = cfg.General.DataDir
	}
	if len(cfg.General.HomeDir) > 0 {
		homeDir = cfg.General.HomeDir
	}
	//check if exist public dir in home
	if _, err := os.Stat(filepath.Join(homeDir, "public")); err != nil {
		log.Warnf("There is no public (www) directory on [%s] directory", homeDir)
		if len(homeDir) == 0 {
			homeDir = appdir
		}
	}
	//needed to create SQLDB when SQLite and debug log
	config.SetLogger(log)
	config.SetDirs(dataDir, logDir, confDir)
	//needed to log all snmp console related commands
	snmp.SetLogger(log)
	snmp.SetLogDir(logDir)

	output.SetLogger(log)
	selfmon.SetLogger(log)
	//devices needs access to all db loaded data
	device.SetDBConfig(&agent.DBConfig)
	device.SetLogDir(logDir)

	measurement.SetConfDir(confDir)
	webui.SetLogger(log)
	webui.SetLogDir(logDir)
	webui.SetConfDir(confDir)
	webui.SetLogMode(cfg.General.LogMode)

	agent.SetLogger(log)

	impexp.SetLogger(log)
	bus.SetLogger(log)
	//
	log.Infof("Set Default directories : \n   - Exec: %s\n   - Config: %s\n   -Logs: %s\n -Home: %s\n", appdir, confDir, logDir, homeDir)
}

func main() {

	defer func() {
		//errorLog.Close()
	}()
	writePIDFile()
	//Init BD config
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGABRT, syscall.SIGINT)
	go func() {
		for {
			select {
			case sig := <-c:
				switch sig {
				case syscall.SIGABRT, syscall.SIGINT:
					log.Infof("Received %v signal: Forcing shutdown", sig)
					os.Exit(1)
				case syscall.SIGTERM:
					log.Infof("Received %v signal: Trigger a ordered shutdown (could take some time)", sig)
					agent.End()
					log.Infof("Exiting for requested user: %v", sig)
					os.Exit(1)
				case syscall.SIGHUP:
					log.Infof("Received HUP signal: Trigger a ordered reload (could take some time -- usually Max gathering time form all nodes --)")
					agent.ReloadConf()
				}

			}
		}

	}()

	agent.MainConfig.Database.InitDB(&agent.MainConfig.General)
	measurement.SetDB(&agent.MainConfig.Database)
	impexp.SetDB(&agent.MainConfig.Database)

	agent.Start()

	webui.WebServer(filepath.Join(homeDir, "public"), httpListen, &agent.MainConfig.HTTP, agent.MainConfig.General.InstanceID)

}
