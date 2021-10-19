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
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
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

	// SNMP and Influx Clients config
	// TODO borrar esto? Mejor que cada gorutina gestione su conex, no en el structs
	// El cliente snmp no es thread safe: ttps://github.com/gosnmp/gosnmp/issues/64#issuecomment-231645164
	snmpClientMap map[string]*snmp.Client
	Influx        *output.InfluxDB `json:"-"`
	// LastError     time.Time
	// Runtime stats
	stats DevStat  // Runtime Internal statistic
	Stats *DevStat // Public info for thread safe accessing to the data ()

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
	// TODO coger un read lock de cada measurement, para poder independicar los measurements y que cada uno
	// pueda escribir en sus datos sin bloqueos
	d.rtData.RLock()
	defer d.rtData.RUnlock()
	result, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		d.Errorf("Error on Get JSON data from device")
		dummy := []byte{}
		return dummy, nil
	}
	return result, err
}

// GetBasicStats get basic info for this device
func (d *SnmpDevice) GetBasicStats() *DevStat {
	d.statsData.RLock()
	defer d.statsData.RUnlock()
	return d.Stats
}

// GetBasicStats get basic info for this device
func (d *SnmpDevice) getBasicStats() *DevStat {
	sum := 0
	for _, m := range d.Measurements {
		sum += len(m.OidSnmpMap)
	}
	stat := d.stats.ThSafeCopy()
	stat.TagMap = d.TagMap
	stat.NumMeasurements = len(d.Measurements)
	stat.NumMetrics = sum
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
				// creating a new measurement runtime object and asigning to array

				// TODO pasar un logger ya específico para este host y este measurement
				// Pasar este measLog en vez de d.log
				/*
					measLog := d.log.WithFields(logrus.Fields{
						"host": d.cfg.Host,
						"measurement": mVal,
					})
					imeas := measurement.New(mVal, d.cfg.MeasFilters, cfg.MFilters, measLog)
				*/
				// TODO pasamos MeasFilters y MFilters porque lo necesitará cuando haga el InitFilters (antes se hacia en un bucle debajo de este for)
				// TODO no pasar la config del measurement, si no una copia.
				// Así evitamos que cada gorutina encargada de cada measurement en cada device
				// puedan intentar modificarla al mismo tiempo (Lo hace SnmpMetric.Init)
				imeas := measurement.New(mVal, d.cfg.MeasFilters, cfg.MFilters, d.log)
				d.Measurements = append(d.Measurements, imeas)
			}
		}
	}

	// Initialize all snmpMetrics  objects and OID array
	// get data first time
	// useful to inicialize counter all value and test device snmp availability
}

// this method puts all metrics as invalid once sent to the backend
// it lets us to know if any of them has not been updated in the gathering process
func (d *SnmpDevice) invalidateMetrics() {
	for _, v := range d.Measurements {
		v.InvalidateMetrics()
	}
}

