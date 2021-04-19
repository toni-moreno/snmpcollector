package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/bus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/device"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"github.com/toni-moreno/snmpcollector/pkg/config"
)

var (
	// Version is the app X.Y.Z version
	Version string
	// Commit is the git commit sha1
	Commit string
	// Branch is the git branch
	Branch string
	// BuildStamp is the build timestamp
	BuildStamp string
)

// RInfo contains the agent's release and version information.
type RInfo struct {
	InstanceID string
	Version    string
	Commit     string
	Branch     string
	BuildStamp string
}

// GetRInfo returns the agent release information.
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
	// Bus is the messaging system used to send messages to the devices
	Bus = bus.NewBus()

	// MainConfig contains the global configuration
	MainConfig config.Config

	// DBConfig contains the database config
	DBConfig config.DBConfig

	log *logrus.Logger
	// reloadMutex guards the reloadProcess flag
	reloadMutex   sync.Mutex
	reloadProcess bool
	// mutex guards the runtime devices map access
	mutex sync.RWMutex
	// devices is the runtime snmp devices map
	devices map[string]*device.SnmpDevice
	// influxdb is the runtime devices output db map
	influxdb map[string]*output.InfluxDB

	selfmonProc *selfmon.SelfMon
	// gatherWg synchronizes device specific goroutines
	gatherWg sync.WaitGroup
	senderWg sync.WaitGroup
)

// SetLogger sets the current log output.
func SetLogger(l *logrus.Logger) {
	log = l
}

// Reload Mutex Related Methods.

// CheckReloadProcess checks if the agent is currently reloading config.
func CheckReloadProcess() bool {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	return reloadProcess
}

// CheckAndSetReloadProcess sets the reloadProcess flag.
// Returns its previous value.
func CheckAndSetReloadProcess() bool {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	retval := reloadProcess
	reloadProcess = true
	return retval
}

// CheckAndUnSetReloadProcess unsets the reloadProcess flag.
// Returns its previous value.
func CheckAndUnSetReloadProcess() bool {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	retval := reloadProcess
	reloadProcess = false
	return retval
}

// PrepareInfluxDBs initializes all configured output DBs in the SQL database.
// If there is no "default" key, creates a dummy output db which does nothing.
func PrepareInfluxDBs() map[string]*output.InfluxDB {
	idb := make(map[string]*output.InfluxDB)

	var defFound bool
	for k, c := range DBConfig.Influxdb {
		if k == "default" {
			defFound = true
		}
		idb[k] = output.NewNotInitInfluxDB(c)
	}
	if defFound == false {
		log.Warn("No Output default found influxdb devices found !!")
		idb["default"] = output.DummyDB
	}
	return idb
}

// GetDevice returns the snmp device with the given id.
// Returns an error if there is an ongoing reload.
func GetDevice(id string) (*device.SnmpDevice, error) {
	var dev *device.SnmpDevice
	var ok bool
	if CheckReloadProcess() == true {
		log.Warning("There is a reload process running while trying to get device info")
		return nil, fmt.Errorf("There is a reload process running.... please wait until finished ")
	}
	mutex.RLock()
	defer mutex.RUnlock()
	if dev, ok = devices[id]; !ok {
		return nil, fmt.Errorf("There is not any device with id %s running", id)
	}
	return dev, nil
}

// GetDeviceJSONInfo returns the device data in JSON format.
// Returns an error if there is an ongoing reload.
func GetDeviceJSONInfo(id string) ([]byte, error) {
	var dev *device.SnmpDevice
	var ok bool
	if CheckReloadProcess() == true {
		log.Warning("There is a reload process running while trying to get device info")
		return nil, fmt.Errorf("There is a reload process running.... please wait until finished ")
	}
	mutex.RLock()
	defer mutex.RUnlock()
	if dev, ok = devices[id]; !ok {
		return nil, fmt.Errorf("there is not any device with id %s running", id)
	}
	return dev.ToJSON()
}

// GetDevStats returns a map with the basic info of each device.
func GetDevStats() map[string]*device.DevStat {
	devstats := make(map[string]*device.DevStat)
	mutex.RLock()
	for k, v := range devices {
		devstats[k] = v.GetBasicStats()
	}
	mutex.RUnlock()
	return devstats
}

// StopInfluxOut stops sending data to output influxDB servers.
func StopInfluxOut(idb map[string]*output.InfluxDB) {
	for k, v := range idb {
		log.Infof("Stopping Influxdb out %s", k)
		v.StopSender()
	}
}

// ReleaseInfluxOut closes the influxDB connections and releases the associated resources.
func ReleaseInfluxOut(idb map[string]*output.InfluxDB) {
	for k, v := range idb {
		log.Infof("Release Influxdb resources %s", k)
		v.End()
	}
}

