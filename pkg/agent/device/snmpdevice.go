package device

import (
	"fmt"
	"os"

	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/agent/selfmon"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/measurement"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

// DevStat minimal info to show users
type DevStat struct {
	Requests           int64
	Gets               int64
	Errors             int64
	ReloadLoopsPending int
	DeviceActive       bool
	DeviceConnected    bool
	NumMeasurements    int
	NumMetrics         int
}

var (
	cfg    *config.SQLConfig
	logDir string
)

// SetDBConfg set agent config
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

	//SNMP and Influx Clients config
	//snmpClient *gosnmp.GoSNMP
	snmpClientMap map[string]*gosnmp.GoSNMP
	Influx        *output.InfluxDB `json:"-"`
	LastError     time.Time
	//Runtime stats
	Requests              int64
	Gets                  int64
	Errors                int64
	LastGatherTime        time.Time
	LastGatherDuration    time.Duration
	LastFltUpdateTime     time.Time
	LastFltUpdateDuration time.Duration
	//runtime controls

	ReloadLoopsPending int

	DeviceActive    bool
	DeviceConnected bool
	StateDebug      bool

	chDebug     chan bool
	chEnabled   chan bool
	chLogLevel  chan string
	chExit      chan bool
	chFltUpdate chan bool
	mutex       sync.Mutex
	selfmon     *selfmon.SelfMon
	CurLogLevel string
	Gather      func() `json:"-"`
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

func (d *SnmpDevice) setGatherStats(start time.Time, duration time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.LastGatherTime = start
	d.LastGatherDuration = duration
}

func (d *SnmpDevice) setFltUpdateStats(start time.Time, duration time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.LastFltUpdateTime = start
	d.LastFltUpdateDuration = duration
}

//ReloadLoopPending needs to be mutex excluded

func (d *SnmpDevice) setReloadLoopsPending(val int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.ReloadLoopsPending = val
}

func (d *SnmpDevice) getReloadLoopsPending() int {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.ReloadLoopsPending
}

func (d *SnmpDevice) decReloadLoopsPending() {
	d.mutex.Lock()
	if d.ReloadLoopsPending > 0 {
		d.ReloadLoopsPending--
	}
	d.mutex.Unlock()
}

func (d *SnmpDevice) GetSelfThreadSafe() *SnmpDevice {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d
}

// GetBasicStats get basic info for this device
func (d *SnmpDevice) GetBasicStats() *DevStat {

	sum := 0
	for _, m := range d.Measurements {
		sum += len(m.OidSnmpMap)
	}
	d.mutex.Lock()
	stat := &DevStat{
		Requests:           d.Requests,
		Gets:               d.Gets,
		Errors:             d.Errors,
		ReloadLoopsPending: d.ReloadLoopsPending,
		DeviceActive:       d.DeviceActive,
		DeviceConnected:    d.DeviceConnected,
		NumMeasurements:    len(d.Measurements),
		NumMetrics:         sum,
	}
	d.mutex.Unlock()
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
		//we assume there is always a default db
		if d.Influx, ok = influxdb["default"]; !ok {
			//but
			return nil, fmt.Errorf("No influx config for snmp device: %s", d.cfg.ID)
		}
	}

	return d.Influx, nil
}

// ForceFltUpdate send info to update the filter counter to the next execution
func (d *SnmpDevice) ForceFltUpdate() {
	d.chFltUpdate <- true
}

// StopGather send signal to stop the Gathering process
func (d *SnmpDevice) StopGather() {
	d.chExit <- true
}

//RTActivate change activatio state in runtime
func (d *SnmpDevice) RTActivate(activate bool) {
	d.chEnabled <- activate
}

//RTActSnmpDebug change snmp debug runtime
func (d *SnmpDevice) RTActSnmpDebug(activate bool) {
	d.chDebug <- activate
}

// RTSetLogLevel set the log level for this device
func (d *SnmpDevice) RTSetLogLevel(level string) {
	d.chLogLevel <- level
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
				c, err := d.InitSnmpConnect(mVal.ID, d.cfg.SnmpDebug)
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
		//check for filters asociated with this measurement
		var mfilter *config.MeasFilterCfg
		for _, f := range d.cfg.MeasFilters {
			//we seach if exist in the filter Database
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
		//Get Data First Time ( usefull for counters)
		m.GetData()
	}
	//Initialize all snmpMetrics  objects and OID array
	//get data first time
	// useful to inicialize counter all value and test device snmp availability
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

	//Init channels
	d.chDebug = make(chan bool)
	d.chEnabled = make(chan bool)
	d.chLogLevel = make(chan string)
	d.chExit = make(chan bool)
	d.chFltUpdate = make(chan bool)
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

	if d.cfg.ConcurrentGather == true {
		d.Gather = d.measConcurrentGatherAndSend
	} else {
		d.Gather = d.measSeqGatherAndSend
	}

	return nil
}

