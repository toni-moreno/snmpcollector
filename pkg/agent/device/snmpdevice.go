package device

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/bus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/measurement"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/stats"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

const (
	// DEFAULT_MAX_OIDS default value if value in config is 0 or less
	DEFAULT_MAX_OIDS = 60
)

var (
	cfg    *config.DBConfig
	logDir string
)

// SetDBConfig set agent config
func SetDBConfig(c *config.DBConfig) {
	cfg = c
}

// SetLogDir set log dir
func SetLogDir(l string) {
	logDir = l
}

// SnmpDevice contains all runtime device related device configu ns and state
type SnmpDevice struct {
	cfg *config.SnmpDeviceCfg
	log *logrus.Logger
	// basic sistem info
	SysInfo *snmp.SysInfo
	// runtime built TagMap
	TagMap map[string]string
	// Refresh data to show in the frontend
	Freq int
	// Measurements array
	Measurements []*measurement.Measurement
	// Variable map
	VarMap map[string]interface{}
	// Influx client shared between measurement goroutines to send data and stats to the backend
	Influx *output.InfluxDB `json:"-"`
	// LastError     time.Time
	// Runtime stats
	stats stats.GatherStats // Runtime Internal statistic
	// TODO asegurarnos que cuando se escriba aquí esté protegido por el rtData
	Stats *stats.GatherStats // Public info for thread safe accessing to the data ()

	// runtime controls
	rtData    sync.RWMutex
	statsData sync.RWMutex

	DeviceActive    bool
	DeviceConnected bool

	Node      *bus.Node `json:"-"`
	isStopped chan bool `json:"-"`

	CurLogLevel     string
	Gather          func()                                                            `json:"-"`
	InitSnmpConnect func(mkey string, debug bool, maxrep uint8) (*snmp.Client, error) `json:"-"`
	// Needed to inicialize measurement selfmon
	selfmon *selfmon.SelfMon
}

// New create and Initialice a device Object
func New(c *config.SnmpDeviceCfg) *SnmpDevice {
	dev := SnmpDevice{}
	dev.Init(c)
	return &dev
}

// GetLogFilePath return current LogFile
func (d *SnmpDevice) GetLogFilePath() string {
	return d.cfg.LogFile
}

// ToJSON return a JSON version of the device data
func (d *SnmpDevice) ToJSON() ([]byte, error) {
	// Get read lock for SnmpDevice struct (protect all values except Measurements)
	d.rtData.RLock()
	defer d.rtData.RUnlock()

	// To avoid data racing while reading SnmpDevice.Measurements, Measurement implements a custom
	// MarshalJSON function, grabbing there the lock

	result, err := json.MarshalIndent(&struct {
		SysInfo         *snmp.SysInfo
		TagMap          map[string]string
		Freq            int
		Measurements    []*measurement.Measurement
		VarMap          map[string]interface{}
		Stats           *stats.GatherStats // Public info for thread safe accessing to the data ()
		DeviceActive    bool
		DeviceConnected bool
		CurLogLevel     string
	}{
		SysInfo:         d.SysInfo,
		TagMap:          d.TagMap,
		Freq:            d.Freq,
		Measurements:    d.Measurements,
		VarMap:          d.VarMap,
		Stats:           d.Stats,
		DeviceActive:    d.DeviceActive,
		DeviceConnected: d.DeviceConnected,
		CurLogLevel:     d.CurLogLevel,
	}, "", "  ")
	if err != nil {
		d.Errorf("Error on Get JSON data from device")
		dummy := []byte{}
		return dummy, nil
	}
	return result, err
}

// GetBasicStats get basic info for this device
func (d *SnmpDevice) GetBasicStats() *stats.GatherStats {
	d.statsData.RLock()
	defer d.statsData.RUnlock()
	return d.Stats
}

// GetBasicStats get basic info for this device
// TODO cuidado con data races
func (d *SnmpDevice) getBasicStats() *stats.GatherStats {
	sum := 0
	d.DeviceConnected = false
	for _, m := range d.Measurements {
		st := m.GetBasicStats()
		if st == nil {
			d.log.Warnf("No Basic stats exist in Device [%s] Measurement: %s", d.cfg.ID, m.ID)
			panic(fmt.Errorf("No Basic stats exist in Device [%s] Measurement: %s", d.cfg.ID, m.ID))
			// continue
		}
		d.stats.Combine(st)
		d.DeviceConnected = d.DeviceConnected || st.Connected
		sum += len(m.OidSnmpMap)
	}
	d.stats.SetStatus(d.DeviceActive, d.DeviceConnected)
	stat := d.stats.ThSafeCopy()
	stat.TagMap = d.TagMap
	// reporting only measurements/metrics to the UI only if connected
	if d.DeviceConnected {
		stat.NumMeasurements = len(d.Measurements)
		stat.NumMetrics = sum
	}
	if d.SysInfo != nil {
		stat.SysDescription = d.SysInfo.SysDescr
	} else {
		stat.SysDescription = ""
	}
	return stat
}