// DeviceProcessStop stops all device polling goroutines
func DeviceProcessStop() {
	Bus.Broadcast(&bus.Message{Type: "exit"})
}

// DeviceProcessStart starts all device polling goroutines
func DeviceProcessStart() {
	mutex.Lock()
	devices = make(map[string]*device.SnmpDevice)
	mutex.Unlock()

	for k, c := range DBConfig.SnmpDevice {
		AddDeviceInRuntime(k, c)
	}
}

// ReleaseDevices releases all devices resources.
func ReleaseDevices() {
	mutex.RLock()
	for _, c := range devices {
		c.End()
	}
	mutex.RUnlock()
}

func init() {
	go Bus.Start()
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
			selfmonProc.SetOutDB(idb)
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

// IsDeviceInRuntime checks if device `id` exists in the runtime array.
func IsDeviceInRuntime(id string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := devices[id]; ok {
		return true
	}
	return false

}

// DeleteDeviceInRuntime removes the device `id` from the runtime array.
func DeleteDeviceInRuntime(id string) error {
	if dev, ok := devices[id]; ok {
		dev.StopGather()
		log.Debugf("Bus retuned from the exit message to the ID device %s", id)
		dev.LeaveBus(Bus)
		dev.End()
		mutex.Lock()
		delete(devices, id)
		mutex.Unlock()
		return nil
	}
	log.Errorf("There is no  %s device in the runtime device list", id)
	return nil
}

// AddDeviceInRuntime initializes each SNMP device and puts the pointer to the global device map.
func AddDeviceInRuntime(k string, cfg *config.SnmpDeviceCfg) {
	// Initialize each SNMP device and put pointer to the global map devices
	dev := device.New(cfg)
	dev.AttachToBus(Bus)
	dev.InitCatalogVar(DBConfig.VarCatalog)
	dev.SetSelfMonitoring(selfmonProc)

	// send a db map to initialize each one its own db if needed
	outdb, _ := dev.GetOutSenderFromMap(influxdb)
	outdb.Init()
	outdb.StartSender(&senderWg)

	mutex.Lock()
	devices[k] = dev
	dev.StartGather(&gatherWg)
	mutex.Unlock()
}

// LoadConf loads the DB conf and initializes the device metric config.
func LoadConf() {
	MainConfig.Database.LoadDbConfig(&DBConfig, MainConfig.General.Location)
	influxdb = PrepareInfluxDBs()

	// begin self monitoring process if needed, before all goroutines
	initSelfMonitoring(influxdb)
	config.InitMetricsCfg(&DBConfig)
}

// Start loads the agent configuration and starts it.
func Start() {
	LoadConf()
	DeviceProcessStart()
}

// End stops all devices polling.
func End() (time.Duration, error) {

	start := time.Now()
	log.Infof("END: begin device Gather processes stop... at %s", start.String())
	// stop all device processes
	DeviceProcessStop()
	log.Info("END: begin selfmon Gather processes stop...")
	// stop the selfmon process
	selfmonProc.StopGather()
	log.Info("END: waiting for all Gather goroutines stop...")
	// wait until Done
	gatherWg.Wait()
	log.Info("END: releasing Device Resources")
	ReleaseDevices()
	log.Info("END: releasing Selfmonitoring Resources")
	selfmonProc.End()
	log.Info("END: begin sender processes stop...")
	//log.Info("DEBUG Gather WAIT %+v", GatherWg)
	//log.Info("DEBUG SENDER WAIT %+v", senderWg)
	// stop all Output Emitter
	StopInfluxOut(influxdb)
	log.Info("END: waiting for all Sender goroutines stop..")
	senderWg.Wait()
	log.Info("END: releasing Sender Resources")
	ReleaseInfluxOut(influxdb)
	log.Infof("END: Finished from %s to %s [Duration : %s]", start.String(), time.Now().String(), time.Since(start).String())
	return time.Since(start), nil
}

// ReloadConf stops the polling, reloads all configuration and restart the polling.
func ReloadConf() (time.Duration, error) {
	start := time.Now()
	if CheckAndSetReloadProcess() == true {
		log.Warnf("RELOADCONF: There is another reload process running while trying to reload at %s  ", start.String())
		return time.Since(start), fmt.Errorf("There is another reload process running.... please wait until finished ")
	}

	log.Infof("RELOADCONF INIT: begin device Gather processes stop... at %s", start.String())
	End()

	log.Info("RELOADCONF: loading configuration Again...")
	LoadConf()
	log.Info("RELOADCONF: Starting all device processes again...")
	// Initialize Devices in Runtime map
	DeviceProcessStart()

	log.Infof("RELOADCONF END: Finished from %s to %s [Duration : %s]", start.String(), time.Now().String(), time.Since(start).String())
	CheckAndUnSetReloadProcess()

	return time.Since(start), nil
}