// End The Oposite of Init() uninitialize all variables
func (d *SnmpDevice) End() {
	close(d.chDebug)
	close(d.chEnabled)
	close(d.chLogLevel)
	close(d.chExit)
	close(d.chFltUpdate)
	//release files
	//os.Close(d.log.Out)
	//release snmp resources
}

// SetSelfMonitoring set the ouput device where send monitoring metrics
func (d *SnmpDevice) SetSelfMonitoring(cfg *selfmon.SelfMon) {
	d.selfmon = cfg
}

// InitSnmpConnect does the  SNMP client conection and retrieve system info
func (d *SnmpDevice) InitSnmpConnect(mkey string, debug bool) (*gosnmp.GoSNMP, error) {
	d.Infof("Beginning SNMP connection")
	client, sysinfo, err := snmp.GetClient(d.cfg, d.log, mkey, debug)
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

func (d *SnmpDevice) addRequests(n int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Requests += n
}

func (d *SnmpDevice) resetCounters() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Gets = 0
	d.Errors = 0
}

func (d *SnmpDevice) addGets(n int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Gets += n
}

func (d *SnmpDevice) addErrors(n int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.Errors += n
}

// StartGather Main GoRutine method to begin snmp data collecting
func (d *SnmpDevice) StartGather(wg *sync.WaitGroup) {

	go d.startGatherGo(wg)
}

func (d *SnmpDevice) startGatherGo(wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	if d.DeviceActive && d.DeviceConnected {
		d.Infof("Begin first InidevInfo")
		startSnmp := time.Now()
		d.InitDevMeasurements()
		elapsedSnmp := time.Since(startSnmp)
		d.setFltUpdateStats(startSnmp, elapsedSnmp)
		d.Infof("snmp INIT runtime measurements/filters took [%s] ", elapsedSnmp)

	} else {
		d.Infof("Can not initialize this device: Is Active: %t  |  Conection Active: %t ", d.DeviceActive, d.snmpClientMap != nil)
	}

	d.Infof("Beginning gather process for device on host (%s)", d.cfg.Host)

	t := time.NewTicker(time.Duration(d.cfg.Freq) * time.Second)
	for {
		//if active
		if d.DeviceActive {
		FORCEINIT:
			//check if device is online
			if d.DeviceConnected == false {
				_, err := d.InitSnmpConnect("init", d.cfg.SnmpDebug)
				if err == nil {
					startSnmp := time.Now()
					d.InitDevMeasurements()
					elapsedSnmp := time.Since(startSnmp)
					d.setFltUpdateStats(startSnmp, elapsedSnmp)
					d.Infof("snmp INIT runtime measurements/filters took [%s] ", elapsedSnmp)
					// Round collection to nearest interval by sleeping
					utils.WaitAlignForNextCicle(d.cfg.Freq, d.log)
					//reprogram the ticker to aligned starts
					t.Stop()
					t = time.NewTicker(time.Duration(d.cfg.Freq) * time.Second)
					goto FORCEINIT //force one iteration now..after device has been connected  dont wait for next ticker (1 complete cicle)
				}
			} else {
				//device active and connected
				d.Infof("Init gather cicle mode Concurrent [ %t ]", d.cfg.ConcurrentGather)
				/*************************
				 *
				 * SNMP Gather data process
				 *
				 ***************************/
				d.resetCounters()
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
					d.setFltUpdateStats(startIdxUpdateStats, elapsedIdxUpdateStats)
					d.Infof("Index reload took [%s]", elapsedIdxUpdateStats)
				}
			}
		} else {
			d.Infof("Gather process is dissabled")
		}
	LOOP:
		for {
			select {
			case <-t.C:
				break LOOP
			case <-d.chExit:
				d.Infof("invoked EXIT from SNMP Gather process ")
				return
			case <-d.chFltUpdate:
				d.setReloadLoopsPending(1)
			case debug := <-d.chDebug:
				d.StateDebug = debug
				d.Infof(" invoked DEBUG  ACTIVE  [%t] ", debug)
				d.snmpClientMap = make(map[string]*gosnmp.GoSNMP)
				for _, m := range d.Measurements {
					c, err := d.InitSnmpConnect(m.ID, debug)
					if err != nil {
						d.Warnf("Error on recreate connection without debug for measurement %s", m.ID)
					}
					m.SetSnmpClient(c)
				}

			case status := <-d.chEnabled:
				d.DeviceActive = status
				d.Infof("device STATUS  ACTIVE  [%t] ", status)
			case level := <-d.chLogLevel:
				l, err := logrus.ParseLevel(level)
				if err != nil {
					d.Warnf("ERROR on Changing LOGLEVEL to [%t] ", level)
				}
				d.log.Level = l
				d.Infof("device loglevel Changed  [%s] ", level)
				d.CurLogLevel = d.log.Level.String()
			}
		}
	}
}