// GetOutSenderFromMap to get info about the sender will use
func (d *SnmpDevice) GetOutSenderFromMap(influxdb map[string]*output.InfluxDB) (*output.InfluxDB, error) {
	if len(d.cfg.OutDB) == 0 {
		d.Warnf("No OutDB configured on the device")
	}
	var ok bool
	name := d.cfg.OutDB
	if d.Influx, ok = influxdb[name]; !ok {
		// we assume there is always a default db
		if d.Influx, ok = influxdb["default"]; !ok {
			// but
			return nil, fmt.Errorf("No influx config for snmp device: %s", d.cfg.ID)
		}
	}

	return d.Influx, nil
}

// ForceGather send message to force a data gather execution
func (d *SnmpDevice) ForceGather() {
	d.Node.SendMsg(&bus.Message{Type: bus.ForceGather})
}

// ForceFltUpdate send info to update the filter counter to the next execution
func (d *SnmpDevice) ForceFltUpdate() {
	d.Node.SendMsg(&bus.Message{Type: bus.FilterUpdate})
}

// SnmpReset send message to init an  SNMP connection reset could be soft/hard
func (d *SnmpDevice) SnmpReset(mode string) {
	switch mode {
	case "hard":
		d.Node.SendMsg(&bus.Message{Type: bus.SNMPResetHard})
	case "soft":
		d.Node.SendMsg(&bus.Message{Type: bus.SNMPReset})
	default:
		d.log.Infof("Unknown mode %s on SNMPRESET ", mode)
	}
}

// StopGather send signal to stop the Gathering process
func (d *SnmpDevice) StopGather() {
	d.Node.SendMsg(&bus.Message{Type: bus.SyncExit})
	<-d.isStopped
	d.log.Info("Exiting from StopGather process...")
}

// RTActivate change activatio state in runtime
func (d *SnmpDevice) RTActivate(activate bool) {
	d.Node.SendMsg(&bus.Message{Type: bus.Enabled, Data: activate})
}

// RTActSnmpDebug change snmp debug runtime
func (d *SnmpDevice) RTActSnmpDebug(activate bool) {
	d.Node.SendMsg(&bus.Message{Type: bus.SNMPDebug, Data: activate})
}

// RTActSnmpMaxRep change snmp MaxRepetitions
func (d *SnmpDevice) RTActSnmpMaxRep(maxrep uint8) {
	d.Node.SendMsg(&bus.Message{Type: bus.SetSNMPMaxRep, Data: maxrep})
}

// RTSetLogLevel set the log level for this device
func (d *SnmpDevice) RTSetLogLevel(level string) {
	d.Node.SendMsg(&bus.Message{Type: bus.LogLevel, Data: level})
}

