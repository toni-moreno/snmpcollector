package device

import (
	"fmt"
	"os"
	"strconv"

	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/agent/bus"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/measurement"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

var (
	cfg    *config.SQLConfig
	logDir string
)

// SetDBConfig set agent config
func SetDBConfig(c *config.SQLConfig) {
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
	//basic sistem info
	SysInfo *snmp.SysInfo
	//runtime built TagMap
	TagMap map[string]string
	//Refresh data to show in the frontend
	Freq int
	//Measurements array
	Measurements []*measurement.Measurement
	//Variable map
	VarMap map[string]interface{}

	//SNMP and Influx Clients config
	//snmpClient *gosnmp.GoSNMP
	snmpClientMap map[string]*gosnmp.GoSNMP
	Influx        *output.InfluxDB `json:"-"`
	//LastError     time.Time
	//Runtime stats
	stats DevStat  //Runtime Internal statistic
	Stats *DevStat //Public info for thread safe accessing to the data ()

	//runtime controls
	rtData             sync.RWMutex
	statsData          sync.RWMutex
	ReloadLoopsPending int

	DeviceActive    bool
	DeviceConnected bool
	StateDebug      bool

	Node *bus.Node `json:"-"`

	CurLogLevel     string
	Gather          func()                                                              `json:"-"`
	InitSnmpConnect func(mkey string, debug bool, maxrep uint8) (*gosnmp.GoSNMP, error) `json:"-"`
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
	stat.ReloadLoopsPending = d.ReloadLoopsPending
	stat.TagMap = d.TagMap
	stat.DeviceActive = d.DeviceActive
	stat.DeviceConnected = d.DeviceConnected
	stat.NumMeasurements = len(d.Measurements)
	stat.NumMetrics = sum
	if d.SysInfo != nil {
		stat.SysDescription = d.SysInfo.SysDescr
	} else {
		stat.SysDescription = ""
	}
	return stat
}

func (d *SnmpDevice) setReloadLoopsPending(val int) {
	d.ReloadLoopsPending = val
}

func (d *SnmpDevice) getReloadLoopsPending() int {
	return d.ReloadLoopsPending
}

func (d *SnmpDevice) decReloadLoopsPending() {
	if d.ReloadLoopsPending > 0 {
		d.ReloadLoopsPending--
	}
}

// GetOutSenderFromMap to get info about the sender will use
func (d *SnmpDevice) GetOutSenderFromMap(influxdb map[string]*output.InfluxDB) (*output.InfluxDB, error) {
	if len(d.cfg.OutDB) == 0 {
		d.Warnf("No OutDB configured on the device")
	}
	var ok bool
	name := d.cfg.OutDB
	if d.Influx, ok = influxdb[name]; !ok {
		//we assume there is always a default db
		if d.Influx, ok = influxdb["default"]; !ok {
			//but
			return nil, fmt.Errorf("No influx config for snmp device: %s", d.cfg.ID)
		}
	}

	return d.Influx, nil
}

// ForceGather send message to force a data gather execution
func (d *SnmpDevice) ForceGather() {
	d.Node.SendMsg(&bus.Message{Type: "forcegather"})
}

// ForceFltUpdate send info to update the filter counter to the next execution
func (d *SnmpDevice) ForceFltUpdate() {
	d.Node.SendMsg(&bus.Message{Type: "filterupdate"})
}

// SnmpReset send message to init an  SNMP connection reset could be soft/hard
func (d *SnmpDevice) SnmpReset(mode string) {
	switch mode {
	case "hard":
		d.Node.SendMsg(&bus.Message{Type: "snmpresethard"})
	case "soft":
		d.Node.SendMsg(&bus.Message{Type: "snmpreset"})
	default:
		d.log.Infof("Unknown mode %s on SNMPRESET ", mode)
	}

}

// StopGather send signal to stop the Gathering process
func (d *SnmpDevice) StopGather() {
	d.Node.SendMsg(&bus.Message{Type: "exit"})
}

//RTActivate change activatio state in runtime
func (d *SnmpDevice) RTActivate(activate bool) {
	d.Node.SendMsg(&bus.Message{Type: "enabled", Data: activate})
}

//RTActSnmpDebug change snmp debug runtime
func (d *SnmpDevice) RTActSnmpDebug(activate bool) {
	d.Node.SendMsg(&bus.Message{Type: "snmpdebug", Data: activate})
}

//RTActSnmpMaxRep change snmp MaxRepetitions
func (d *SnmpDevice) RTActSnmpMaxRep(maxrep uint8) {
	d.Node.SendMsg(&bus.Message{Type: "setsnmpmaxrep", Data: maxrep})
}

// RTSetLogLevel set the log level for this device
func (d *SnmpDevice) RTSetLogLevel(level string) {
	d.Node.SendMsg(&bus.Message{Type: "loglevel", Data: level})
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

	//Alloc array
	d.Measurements = make([]*measurement.Measurement, 0, 0)
	d.Debugf("---Init device measurements from groups %s------------------", d.cfg.Host)
	//for this device get MeasurementGroups and search all measurements

	for _, devMeas := range d.cfg.MeasurementGroups {
		//Selecting all Metric Groups that matches with device.MeasurementGroups
		selGroups := make(map[string]*config.MGroupsCfg, 0)
		//var RegExp = regexp.MustCompile(devMeas)
		for key, val := range cfg.GetGroups {
			if key == devMeas {
				selGroups[key] = val
			}
		}
		d.Debugf("this device has this SELECTED GROUPS: %+v", selGroups)
		//Only For selected Groups we will get all selected measurements and we will remove repeated values
		var selMeas []string
		for key, val := range selGroups {
			d.Debugf("Selecting from group %s", key)
			for _, item := range val.Measurements {
				d.Debugf("Selecting measurements  %s from group %s", item, key)
				selMeas = append(selMeas, item)
			}
		}
		//remove duplicated measurements if needed
		selMeasUniq := utils.RemoveDuplicatesUnordered(selMeas)
		//Now we know what measurements names  will send influx from this device

		d.Debugf("DEVICE MEASUREMENT: %s HOST: %s ", devMeas, d.cfg.Host)
		for _, val := range selMeasUniq {
			//check if measurement exist
			if mVal, ok := cfg.Measurements[val]; !ok {
				d.Warnf("no measurement configured with name %s in host : %s", val, d.cfg.Host)
			} else {
				d.Debugf("MEASUREMENT CFG KEY: %s VALUE %s | Connection [%s] %+v", val, mVal.Name, val, d.snmpClientMap[mVal.ID])
				//
				c, err := d.InitSnmpConnect(mVal.ID, d.cfg.SnmpDebug, 0)
				if err != nil {
					d.Errorf("Error on snmpconnection initialization on measurement %s : Error: %s", mVal.ID, err)
					continue
				}
				//creating a new measurement runtime object and asigning to array
				imeas, err := measurement.New(mVal, d.log, c, d.cfg.DisableBulk)
				if err != nil {
					d.Errorf("Error on measurement initialization  Error: %s", err)
					continue
				}
				d.Measurements = append(d.Measurements, imeas)
			}
		}
	}

	/*For each  measurement look for filters and  Add to the measurement with this Filter after it initializes the runtime for the measurement  	*/

	for _, m := range d.Measurements {
		//check for filters associated with this measurement
		var mfilter *config.MeasFilterCfg
		for _, f := range d.cfg.MeasFilters {
			//we search if exist in the filter Database
			if filter, ok := cfg.MFilters[f]; ok {
				if filter.IDMeasurementCfg == m.ID {
					mfilter = filter
					break
				}
			}
		}
		if mfilter != nil {
			d.Debugf("filters %s found for device  and measurement %s ", mfilter.ID, m.ID)
			err := m.AddFilter(mfilter)
			if err != nil {
				d.Errorf("Error on initialize Filter for Measurement %s , Error:%s no data will be gathered for this measurement", m.ID, err)
			}
		} else {
			d.Debugf("no filters found for device on measurement %s", m.ID)
		}
		//Initialize internal structs after
		m.InitBuildRuntime()
		//Get Data First Time ( useful for counters)
		m.GetData()
	}
	//Initialize all snmpMetrics  objects and OID array
	//get data first time
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
	//log.Infof("Initializing device %s\n", d.cfg.ID)

	//Init Logger
	if d.cfg.Freq == 0 {
		d.cfg.Freq = 60
	}
	if len(d.cfg.LogFile) == 0 {
		d.cfg.LogFile = logDir + "/" + d.cfg.ID + ".log"

	}
	if len(d.cfg.LogLevel) == 0 {
		d.cfg.LogLevel = "info"
	}

	f, _ := os.OpenFile(d.cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	d.log = logrus.New()
	d.log.Out = f
	l, _ := logrus.ParseLevel(d.cfg.LogLevel)
	d.log.Level = l
	d.CurLogLevel = d.log.Level.String()
	//Formatter for time
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	d.log.Formatter = customFormatter
	customFormatter.FullTimestamp = true

	d.StateDebug = d.cfg.SnmpDebug

	d.setReloadLoopsPending(d.cfg.UpdateFltFreq)

	d.DeviceActive = d.cfg.Active

	//Init Device Tags

	d.TagMap = make(map[string]string)
	if len(d.cfg.DeviceTagName) == 0 {
		d.cfg.DeviceTagName = "device"
	}

	d.Freq = d.cfg.Freq

	d.snmpClientMap = make(map[string]*gosnmp.GoSNMP)

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

	if d.cfg.ConcurrentGather == true {
		d.Gather = d.measConcurrentGatherAndSend
		d.InitSnmpConnect = d.initSnmpConnectConcurrent
	} else {
		d.Gather = d.measSeqGatherAndSend
		d.InitSnmpConnect = d.initSnmpConnectSequential
	}
	d.statsData.Lock()
	d.Stats = d.getBasicStats()
	d.statsData.Unlock()
	return nil
}

// InitVars Initialize Global Variables on the device
func (d *SnmpDevice) InitCatalogVar(globalmap map[string]interface{}) {
	// Init Device Custom Variables
	d.VarMap = make(map[string]interface{}, len(globalmap))
	//copy global map to device map
	for k, v := range globalmap {
		d.VarMap[k] = v
	}

	if len(d.cfg.DeviceVars) > 0 {
		for _, tag := range d.cfg.DeviceVars {
			s := strings.Split(tag, "=")
			if len(s) == 2 {
				key, value := s[0], s[1]
				//check if exist
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

// End The Opposite of Init() uninitialize all variables
func (d *SnmpDevice) End() {
	d.Node.Close()
	for _, val := range d.snmpClientMap {
		snmp.Release(val)
	}
	//release files
	//os.Close(d.log.Out)
	//release snmp resources
}

// SetSelfMonitoring set the output device where send monitoring metrics
func (d *SnmpDevice) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	d.stats.SetSelfMonitoring(cfg)
}

// initSnmpConnectConcurrent does the  SNMP client connection and retrieve system info
func (d *SnmpDevice) initSnmpConnectConcurrent(mkey string, debug bool, maxrep uint8) (*gosnmp.GoSNMP, error) {
	if val, ok := d.snmpClientMap[mkey]; ok {
		if val != nil {
			d.Infof("Releaseing SNMP connection for measurement %s", mkey)
			snmp.Release(val)
		}
	}
	d.Infof("Beginning SNMP connection for measurement %s", mkey)
	client, sysinfo, err := snmp.GetClient(d.cfg, d.log, mkey, debug, maxrep)
	if err != nil {
		d.DeviceConnected = false
		d.Errorf("Client connect error to device  error :%s", err)
		d.snmpClientMap[mkey] = nil
		return nil, err
	}

	d.Infof("SNMP connection stablished Successfully for device  and measurement %s", mkey)
	d.snmpClientMap[mkey] = client
	d.SysInfo = sysinfo
	d.DeviceConnected = true
	return client, nil
}

// initSnmpConnectConcurrent does the  SNMP client connection and retrieve system info
func (d *SnmpDevice) initSnmpConnectSequential(mkey string, debug bool, maxrep uint8) (*gosnmp.GoSNMP, error) {
	//in sequential this
	if val, ok := d.snmpClientMap["init"]; ok {
		if val != nil {
			d.Infof("Previous SNMP connection found")
			d.snmpClientMap[mkey] = val
			return val, nil
		}
	}
	d.Infof("Beginning SNMP connection Sequential")
	client, sysinfo, err := snmp.GetClient(d.cfg, d.log, mkey, debug, maxrep)
	if err != nil {
		d.DeviceConnected = false
		d.Errorf("Client connect error to device  error :%s", err)
		d.snmpClientMap[mkey] = nil
		return nil, err
	}

	d.Infof("SNMP connection stablished Successfully for device  and measurement %s", mkey)
	d.snmpClientMap[mkey] = client
	d.SysInfo = sysinfo
	d.DeviceConnected = true
	return client, nil
}

// CheckDeviceConnectivity check if device snmp connection is ok by checking SnmpOIDGetProcessed stats
func (d *SnmpDevice) CheckDeviceConnectivity() {

	ProcessedStat := d.stats.GetCounter(SnmpOIDGetProcessed)

	if value, ok := ProcessedStat.(int); ok {
		//check if no processed SNMP data (when this happens means there is not connectivity with the device )
		if value == 0 {
			d.DeviceConnected = false
		}
	} else {
		d.Warnf("Error in check Processd Stats %#+v ", ProcessedStat)
	}
}

func (d *SnmpDevice) snmpRelease() {
	for _, v := range d.snmpClientMap {
		if v != nil {
			d.Infof("Releasing snmp connection for %s", "init")
			snmp.Release(v)
		}
	}
}

func (d *SnmpDevice) snmpReset(debug bool, maxrep uint8) {
	//On sequential we need release connection first , concurrent has an automatic self release system
	d.Infof("Reseting snmp connections DEBUG  ACTIVE  [%t] ", debug)
	if !d.cfg.ConcurrentGather {
		if val, ok := d.snmpClientMap["init"]; ok {
			if val != nil {
				d.Infof("Releasing snmp connection for %s", "init")
				snmp.Release(val)
			}
		}
	}
	//begin reset process
	d.snmpClientMap = make(map[string]*gosnmp.GoSNMP)
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
		}
	}
}

func (d *SnmpDevice) gatherAndProcessData(t *time.Ticker, force bool) *time.Ticker {
	d.rtData.Lock()
	//if active
	if d.DeviceActive || force {
	FORCEINIT:
		//check if device has active snmp connections and Initialize if not
		if d.DeviceConnected == false {
			_, err := d.InitSnmpConnect("init", d.cfg.SnmpDebug, 0)
			if err == nil {
				startSnmp := time.Now()
				d.InitDevMeasurements()
				elapsedSnmp := time.Since(startSnmp)
				d.stats.SetFltUpdateStats(startSnmp, elapsedSnmp)
				d.Infof("snmp INIT runtime measurements/filters took [%s] ", elapsedSnmp)
				if force == false {
					// Round collection to nearest interval by sleeping
					//and reprogram the ticker to aligned starts
					// only when no extra gather(forced from web-ui)
					utils.WaitAlignForNextCicle(d.cfg.Freq, d.log)
					t.Stop()
					t = time.NewTicker(time.Duration(d.cfg.Freq) * time.Second)
					//force one iteration now..after device has been connected  dont wait for next
					//ticker (1 complete cicle)
				}
				goto FORCEINIT
			}
		} else {
			//device active and connected
			d.Infof("Init gather cicle mode Concurrent [ %t ]", d.cfg.ConcurrentGather)
			/*************************
			 *
			 * SNMP Gather data process
			 *
			 ***************************/
			d.invalidateMetrics()
			d.stats.ResetCounters()
			d.Gather()

			/*******************************************
			 *
			 * Reload Indexes/Filters process(if needed)
			 *
			 *******************************************/
			//Check if reload needed with d.ReloadLoopsPending if a posivive value on negative this will disabled

			d.decReloadLoopsPending()

			if d.getReloadLoopsPending() == 0 {
				startIdxUpdateStats := time.Now()
				for _, m := range d.Measurements {
					if m.GetMode() == "value" {
						continue
					}
					changed, err := m.UpdateFilter()
					if err != nil {
						d.Errorf("Error on update Indexes/filter : ERR: %s", err)
						continue
					}
					if changed {
						m.InitBuildRuntime()
					}
				}

				d.setReloadLoopsPending(d.cfg.UpdateFltFreq)
				elapsedIdxUpdateStats := time.Since(startIdxUpdateStats)
				d.stats.SetFltUpdateStats(startIdxUpdateStats, elapsedIdxUpdateStats)
			}

			d.CheckDeviceConnectivity()

			d.stats.Send()
		}
	} else {
		d.Infof("Gather process is disabled")
	}
	//get Ready a copy of the stats to

	d.statsData.Lock()
	d.Stats = d.getBasicStats()
	d.statsData.Unlock()
	d.rtData.Unlock()
	return t
}

// StartGather Main GoRutine method to begin snmp data collecting
func (d *SnmpDevice) StartGather(wg *sync.WaitGroup) {
	wg.Add(1)
	go d.startGatherGo(wg)
}

func (d *SnmpDevice) startGatherGo(wg *sync.WaitGroup) {
	defer wg.Done()

	if d.DeviceActive && d.DeviceConnected {
		d.Infof("Begin first InidevInfo")
		startSnmp := time.Now()
		d.rtData.Lock()
		d.InitDevMeasurements()
		d.rtData.Unlock()
		elapsedSnmp := time.Since(startSnmp)
		d.stats.SetFltUpdateStats(startSnmp, elapsedSnmp)
		d.Infof("snmp INIT runtime measurements/filters took [%s] ", elapsedSnmp)
	} else {
		d.Infof("Can not initialize this device: Is Active: %t  |  Connection Active: %t ", d.DeviceActive, d.snmpClientMap != nil)
	}

	d.Infof("Beginning gather process for device on host (%s)", d.cfg.Host)

	t := time.NewTicker(time.Duration(d.cfg.Freq) * time.Second)
	for {

		t = d.gatherAndProcessData(t, false)

	LOOP:
		for {
			select {
			case <-t.C:
				break LOOP
			case val := <-d.Node.Read:
				d.Infof("Received Message...%s: %+v", val.Type, val.Data)
				switch val.Type {
				case "forcegather":
					d.Infof("invoked Force Data Gather And Process")
					d.gatherAndProcessData(t, true)
				case "exit":
					d.Infof("invoked EXIT from SNMP Gather process ")
					return
				case "filterupdate":
					d.rtData.Lock()
					d.setReloadLoopsPending(1)
					d.rtData.Unlock()
				case "snmpresethard":
					d.rtData.Lock()
					//when no connection availables on the first initialization
					//measurmentes should be initialized again
					d.snmpRelease()
					d.InitDevMeasurements()
					d.rtData.Unlock()
				case "snmpreset":
					d.rtData.Lock()
					d.snmpReset(false, 0)
					d.rtData.Unlock()
				case "snmpdebug":
					debug := val.Data.(bool)
					d.rtData.Lock()
					d.StateDebug = debug
					d.snmpReset(debug, 0)
					d.rtData.Unlock()
				case "setsnmpmaxrep":
					maxrep := val.Data.(uint8)
					d.rtData.Lock()
					d.snmpReset(false, maxrep)
					d.rtData.Unlock()
				case "enabled":
					status := val.Data.(bool)
					d.rtData.Lock()
					d.DeviceActive = status
					d.Infof("device STATUS  ACTIVE  [%t] ", status)
					d.rtData.Unlock()
				case "loglevel":
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
				}
			}
			//Some online actions can change Stats
			d.statsData.Lock()
			d.Stats = d.getBasicStats()
			d.statsData.Unlock()
		}
	}
}
