package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/toni-moreno/snmpcollector/pkg/agent/bus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/device"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output/backend"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/stats"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
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

// swagger:model RInfo
// RInfo contains the agent release and version information.
type RInfo struct {
	// InstanceID the unique name identificator for this agent
	InstanceID string
	// Version is the app X.Y.Z version
	Version string
	// Commit is the git commit sha1
	Commit string
	// Branch is the git branch
	Branch string
	// BuildStamp is the build timestamp
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

	log utils.Logger
	// reloadMutex guards the reloadProcess flag
	reloadMutex   sync.Mutex
	reloadProcess bool
	// mutex guards the runtime devices map access
	mutex sync.RWMutex
	// devices is the runtime snmp devices map
	devices map[string]*device.SnmpDevice
	// outputs is the runtime devices output db map
	outputs map[string]*output.SinkDB

	selfmonProc *selfmon.SelfMon
	// gatherWg synchronizes device specific goroutines
	gatherWg sync.WaitGroup
	senderWg sync.WaitGroup
)

// SetLogger sets the current log output.
func SetLogger(l utils.Logger) {
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

// PrepareOutputs initializes all configured output DBs in the SQL database.
// If there is no "default" key, creates a dummy output db which does nothing.
func PrepareOutputs() map[string]*output.SinkDB {
	outtemps := make(map[string]*output.SinkDB)
	var err error

	// Outputs finally follows his own lifecycle and has relation with backend as 1:1
	// we need to retrieve all available outputs and retrieve the backend configuration based on map
	// the influx/kafka params should be retrieved with special DBConfig as an interface
	for id, out := range DBConfig.Outputs {
		// load config
		sinkDB := &output.SinkDB{
			Cfg: out,
		}
		// supported outputs - kafka, influxdb
		switch out.BackendType {
		case "kafka":
			if bcfg, ok := DBConfig.Kafka[out.Backend]; ok {
				sinkDB.Backend, err = backend.NewNotInitKafka(bcfg)
				if err != nil {
					log.Errorf("Unable to initialize Kafka backend - %s", out.Backend)
					continue
				}
			} else {
				log.Errorf("Kafka server not found, output %s not registered", out.ID)
				continue
			}
		case "influxdb":
			if bcfg, ok := DBConfig.Influxdb[out.Backend]; ok {
				sinkDB.Backend = backend.NewNotInitInfluxDB(bcfg)
			} else {
				log.Errorf("Influx server not found, output %s not registered", out.ID)
				continue
			}
		}
		// register on temp outputs
		outtemps[id] = sinkDB
	}

	if _, ok := outtemps["default"]; !ok {
		log.Warn("No Output default devices found")
		cfg := config.OutputCfg{
			ID:             "dummy",
			BackendType:    "dummy",
			Backend:        "dummy",
			Active:         false,
			EnqueueOnError: false,
			BufferSize:     0,
		}
		outtemps["default"] = &output.SinkDB{
			Cfg:     &cfg,
			Backend: backend.NewNotInitDummyDB(),
		}
	}
	//return idb
	return outtemps
}

// GetDevice returns the snmp device with the given id.
// Returns an error if there is an ongoing reload.
func GetDevice(id string) (*device.SnmpDevice, error) {
	var dev *device.SnmpDevice
	var ok bool
	if CheckReloadProcess() {
		log.Warning("There is a reload process running while trying to get device info")
		return nil, fmt.Errorf("there is a reload process running.... please wait until finished ")
	}
	mutex.RLock()
	defer mutex.RUnlock()
	if dev, ok = devices[id]; !ok {
		return nil, fmt.Errorf("there is not any device with id %s running", id)
	}
	return dev, nil
}

// GetDeviceJSONInfo returns the device data in JSON format.
// Returns an error if there is an ongoing reload.
func GetDeviceJSONInfo(id string) ([]byte, error) {
	var dev *device.SnmpDevice
	var ok bool
	if CheckReloadProcess() {
		log.Warnf("There is a reload process running while trying to get device info")
		return nil, fmt.Errorf("there is a reload process running.... please wait until finished ")
	}
	mutex.RLock()
	defer mutex.RUnlock()
	if dev, ok = devices[id]; !ok {
		return nil, fmt.Errorf("there is not any device with id %s running", id)
	}
	return dev.ToJSON()
}

// GetDevStats returns a map with the basic info of each device.
func GetDevStats() map[string]*stats.GatherStats {
	devstats := make(map[string]*stats.GatherStats)
	mutex.RLock()
	for k, v := range devices {
		devstats[k] = v.GetBasicStats()
	}
	mutex.RUnlock()
	return devstats
}

// StopInfluxOut stops sending data to output outputs servers.
func StopOutputs(idb map[string]*output.SinkDB) {
	for k, v := range idb {
		log.Infof("Stopping output %s", k)
		v.StopSender()
	}
}

// ReleaseOutputs closes the output connections and releases the associated resources.
func ReleaseOutputs(idb map[string]*output.SinkDB) {
	for k, v := range idb {
		log.Infof("Release output resources %s", k)
		v.End()
	}
}

// DeviceProcessStop stops all device polling goroutines.
func DeviceProcessStop() {
	Bus.Broadcast(&bus.Message{Type: bus.Exit})
}

// DeviceProcessStart starts all device polling goroutines.
func DeviceProcessStart() {
	mutex.Lock()
	devices = make(map[string]*device.SnmpDevice)
	mutex.Unlock()

	for k, c := range DBConfig.SnmpDevice {
		AddDeviceInRuntime(k, c)
	}
}

func init() {
	go Bus.Start()
}

// initSelfMonitoring initialize the selfmon gourutine.
func initSelfMonitoring(outdbs map[string]*output.SinkDB) {
	// create new selfmon
	selfmonProc = selfmon.NewNotInit(&MainConfig.Selfmon)

	// if its enabled, try to find the default backend
	if MainConfig.Selfmon.Enabled {
		if outdb, ok := outdbs["default"]; ok {
			// initialize the related SinkDB and related sender
			// declare our sinkdb as a container for generic output
			err := outdb.Init()
			if err != nil {
				log.Errorf("Unable to init output - %s", err)
				return
			}
			outdb.StartSender(&senderWg)

			// initialize the selfmon proc and attach all the outputs
			selfmonProc.Init()
			selfmonProc.SetOutDB(outdbs)
			selfmonProc.SetOutput(outdb)

			log.Printf("SELFMON enabled %+v", MainConfig.Selfmon)
			// Begin the statistic reporting
			selfmonProc.StartGather(&gatherWg)
		} else {
			MainConfig.Selfmon.Enabled = false
			log.Errorf("SELFMON disabled, no default db found. Selfmon: [%+v] OutDB list: [%+v]\n", MainConfig.Selfmon, outdbs)
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
	// Avoid modifications to devices while deleting device.
	mutex.Lock()
	defer mutex.Unlock()
	if dev, ok := devices[id]; ok {
		// Stop all device processes and its measurements. Once finished they will be removed
		// from the bus and node closed (snmp connections for measurements will be closed).
		dev.StopGather()
		log.Debugf("Bus retuned from the exit message to the ID device %s", id)
		delete(devices, id)
		return nil
	}
	log.Errorf("There is no  %s device in the runtime device list", id)
	return nil
}

// AddDeviceInRuntime initializes each SNMP device and puts the pointer to the global device map.
func AddDeviceInRuntime(k string, cfg *config.SnmpDeviceCfg) {
	// Initialize each SNMP device and put pointer to the global map devices.
	dev := device.New(cfg)
	dev.AttachToBus(Bus)
	dev.InitCatalogVar(DBConfig.VarCatalog)
	dev.SetSelfMonitoring(selfmonProc)

	// send a db map to initialize each one its own db if needed
	outdb, err := dev.GetOutSenderFromMap(outputs)
	if err != nil {
		log.Errorf("Unable to retrieve output from map - %s", err)
		return
	}
	// declare our sinkdb as a container for generic output
	err = outdb.Init()
	if err != nil {
		log.Errorf("Unable to init output - %s", err)
		return
	}
	outdb.StartSender(&senderWg)
	mutex.Lock()
	devices[k] = dev
	// Start gather goroutine for device and add it to the wait group for gather goroutines.
	gatherWg.Add(1)
	go func() {
		defer gatherWg.Done()
		dev.StartGather()
		log.Infof("Device %s finished", cfg.ID)
		// If device goroutine has finished, leave the bus so it won't get blocked trying
		// to send messages to a not running device.
		dev.LeaveBus(Bus)
	}()
	mutex.Unlock()
}

// LoadConf loads the DB conf and initializes the device metric config.
func LoadConf() {
	MainConfig.Database.LoadDbConfig(&DBConfig)
	outputs = PrepareOutputs()

	// begin self monitoring process if needed, before all goroutines
	initSelfMonitoring(outputs)
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
	// Stop all device processes and its measurements. Once finished they will be removed
	// from the bus and node closed (snmp connections for measurements will be closed)
	DeviceProcessStop()
	log.Info("END: begin selfmon Gather processes stop...")
	// stop the selfmon process
	selfmonProc.StopGather()
	log.Info("END: waiting for all Gather goroutines stop...")
	// wait until Done
	gatherWg.Wait()
	log.Info("END: releasing Selfmonitoring Resources")
	selfmonProc.End()
	log.Info("END: begin sender processes stop...")
	// log.Info("DEBUG Gather WAIT %+v", GatherWg)
	// log.Info("DEBUG SENDER WAIT %+v", senderWg)
	// stop all Output Emitter
	StopOutputs(outputs)
	log.Info("END: waiting for all Sender goroutines stop..")
	senderWg.Wait()
	log.Info("END: releasing Sender Resources")
	ReleaseOutputs(outputs)
	log.Infof("END: Finished from %s to %s [Duration : %s]", start.String(), time.Now().String(), time.Since(start).String())
	return time.Since(start), nil
}

// ReloadConf stops the polling, reloads all configuration and restart the polling.
func ReloadConf() (time.Duration, error) {
	start := time.Now()
	if CheckAndSetReloadProcess() {
		log.Warnf("RELOADCONF: There is another reload process running while trying to reload at %s  ", start.String())
		return time.Since(start), fmt.Errorf("there is another reload process running.... please wait until finished ")
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