/*
InitDevMeasurements  does the following
- look for all defined measurements from the template grups
- allocate and fill array for all measurements defined for this device
- for each indexed measurement  load device labels from IndexedOID and fiter them if defined measurement filters.
- Initialice each SnmpMetric from each measuremet.
*/
//InitDevMeasurements generte all internal structs from SNMP device
func (d *SnmpDevice) InitDevMeasurements() {
	// Alloc array
	d.Measurements = make([]*measurement.Measurement, 0, 0)
	d.Debugf("---Init device measurements from groups %s------------------", d.cfg.Host)
	// for this device get MeasurementGroups and search all measurements

	// For each measurement group in the device, get the group config
	for _, devMeas := range d.cfg.MeasurementGroups {
		// Selecting all Metric Groups that matches with device.MeasurementGroups
		selGroups := make(map[string]*config.MGroupsCfg, 0)
		// var RegExp = regexp.MustCompile(devMeas)
		for key, val := range cfg.GetGroups {
			if key == devMeas {
				selGroups[key] = val
			}
		}
		d.Debugf("this device has this SELECTED GROUPS: %+v", selGroups)
		// Only For selected Groups we will get all selected measurements and we will remove repeated values
		var selMeas []string
		for key, val := range selGroups {
			d.Debugf("Selecting from group %s", key)
			for _, item := range val.Measurements {
				d.Debugf("Selecting measurements  %s from group %s", item, key)
				selMeas = append(selMeas, item)
			}
		}
		// remove duplicated measurements if needed
		selMeasUniq := utils.RemoveDuplicatesUnordered(selMeas)
		// Now we know what measurements names  will send influx from this device

		d.Debugf("DEVICE MEASUREMENT: %s HOST: %s ", devMeas, d.cfg.Host)
		for _, val := range selMeasUniq {
			// check if measurement exist
			if mVal, ok := cfg.Measurements[val]; !ok {
				d.Warnf("no measurement configured with name %s in host : %s", val, d.cfg.Host)
			} else {
				d.Debugf("MEASUREMENT CFG KEY: %s VALUE %s", val, mVal.Name)

				// Pass a logger with predefined values to distinguish this host-measurement
				measLog := d.log.WithFields(logrus.Fields{
					"device":      d.cfg.Host,
					"measurement": val,
				})

				mstat := stats.GatherStats{}
				mstat.Init("measurement", mVal.Name, d.TagMap, measLog)
				mstat.SetSelfMonitoring(d.selfmon)
				// creating a new measurement runtime object and asigning to array
				// MeasFilters and MFitlers used in the InitFilters function used in the initialization of the measurement goroutine
				imeas := measurement.New(mVal, d.cfg.MeasFilters, cfg.MFilters, d.cfg.Active, measLog)
				imeas.SetStats(mstat)
				d.Measurements = append(d.Measurements, imeas)
			}
		}
	}

	// Initialize all snmpMetrics  objects and OID array
	// get data first time
	// useful to inicialize counter all value and test device snmp availability
}

/*
Init  does the following

- Initialize not set variables to some defaults
- Initialize logfile for this device
- Initialize comunication channels and initial device state
No need to get the rtData lock because the device it's not yet referenced.
*/
func (d *SnmpDevice) Init(c *config.SnmpDeviceCfg) error {
	if c == nil {
		return fmt.Errorf("Error on initialice device, configuration struct is nil")
	}
	d.cfg = c
	d.isStopped = make(chan bool)
	// log.Infof("Initializing device %s\n", d.cfg.ID)

	// Init Logger
	if d.cfg.Freq == 0 {
		d.cfg.Freq = 60
	}
	if len(d.cfg.LogFile) == 0 {
		d.cfg.LogFile = logDir + "/" + d.cfg.ID + ".log"
	}
	if len(d.cfg.LogLevel) == 0 {
		d.cfg.LogLevel = "info"
	}

	f, _ := os.OpenFile(d.cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644)
	d.log = logrus.New()
	d.log.Out = f
	l, _ := logrus.ParseLevel(d.cfg.LogLevel)
	d.log.Level = l
	d.CurLogLevel = d.log.Level.String()
	// Formatter for time
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	d.log.Formatter = customFormatter
	customFormatter.FullTimestamp = true

	d.DeviceActive = d.cfg.Active
	d.stats.SetStatus(d.DeviceActive, false)

	// Init Device Tags

	if len(d.cfg.DeviceTagName) == 0 {
		d.cfg.DeviceTagName = "device"
	}

	d.Freq = d.cfg.Freq

	var val string

	switch d.cfg.DeviceTagValue {
	case "id":
		val = d.cfg.ID
	case "host":
		val = d.cfg.Host
	default:
		val = d.cfg.ID
		d.Warnf("Unkwnown DeviceTagValue %s set ID (%s) as value", d.cfg.DeviceTagValue, val)
	}

	d.TagMap = make(map[string]string)
	d.TagMap[d.cfg.DeviceTagName] = val

	if len(d.cfg.ExtraTags) > 0 {
		for _, tag := range d.cfg.ExtraTags {
			s := strings.Split(tag, "=")
			if len(s) == 2 {
				key, value := s[0], s[1]
				d.TagMap[key] = value
			} else {
				d.Errorf("Error on tag definition TAG=VALUE [ %s ]", tag)
			}
		}
	} else {
		d.Warnf("No map detected in device")
	}
	// Init stats
	d.stats.Init("device", d.cfg.ID, d.TagMap, d.log)

	d.statsData.Lock()
	d.Stats = d.getBasicStats()
	d.statsData.Unlock()
	return nil
}