/*
Init  does the following

- Initialize not set variables to some defaults
- Initialize logfile for this device
- Initialize comunication channels and initial device state
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

	d.TagMap = make(map[string]string)
	if len(d.cfg.DeviceTagName) == 0 {
		d.cfg.DeviceTagName = "device"
	}

	d.Freq = d.cfg.Freq

	d.snmpClientMap = make(map[string]*snmp.Client)

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
	d.stats.Init(d.cfg.ID, d.TagMap, d.log)

	d.statsData.Lock()
	d.Stats = d.getBasicStats()
	d.statsData.Unlock()
	return nil
}

// InitCatalogVar Initialize Global Variables on the device
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
	for _, val := range d.snmpClientMap {
		val.Release()
	}
	// release files
	// os.Close(d.log.Out)
	// release snmp resources
}

// SetSelfMonitoring set the output device where send monitoring metrics
func (d *SnmpDevice) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	d.stats.SetSelfMonitoring(cfg)
}

// CheckDeviceConnectivity check if device snmp connection is ok by checking SnmpOIDGetProcessed stats
func (d *SnmpDevice) CheckDeviceConnectivity() {
	ProcessedStat := d.stats.GetCounter(SnmpOIDGetProcessed)

	if value, ok := ProcessedStat.(int); ok {
		// check if no processed SNMP data (when this happens means there is not connectivity with the device )
		if value == 0 {
			d.DeviceConnected = false
			d.stats.SetStatus(d.DeviceActive, false)
		}
	} else {
		d.Warnf("Error in check Processd Stats %#+v ", ProcessedStat)
	}
}

// TODO borrar esto? Mejor que cada gorutina gestione su conex, no en el struct
func (d *SnmpDevice) snmpRelease() {
	for _, v := range d.snmpClientMap {
		if v != nil {
			d.Infof("Releasing snmp connection for %s", "init")
			v.Release()
		}
	}
}

// TODO borrar esto? Mejor que cada gorutina gestione su conex, no en el struct
func (d *SnmpDevice) releaseClientMap() {
	if !d.cfg.ConcurrentGather {
		if val, ok := d.snmpClientMap["init"]; ok {
			if val != nil {
				d.Infof("Releasing snmp connection for %s", "init")
				val.Release()
			}
		}
	} else {
		for k, val := range d.snmpClientMap {
			if val != nil {
				d.Infof("Releasing snmp connection for %s", k)
				val.Release()
			}
		}
	}
	// begin reset process
	d.snmpClientMap = make(map[string]*snmp.Client)
}

// TODO implementar en los measurements
/*
func (d *SnmpDevice) snmpReset(debug bool, maxrep uint8) {
	// On sequential we need release connection first , concurrent has an automatic self release system
	d.Infof("Reseting snmp connections DEBUG  ACTIVE  [%t] ", debug)
	d.releaseClientMap()
	initerrors := 0
	if d.cfg.ConcurrentGather == false {
		c, err := d.InitSnmpConnect("init", debug, maxrep)
		if err == nil {
			for _, m := range d.Measurements {
				m.SetSnmpClient(c)
			}
		} else {
			d.Errorf("Error on reset snmp connection  on device %s: disconnecting now...  ", d.cfg.ID)
			d.DeviceConnected = false
			d.stats.SetStatus(d.DeviceActive, false)
		}
	} else {
		for _, m := range d.Measurements {
			c, err := d.InitSnmpConnect(m.ID, debug, maxrep)
			if err != nil {
				d.Warnf("Error on recreate connection without debug for measurement %s", m.ID)
				initerrors++
			} else {
				m.SetSnmpClient(c)
			}
		}
		if initerrors > 0 {
			d.Warnf("Error on reset snmp connection for %d  measurements", initerrors)
		}
		if initerrors == len(d.Measurements) {
			d.Errorf("Error on reset snmp connection all (%d) measurements without valid connection : disconnecting now...  ", initerrors)
			d.DeviceConnected = false
			d.stats.SetStatus(d.DeviceActive, false)
		}
	}
}
*/

