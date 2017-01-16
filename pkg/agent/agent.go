package agent

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/device"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"github.com/toni-moreno/snmpcollector/pkg/config"

	"sync"
	"time"
)

var (
	Version    string
	Commit     string
	Branch     string
	BuildStamp string
)

// RInfo  Release basic version info for the agent
type RInfo struct {
	InstanceID string
	Version    string
	Commit     string
	Branch     string
	BuildStamp string
}

func GetRInfo() *RInfo {
	info := &RInfo{
		InstanceID: MainConfig.General.InstanceID,
		Version:    Version,
		Commit:     Commit,
		Branch:     Branch,
		BuildStamp: BuildStamp,
	}
	return info
}

var (

	// MainConfig has all configuration
	MainConfig config.Config

	// DBConfig db config
	DBConfig config.SQLConfig

	log *logrus.Logger
	//mutex for devices map
	mutex sync.Mutex
	//runtime devices
	devices map[string]*device.SnmpDevice
	//runtime output db's
	influxdb map[string]*output.InfluxDB

	selfmonProc *selfmon.SelfMon
	// for synchronize  deivce specific goroutines
	gatherWg sync.WaitGroup
	senderWg sync.WaitGroup
)

// SetLogger set log output
func SetLogger(l *logrus.Logger) {
	log = l
}

//PrepareInfluxDBs review all configured db's in the SQL database
// and check if exist at least a "default", if not creates a dummy db which does nothing
func PrepareInfluxDBs() map[string]*output.InfluxDB {
	idb := make(map[string]*output.InfluxDB)

	var defFound bool
	for k, c := range DBConfig.Influxdb {
		//Inticialize each SNMP device
		if k == "default" {
			defFound = true
		}
		idb[k] = output.NewNotInitInfluxDB(c)
	}
	if defFound == false {
		//no devices configured  as default device we need to set some device as itcan send data transparant to snmpdevices goroutines
		log.Warn("No Output default found influxdb devices found !!")
		idb["default"] = output.DummyDB
	}
	return idb
}

//GetDevice is a safe method to get a Device Object
func GetDevice(id string) (*device.SnmpDevice, error) {
	var dev *device.SnmpDevice
	var ok bool
	mutex.Lock()
	if dev, ok = devices[id]; !ok {
		return nil, fmt.Errorf("there is not any device with id %s running", id)
	}
	mutex.Unlock()
	return dev, nil
}

// GetDevStats xx
func GetDevStats() map[string]*device.DevStat {
	devstats := make(map[string]*device.DevStat)
	mutex.Lock()
	for k, v := range devices {
		devstats[k] = v.GetBasicStats()
	}
	mutex.Unlock()
	return devstats
}

// StopInfluxOut xx
func StopInfluxOut(idb map[string]*output.InfluxDB) {
	for k, v := range idb {
		log.Infof("Stopping Influxdb out %s", k)
		v.StopSender()
	}
}

// ReleaseInfluxOut xx
func ReleaseInfluxOut(idb map[string]*output.InfluxDB) {
	for k, v := range idb {
		log.Infof("Release Influxdb resources %s", k)
		v.End()
	}
}

// DeviceProcessStop stop all device goroutines
func DeviceProcessStop() {
	mutex.Lock()
	for _, c := range devices {
		c.StopGather()
	}
	mutex.Unlock()
}

// DeviceProcessStart start all devices goroutines
func DeviceProcessStart() {
	mutex.Lock()
	for _, c := range devices {
		c.StartGather(&gatherWg)
	}
	mutex.Unlock()
}

// ReleaseDevices Executes End for each device
func ReleaseDevices() {
	mutex.Lock()
	for _, c := range devices {
		c.End()
	}
	mutex.Unlock()
}

func initSelfMonitoring(idb map[string]*output.InfluxDB) {
	log.Debugf("INFLUXDB2: %+v", idb)
	selfmonProc = selfmon.NewNotInit(&MainConfig.Selfmon)

	if MainConfig.Selfmon.Enabled {
		if val, ok := idb["default"]; ok {
			//only executed if a "default" influxdb exist
			val.Init()
			val.StartSender(&senderWg)

			selfmonProc.Init()
			selfmonProc.SetOutput(val)

			log.Printf("SELFMON enabled %+v", MainConfig.Selfmon)
			//Begin the statistic reporting
			selfmonProc.StartGather(&gatherWg)
		} else {
			MainConfig.Selfmon.Enabled = false
			log.Errorf("SELFMON disabled becaouse of no default db found !!! SELFMON[ %+v ]  INFLUXLIST[ %+v]\n", MainConfig.Selfmon, idb)
		}
	} else {
		log.Printf("SELFMON disabled %+v\n", MainConfig.Selfmon)
	}
}

// LoadConf call to initialize alln configurations
func LoadConf() {
	//Load all database info to Cfg struct
	MainConfig.Database.LoadDbConfig(&DBConfig)
	//Prepare the InfluxDataBases Configuration
	influxdb = PrepareInfluxDBs()

	// beginning self monitoring process if needed.( before each other gorotines could begin)

	initSelfMonitoring(influxdb)

	//Initialize Device Metrics CFG

	config.InitMetricsCfg(&DBConfig)

	//Initialize Device Runtime map

	devices = make(map[string]*device.SnmpDevice)

	for k, c := range DBConfig.SnmpDevice {
		//Inticialize each SNMP device and put pointer to the global map devices
		dev := device.New(c)
		dev.SetSelfMonitoring(selfmonProc)
		//send db's map to initialize each one its own db if needed and not yet initialized

		outdb, _ := dev.GetOutSenderFromMap(influxdb)
		outdb.Init()
		outdb.StartSender(&senderWg)

		mutex.Lock()
		devices[k] = dev
		mutex.Unlock()
	}

	//beginning  the gather process
}

// ReloadConf call to reinitialize alln configurations
func ReloadConf() time.Duration {
	start := time.Now()
	log.Info("RELOADCONF: begin device Gather processes stop...")
	//stop all device prcesses
	DeviceProcessStop()
	log.Info("RELOADCONF: begin selfmon Gather processes stop...")
	//stop the selfmon process
	selfmonProc.StopGather()
	log.Info("RELOADCONF: waiting for all Gather gorotines stop...")
	//wait until Done
	gatherWg.Wait()
	log.Info("RELOADCONF: releasing Device Resources")
	ReleaseDevices()
	log.Info("RELOADCONF: releasing Seflmonitoring Resources")
	selfmonProc.End()
	log.Info("RELOADCONF: begin sender processes stop...")
	//stop all Output Emmiter
	//log.Info("DEBUG Gather WAIT %+v", GatherWg)
	//log.Info("DEBUG SENDER WAIT %+v", senderWg)
	StopInfluxOut(influxdb)
	log.Info("RELOADCONF: waiting for all Sender gorotines stop..")
	senderWg.Wait()
	log.Info("RELOADCONF: releasing Sender Resources")
	ReleaseInfluxOut(influxdb)

	log.Info("RELOADCONF: ĺoading configuration Again...")
	LoadConf()
	log.Info("RELOADCONF: Starting all device processes again...")
	DeviceProcessStart()
	return time.Since(start)
}