// InitCatalogVar Initialize Global Variables on the device
// No need to protect from reads because device is not yet referenced.
func (d *SnmpDevice) InitCatalogVar(globalmap map[string]interface{}) {
	// Init Device Custom Variables
	d.VarMap = make(map[string]interface{}, len(globalmap))
	// copy global map to device map
	for k, v := range globalmap {
		d.VarMap[k] = v
	}

	if len(d.cfg.DeviceVars) > 0 {
		for _, tag := range d.cfg.DeviceVars {
			s := strings.Split(tag, "=")
			if len(s) == 2 {
				key, value := s[0], s[1]
				// check if exist
				if v, ok := d.VarMap[key]; ok {
					var err error
					switch v.(type) {

					case int64:
						d.VarMap[key], err = strconv.ParseInt(value, 10, 64)
					case string:
						d.VarMap[key] = value
					case float64:
						d.VarMap[key], err = strconv.ParseFloat(value, 64)
					}
					if err != nil {
						d.Errorf("There is an Error on the Type Conversion: %s ", err)
					}
				} else {
					d.Warnf("The Variable with KEY %s doens't exist in the  variable catalog ", key)
				}
			} else {
				d.Errorf("Error on Custom Variable definition VAR_NAME=VALUE [ %s ]", tag)
			}
		}
	} else {
		d.Warnf("No Custom Variables detected in device")
	}
}

// AttachToBus add this device to a communition bus
func (d *SnmpDevice) AttachToBus(b *bus.Bus) {
	d.Node = bus.NewNode(d.cfg.ID)
	b.Join(d.Node)
}

// LeaveBus add this device to a communition bus
func (d *SnmpDevice) LeaveBus(b *bus.Bus) {
	b.Leave(d.Node)
}

// End The Opposite of Init() uninitialize all variables
func (d *SnmpDevice) End() {
	d.Node.Close()
	// TODO que debería hacer esta variable.
	// Cuando se llama desde ReleaseDevices parece que solo se quiere liberar las conex snmp (que si estamos cerrando, y siendo UDP, no se si tiene mucho sentido).
	// Cuando se llama desde DeleteDeviceInRunTime ya se ha hecho un StopGather, está función sería la que cerrase las conex?
	// En cualquier caso dejo comentado el for porque d.snmpClientMap ya no existe
	/*
		for _, val := range d.snmpClientMap {
			val.Release()
		}
	*/
	// release files
	// os.Close(d.log.Out)
	// release snmp resources
}

// SetSelfMonitoring set the output device where send monitoring metrics
func (d *SnmpDevice) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	d.selfmon = cfg
	d.stats.SetSelfMonitoring(cfg)
}

func (d *SnmpDevice) firstSnmpConnect(connectionParams snmp.ConnectionParams) bool {
	var connected bool
	snmpClient := snmp.Client{
		ID:               d.cfg.Host,
		DisableBulk:      d.cfg.DisableBulk,
		ConnectionParams: connectionParams,
		Log:              d.log,
	}
	sysinfo, err := snmpClient.Connect(d.cfg.SystemOIDs)
	if err != nil {
		d.Errorf("unable to connect")
		d.DeviceConnected = false
		connected = false
	} else {
		// Avoid data race while modifying d.SysInfo and d.Measurements (in InitDevMeasurements)
		d.rtData.Lock()
		d.SysInfo = sysinfo
		d.rtData.Unlock()
		connected = true
		// send counters when device active and not connected ( no reset needed only status fields/tags are sen
		// d.stats.Send() // TODO gestion de stats
	}
	d.stats.SetStatus(d.DeviceActive, d.DeviceConnected)
	return connected
}