// StartGather Main GoRutine method to begin snmp data collecting
func (d *SnmpDevice) StartGather() {
	d.Infof("Initializating gather process for device on host (%s)", d.cfg.Host)
	// Wait for the device to become active
	// This loop will wait for commands and finish if we device recive the command to exit or to activate
	// Others commands will be logged and ignored
	for !d.DeviceActive {
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
				status := val.Data.(bool)
				d.rtData.Lock()
				d.DeviceActive = status
				if status {
					d.stats.SetActive(true)
				} else {
					d.stats.SetStatus(false, false)
				}
				d.Infof("Device %s STATUS ACTIVE [%t] ", d.cfg.Host, status)
				d.rtData.Unlock()
			default:
				d.Warnf("Device %s in waiting state. Ignored command: %s (%s)", d.cfg.Host, val.Type, val.Data)
			}
		}
	}

	d.Infof("Device on host (%s) is active. Trying to connect ", d.cfg.Host)

	// Try to establish the first connection. Loop will be finished when the connection has been established
	for {
		// If device receive a command to deactivate, stop trying to connect
		if d.DeviceActive {
			// This will try to connect to the device, gather SysInfo and set DeviceConnected to true
			// This connection will be stored on d.snmpClientMap["init"]
			/* TODO Necesitamos esta conex "init" para algo?
			_, err := d.InitSnmpConnect("init", d.cfg.SnmpDebug, 0)
			if err == nil {
				startSnmp := time.Now()
				// Create data structures to store data and connect to the device go gather system info
				// TODO esto ahora mismo inicializa todo, partirlo para llevarlo a la gorutina del meas
				d.InitDevMeasurements()
				elapsedSnmp := time.Since(startSnmp)
				d.stats.SetFltUpdateStats(startSnmp, elapsedSnmp)
				d.Infof("snmp INIT runtime measurements/filters took [%s] ", elapsedSnmp)
				// Connection established, progress
				break
			}
			*/
			d.InitDevMeasurements()
			break
			// send counters when device active and not connected ( no reset needed only status fields/tags are sen
			// d.stats.Send() // TODO gestion de stats
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
				if status {
					d.stats.SetActive(true)
				} else {
					d.stats.SetStatus(false, false)
				}

				d.Infof("Device %s STATUS ACTIVE [%t] ", d.cfg.Host, status)
				d.rtData.Unlock()
			default:
				d.Warnf("Device %s in connecting state. Ignored command: %s (%s)", d.cfg.Host, val.Type, val.Data)
			}
		}
	}
	// Device is active and connected

	// Create a bus to control all goroutines created to manage this device
	deviceControlBus := bus.NewBus()

	// Control when all goroutines have finished
	var deviceWG sync.WaitGroup

	for _, meas := range d.Measurements {
		// Start gather goroutine for device and add it to the wait group for gather goroutines
		deviceWG.Add(1)
		go func(m *measurement.Measurement) {
			defer deviceWG.Done()

			// Add the measurement as a node to the bus
			node := bus.NewNode(fmt.Sprintf("%s-%s", d.cfg.ID, m.ID))
			deviceControlBus.Join(node)

			m.GatherLoop(
				node,
				d.Freq,
				d.cfg.UpdateFltFreq,
				d.cfg.SnmpDebug,
				d.cfg.Host,
				d.cfg.MaxRepetitions,
				d.cfg.SnmpVersion,
				d.cfg.Community,
				d.cfg.Port,
				d.cfg.Timeout,
				d.cfg.Retries,
				d.cfg.V3AuthUser,
				d.cfg.V3SecLevel,
				d.cfg.V3AuthPass,
				d.cfg.V3PrivPass,
				d.cfg.V3PrivProt,
				d.cfg.V3AuthProt,
				d.cfg.V3ContextName,
				d.cfg.V3ContextEngineID,
				d.cfg.ID,
				d.cfg.SystemOIDs,
				d.cfg.MaxOids,
				d.cfg.DisableBulk,
				d.VarMap,
				d.TagMap,
				d.Influx,
			)
		}(meas)
	}

	// Check if there is some message in the bus to be processed.
	for {
		select {
		case val := <-d.Node.Read:
			d.Infof("Received Message: %s (%+v)", val.Type, val.Data)
			switch val.Type {
			case bus.SyncExit:
				d.Infof("invoked Syncronous EXIT from SNMP Gather process ")
				// Signal all measurement goroutines to exit
				deviceControlBus.Broadcast(&bus.Message{Type: val.Type})
				// Wait till all measurement goroutines have finished
				deviceWG.Wait()
				// Signal the caller of this command that all have finished correctly
				d.isStopped <- true
				return
			case bus.LogLevel:
				// TODO tenemos que pasarle algo a los measurements para cambiar su loglevel?
				level := val.Data.(string)
				l, err := logrus.ParseLevel(level)
				if err != nil {
					d.Warnf("ERROR on Changing LOGLEVEL to [%t] ", level)
					break
				}
				d.rtData.Lock()
				d.log.Level = l
				d.Infof("device loglevel Changed  [%s] ", level)
				d.CurLogLevel = d.log.Level.String()
				d.rtData.Unlock()

			default: // exit, snmpresethard, snmpdebug, setsnmpmaxrep, forcegather, enabled, filterupdate
				d.Infof("invoked %s, passing message to measurements", val)
				deviceControlBus.Broadcast(val)

			}
		}

		// Some online actions can change Stats
		d.statsData.Lock()
		d.Stats = d.getBasicStats()
		d.statsData.Unlock()

	}
}