// StartGather Main GoRutine method to begin snmp data collecting
func (d *SnmpDevice) StartGather() {
	d.Infof("Initializating gather process for device on host (%s)", d.cfg.Host)

	// Define a default value for maxOids if its zero
	maxOids := d.cfg.MaxOids
	if maxOids <= 0 {
		maxOids = DEFAULT_MAX_OIDS
	}

	// Organize the config needed to establish a SNMP connection.
	// Will be used when a new snmp connection is needed.
	connectionParams := snmp.ConnectionParams{
		Host:           d.cfg.Host,
		Port:           d.cfg.Port,
		Timeout:        d.cfg.Timeout,
		Retries:        d.cfg.Retries,
		SnmpVersion:    d.cfg.SnmpVersion,
		Community:      d.cfg.Community,
		MaxRepetitions: d.cfg.MaxRepetitions,
		MaxOids:        maxOids,
		Debug:          d.cfg.SnmpDebug,
		V3Params: snmp.V3Params{
			SecLevel:        d.cfg.V3SecLevel,
			AuthUser:        d.cfg.V3AuthUser,
			AuthPass:        d.cfg.V3AuthPass,
			PrivPass:        d.cfg.V3PrivPass,
			PrivProt:        d.cfg.V3PrivProt,
			AuthProt:        d.cfg.V3AuthProt,
			ContextName:     d.cfg.V3ContextName,
			ContextEngineID: d.cfg.V3ContextEngineID,
		},
	}

	// Check if the values are valid, for example, if we have a community if the connection is v2c
	err := connectionParams.Validation()
	if err != nil {
		d.log.Errorf("SNMP parameter validation: %v", err)
		return
	}

	// Wait for the device to become active
	// This loop will wait for commands and finish if we device recive the command to exit or to activate
	// Others commands will be logged and ignored
	/* 	for !d.DeviceActive {
	   		select {
	   		case val := <-d.Node.Read:
	   			d.Infof("Device %s received message while waiting state. %s: %+v", d.cfg.Host, val.Type, val.Data)
	   			switch val.Type {
	   			case bus.Exit:
	   				d.Infof("invoked Asyncronous EXIT from SNMP Gather process ")
	   				return
	   			case bus.SyncExit:
	   				d.Infof("invoked Syncronous EXIT from SNMP Gather process ")
	   				d.isStopped <- true
	   				return
	   			case bus.Enabled:
	   				enabled, ok := val.Data.(bool)
	   				if !ok {
	   					d.Errorf("invalid value for enabled bus message: %v", val.Data)
	   					continue
	   				}

	   				d.rtData.Lock()
	   				d.DeviceActive = enabled
	   				d.rtData.Unlock()
	   				// TODO gestionar stats
	   				if enabled {
	   					d.stats.SetActive(true)
	   				} else {
	   					d.stats.SetStatus(false, false)
	   				}
	   				d.Infof("Device %s STATUS ACTIVE [%t] ", d.cfg.Host, enabled)
	   			default:
	   				d.Warnf("Device %s in waiting state. Ignored command: %s (%s)", d.cfg.Host, val.Type, val.Data)
	   			}
	   		}
	   	}
	*/
	d.Infof("Device on host (%s) is active. Trying to connect ", d.cfg.Host)

	// Try to establish the first connection. Loop will be finished when the connection has been established
	/* 	for {
		d.log.Infof("YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY")
		// If device receive a command to deactivate, stop trying to connect
		if d.DeviceActive {
			snmpClient := snmp.Client{
				ID:               d.cfg.Host,
				DisableBulk:      d.cfg.DisableBulk,
				ConnectionParams: connectionParams,
				Log:              d.log,
			}
			sysinfo, err := snmpClient.Connect(d.cfg.SystemOIDs)
			if err != nil {
				d.Errorf("unable to connect")
			} else {
				// Avoid data race while modifying d.SysInfo and d.Measurements (in InitDevMeasurements)
				d.rtData.Lock()
				d.SysInfo = sysinfo
				d.InitDevMeasurements()
				d.rtData.Unlock()
				break
				// send counters when device active and not connected ( no reset needed only status fields/tags are sen
				// d.stats.Send() // TODO gestion de stats
			}
		}

		// Wait device freq to retry to reconnect and see if there a new messages on the bus to be processed
		select {
		case <-time.NewTimer(time.Duration(d.cfg.Freq) * time.Second).C:

			// Try to reconnect after d.cfg.Freq seconds
		case val := <-d.Node.Read:
			d.Infof("Device %s received message while waiting state. %s: %+v", d.cfg.Host, val.Type, val.Data)
			switch val.Type {
			case bus.Exit:
				d.Infof("invoked Asyncronous EXIT from SNMP Gather process ")
				return
			case bus.SyncExit:
				d.Infof("invoked Syncronous EXIT from SNMP Gather process ")
				d.isStopped <- true
				return
			case bus.Enabled:
				status := val.Data.(bool)
				d.rtData.Lock()
				d.DeviceActive = status
				d.rtData.Unlock()
				if status {
					d.stats.SetActive(true)
				} else {
					d.stats.SetStatus(false, false)
				}

				d.Infof("Device %s STATUS ACTIVE [%t] ", d.cfg.Host, status)
			default:
				d.Warnf("Device %s in connecting state. Ignored command: %s (%s)", d.cfg.Host, val.Type, val.Data)
			}
		}
	} */

	// Device is active and connected

	d.InitDevMeasurements()

	// Create a bus to control all goroutines created to manage this device
	deviceControlBus := bus.NewBus()
	go deviceControlBus.Start()

	// Control when all goroutines have finished
	var deviceWG sync.WaitGroup

	// gatherLock will be used to gather data sequentially is cfg.ConcurrentGather is true.
	// If false, it will be set to nil and ignored by the measurements goroutines while gathering data.
	gatherLock := &sync.Mutex{}
	if d.cfg.ConcurrentGather {
		gatherLock = nil
	}

	for _, meas := range d.Measurements {
		// Start gather goroutine for device and add it to the wait group for gather goroutines
		deviceWG.Add(1)
		go func(m *measurement.Measurement) {
			defer deviceWG.Done()

			identifier := fmt.Sprintf("%s-%s", d.cfg.ID, m.ID)

			// Add the measurement as a node to the bus
			node := bus.NewNode(identifier)
			deviceControlBus.Join(node)

			// Create the SNMP client for each measurement.
			// This client is just the data needed to connect, it does not start any connection yet.
			// Here is created just the building blocks to be able to create the goSNMP client.
			// We leave to the Measurement to handle the creation and destruction of that client.
			snmpClient := snmp.Client{
				ID:               identifier,
				DisableBulk:      d.cfg.DisableBulk,
				ConnectionParams: connectionParams,
				Log:              m.Log,
			}

			// Start the loop that will gather metrics and handle signals
			m.GatherLoop(node, snmpClient, d.Freq, d.cfg.UpdateFltFreq, d.VarMap, d.TagMap, d.cfg.SystemOIDs, d.Influx, gatherLock)
		}(meas)
	}

	if d.DeviceActive {
		d.firstSnmpConnect(connectionParams)
	}

	deviceTicker := time.NewTimer(time.Duration(d.cfg.Freq) * time.Second)
	// Wait for commands
	for {
		select {
		case <-deviceTicker.C:
			if d.DeviceActive && !d.DeviceConnected {
				// connect
				d.firstSnmpConnect(connectionParams)
			}

			// Some online actions can change Stats
			d.statsData.Lock()
			d.Stats = d.getBasicStats()
			d.statsData.Unlock()
			d.stats.Send()
			d.stats.ResetCounters()
			// Try to reconnect after d.cfg.Freq seconds
		case val := <-d.Node.Read:
			d.Infof("Received Message: %s (%+v)", val.Type, val.Data)
			switch val.Type {
			case bus.Exit:
				d.log.Infof("invoked asyncronous EXIT from SNMP Gather process ")
				// This broadcast is blocking, will wait till all measurement goroutines have received the message
				deviceControlBus.Broadcast(&bus.Message{Type: val.Type})
				return
			case bus.SyncExit:
				d.log.Infof("invoked syncronous EXIT from SNMP Gather process ")
				// Signal all measurement goroutines to exit
				deviceControlBus.Broadcast(&bus.Message{Type: val.Type})
				// Wait till all measurement goroutines have finished
				deviceWG.Wait()
				// Signal the caller of this command that all have finished correctly
				d.isStopped <- true
				return
			case bus.LogLevel:
				level := val.Data.(string)
				l, err := logrus.ParseLevel(level)
				if err != nil {
					d.Warnf("ERROR on Changing LOGLEVEL to [%t] ", level)
					break
				}
				d.rtData.Lock()
				// Changing log level here affects all "child" loggers (those passed to measurements goroutines)
				d.log.Level = l
				d.Infof("device loglevel Changed  [%s] ", level)
				d.CurLogLevel = d.log.Level.String()
				d.rtData.Unlock()

			default: // exit, snmpresethard, snmpdebug, setsnmpmaxrep, forcegather, enabled, filterupdate
				d.Infof("invoked %+v, passing message to measurements", val)
				// Blocking operation. Waits till all measurements have received it
				deviceControlBus.Broadcast(val)
				d.Infof("messaged %+v received by all measurements", val)

			}
		}
	}
}
