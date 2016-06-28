package main

import (
	"fmt"
	//"io/ioutil"
	olog "log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
)

// SysInfo basic information for any SNMP device
type SysInfo struct {
	sysDescr    string
	sysUptime   time.Duration
	sysContact  string
	sysName     string
	sysLocation string
}

// SnmpDeviceCfg contains all snmp related device definitions
type SnmpDeviceCfg struct {
	ID string
	//snmp connection config
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
	Retries int    `toml:"retries"`
	Timeout int    `toml:"timeout"`
	Repeat  int    `toml:"repeat"`
	//snmp auth  config
	SnmpVersion string `toml:"snmpversion"`
	Community   string `toml:"community"`
	V3SecLevel  string `toml:"v3seclevel"`
	V3AuthUser  string `toml:"v3authuser"`
	V3AuthPass  string `toml:"v3authpass"`
	V3AuthProt  string `toml:"v3authprot"`
	V3PrivPass  string `toml:"v3privpass"`
	V3PrivProt  string `toml:"v3privprot"`
	//snmp runtime config
	Freq int `toml:"freq"`
	//PortFile string `toml:"portfile"`
	Config    string `toml:"config"`
	LogLevel  string `toml:"loglevel"`
	LogFile   string `toml:"logfile"`
	log       *logrus.Logger
	SnmpDebug bool `toml:"snmpdebug"`
	//influx tags
	DeviceTagName  string   `toml:"devicetagname"`
	DeviceTagValue string   `toml:"devicetagvalue"`
	ExtraTags      []string `toml:"extra-tags"`
	TagMap         map[string]string
	//Filters for measurements
	MetricGroups []string   `toml:"metricgroups"`
	MeasFilters  [][]string `toml:"measfilters"`

	//Measurments array

	InfmeasArray []*InfluxMeasurement

	//SNMP and Influx Clients config
	snmpClient *gosnmp.GoSNMP
	Influx     *InfluxConfig
	LastError  time.Time
	//Runtime stats
	Requests int64
	Gets     int64
	Errors   int64
	//runtime controls
	/*debugging chan bool
	enabled   chan chan bool*/

	chDebug      chan bool
	chEnabled    chan bool
	deviceActive bool
	stateDebug   bool
	//basic sistem info
	sysInfo SysInfo
}

// GetSysInfo got system basic info from a snmp client
func (c *SnmpDeviceCfg) GetSysInfo(client *gosnmp.GoSNMP) (SysInfo, error) {
	//Get Basic System Info
	// sysDescr     .1.3.6.1.2.1.1.1.0
	// sysUpTime    .1.3.6.1.2.1.1.3.0
	// sysContact   .1.3.6.1.2.1.1.4.0
	// sysName      .1.3.6.1.2.1.1.5.0
	// sysLocation  .1.3.6.1.2.1.1.6.0
	sysOids := []string{
		".1.3.6.1.2.1.1.1.0",
		".1.3.6.1.2.1.1.3.0",
		".1.3.6.1.2.1.1.4.0",
		".1.3.6.1.2.1.1.5.0",
		".1.3.6.1.2.1.1.6.0"}

	info := SysInfo{sysDescr: "", sysUptime: time.Duration(0), sysContact: "", sysName: "", sysLocation: ""}
	pkt, err := client.Get(sysOids)

	if err != nil {
		c.log.Errorf("Error on getting initial basic system, Info to device %s: %s", c.Host, err)
		return info, err
	}

	for idx, pdu := range pkt.Variables {
		c.log.Debugf("DEBUG pdu:%+v", pdu)
		if pdu.Value == nil {
			continue
		}
		switch idx {
		case 0: // sysDescr     .1.3.6.1.2.1.1.1.0
			if pdu.Type == gosnmp.OctetString {
				info.sysDescr = string(pdu.Value.([]byte))
			} else {
				c.log.Warnf("Error on getting system %s sysDescr return data of type %v", c.Host, pdu.Type)
			}
		case 1: // sysUpTime    .1.3.6.1.2.1.1.3.0
			if pdu.Type == gosnmp.TimeTicks {
				seconds := uint32(pdu.Value.(int)) / 100
				info.sysUptime = time.Duration(seconds) * time.Second
			} else {
				c.log.Warnf("Error on getting system %s sysDescr return data of type %v", c.Host, pdu.Type)
			}
		case 2: // sysContact   .1.3.6.1.2.1.1.4.0
			if pdu.Type == gosnmp.OctetString {
				info.sysContact = string(pdu.Value.([]byte))
			} else {
				c.log.Warnf("Error on getting system %s sysContact return data of type %v", c.Host, pdu.Type)
			}
		case 3: // sysName      .1.3.6.1.2.1.1.5.0
			if pdu.Type == gosnmp.OctetString {
				info.sysName = string(pdu.Value.([]byte))
			} else {
				c.log.Warnf("Error on getting system %s sysName return data of type %v", c.Host, pdu.Type)
			}
		case 4: // sysLocation  .1.3.6.1.2.1.1.6.0
			if pdu.Type == gosnmp.OctetString {
				info.sysDescr = string(pdu.Value.([]byte))
			} else {
				c.log.Warnf("Error on getting system %s sysLocation return data of type %v", c.Host, pdu.Type)
			}
		}
	}
	return info, nil
}

//InitDevSnmpInfo generte all internal structs from SNMP device
func (c *SnmpDeviceCfg) InitDevSnmpInfo() {

	//Alloc array
	c.InfmeasArray = make([]*InfluxMeasurement, 0, 0)
	c.log.Debugf("-----------------Init device %s------------------", c.Host)
	//for this device get MetricGroups and search all measurements

	//log.Printf("SNMPDEV: %+v", c)
	for _, devMeas := range c.MetricGroups {

		//Selecting all Metric Groups that matches with device.MetricGroups

		selGroups := make(map[string]*MGroupsCfg, 0)
		var RegExp = regexp.MustCompile(devMeas)
		for key, val := range cfg.GetGroups {
			if RegExp.MatchString(key) {
				selGroups[key] = val
			}

		}

		c.log.Debugf("SNMP device %s has this SELECTED GROUPS: %+v", c.ID, selGroups)

		//Only For selected Groups we will get all selected measurements and we will remove repeated values

		var selMeas []string
		for key, val := range selGroups {
			c.log.Debugln("Selecting from group", key)
			for _, item := range val.Measurements {
				c.log.Debugln("Selecting measurements", item, "from group", key)
				selMeas = append(selMeas, item)
			}
		}

		selMeasUniq := removeDuplicatesUnordered(selMeas)

		//Now we know what measurements names  will send influx from this device

		c.log.Debugln("DEVICE MEASUREMENT: ", devMeas, "HOST: ", c.Host)
		for _, val := range selMeasUniq {
			//check if measurement exist
			if mVal, ok := cfg.Measurements[val]; !ok {
				c.log.Warnln("no measurement configured with name ", val, "in host :", c.Host)
			} else {
				c.log.Debugln("MEASUREMENT CFG KEY:", val, " VALUE ", mVal.Name)
				imeas := &InfluxMeasurement{
					cfg: mVal,
					log: c.log,
				}
				c.InfmeasArray = append(c.InfmeasArray, imeas)
			}
		}
	}

	/*For each Indexed measurement
	a) LoadLabels for all device available tags
	b) apply filters , and get list of names Indexed tames for add to IndexTAG

	Filter format
	-------------
	F[0] 	= measurement name
	F[1] 	= FilterType ( know values "file" , "OIDCondition" )
	F[2] 	= Filename ( if F[1] = file)
		= OIDname for condition ( if[F1] = OIDCondition )
	F[3]  = Condition Type ( if[F1] = OIDCondition ) (known values "eq","lt","gt")
	F[4]  = Value for condition
	*/

	for _, m := range c.InfmeasArray {
		//loading all posible values.
		if m.cfg.GetMode == "indexed" {
			c.log.Infof("Loading Indexed values in : %s", m.cfg.id)
			m.loadIndexedLabels(c)
		}
		//loading filters
		c.log.Debugf("Looking for filters set to: %s ", m.cfg.id)
		var flt string
		for _, f := range c.MeasFilters {
			if f[0] == m.cfg.id {
				c.log.Debugf("filter Found  %s  (type %s)", m.cfg.id, f[1])
				if m.cfg.GetMode == "indexed" {
					flt = f[1]
					//OK we can apply filters
					switch {
					case flt == "file":
						enable := f[3] == "EnableAlias"
						m.Filter = &MeasFilterCfg{
							fType:       flt,
							FileName:    f[2],
							enableAlias: enable,
						}
						m.applyFileFilter(m.Filter.FileName, enable)
					case flt == "OIDCondition":
						m.Filter = &MeasFilterCfg{
							fType:     flt,
							OIDCond:   f[2],
							condType:  f[3],
							condValue: f[4],
						}

						m.applyOIDCondFilter(c,
							m.Filter.OIDCond,
							m.Filter.condType,
							m.Filter.condValue)
					default:
						c.log.Errorf("Invalid  GetMode Type %s for measurement: %s", flt, m.cfg.id)
					}

				} else {
					//no filters enabled  on not indexed measurements
					c.log.Debugf("Filters %s not match with indexed measurements: %s", f[0], m.cfg.id)
				}

			} else {
				c.log.Infof("Filter not found for measurement:i %s", m.cfg.id)
			}
		}
		//Loading final Values to query with snmp
		if len(c.MeasFilters) > 0 {
			m.filterIndexedLabels(flt)
		} else {
			m.IndexedLabels()
		}
		c.log.Debugf("MEASUREMENT HOST:%s | %s | %+v\n", c.Host, m.cfg.id, m)
	}
	//Initialize all snmpMetrics  objects and OID array
	for _, m := range c.InfmeasArray {
		c.log.Debug("Initialize OID array")
		m.values = make(map[string]map[string]*SnmpMetric)

		//create metrics.
		switch m.cfg.GetMode {
		case "value":
			//for each field
			idx := make(map[string]*SnmpMetric)
			for _, smcfg := range m.cfg.fieldMetric {
				c.log.Debugf("initializing [value]metric cfgi %s", smcfg.id)
				metric := &SnmpMetric{cfg: smcfg, realOID: smcfg.BaseOID}
				metric.Init()
				idx[smcfg.id] = metric
			}
			m.values["0"] = idx

		case "indexed":
			//for each field an each index (previously initialized)
			for key, label := range m.CurIndexedLabels {
				idx := make(map[string]*SnmpMetric)
				c.log.Debugf("initializing [indexed] metric cfg for [%s/%s]", key, label)
				for _, smcfg := range m.cfg.fieldMetric {
					metric := &SnmpMetric{cfg: smcfg, realOID: smcfg.BaseOID + "." + key}
					metric.Init()
					idx[smcfg.id] = metric
				}
				m.values[label] = idx
			}

		default:
			c.log.Errorf("Unknown Measurement GetMode Config :%s", m.cfg.GetMode)
		}
		c.log.Debugf("ARRAY VALUES for host %s :%s : %+v", c.Host, m.cfg.Name, m.values)
		//building real OID array for SNMPWALK and OID=> snmpMetric map to asign results to each object
		m.snmpOids = []string{}
		m.oidSnmpMap = make(map[string]*SnmpMetric)
		//metric level
		for kIdx, vIdx := range m.values {
			c.log.Debugf("KEY iDX %s", kIdx)
			//index level
			for kM, vM := range vIdx {
				c.log.Debugf("KEY METRIC %s OID %s", kM, vM.realOID)
				m.snmpOids = append(m.snmpOids, vM.realOID)
				m.oidSnmpMap[vM.realOID] = vM

			}
		}
		//		log.Printf("DEBUG oid map %+v", m.oidSnmpMap)

	}
	//get data first time
	// useful to inicialize counter all value and test device snmp availability
	for _, m := range c.InfmeasArray {
		if m.cfg.GetMode == "value" || c.SnmpVersion == "1" {
			_, _, err := m.SnmpGetData(c.snmpClient)
			if err != nil {
				c.log.Fatalf("SNMP First Get Data error for host: %s", c.Host)
			}
		} else {
			_, _, err := m.SnmpBulkData(c.snmpClient)
			if err != nil {
				c.log.Fatalf("SNMP First Get Data error for host: %s", c.Host)
			}

		}

	}

}

/*
Init  does the following

- allocate and fill array for all measurements defined for this device
- for each indexed measurement  load device labels from IndexedOID and fiter them if defined measurement filters.
- Initialice each SnmpMetric from each measuremet.
*/
func (c *SnmpDeviceCfg) Init(name string) {
	log.Infof("Initializing device %s\n", name)
	//Init id
	c.ID = name
	//Init Logger

	if len(c.LogFile) == 0 {
		c.LogFile = cfg.General.LogDir + "/" + name + ".log"

	}
	if len(c.LogLevel) == 0 {
		c.LogLevel = "info"
	}

	f, err := os.OpenFile(c.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	c.log = logrus.New()
	c.log.Out = f
	l, _ := logrus.ParseLevel(c.LogLevel)
	c.log.Level = l

	//Init channels
	c.chDebug = make(chan bool)
	c.chEnabled = make(chan bool)
	c.deviceActive = true

	//Init Device Tags

	c.TagMap = make(map[string]string)
	if len(c.DeviceTagName) == 0 {
		c.DeviceTagName = "device"
	}

	var val string

	switch c.DeviceTagValue {
	case "id":
		val = c.ID
	case "host":
		val = c.Host
	default:
		val = c.ID
		c.log.Warnf("Unkwnown DeviceTagValue %s set ID (%s) as value", c.DeviceTagValue, val)
	}

	c.TagMap[c.DeviceTagName] = val

	if len(c.ExtraTags) > 0 {
		for _, tag := range c.ExtraTags {
			s := strings.Split(tag, "=")
			key, value := s[0], s[1]
			c.TagMap[key] = value
		}
	} else {
		c.log.Warnf("No map detected in device %s\n", name)
	}

	//Init SNMP client device

	client, err := snmpClient(c)
	if err != nil {
		c.log.Errorf("Client connect error to device: %s  error :%s", c.ID, err)
		c.snmpClient = nil
		return
	}
	c.snmpClient = client
	c.InitDevSnmpInfo()
}

func (c *SnmpDeviceCfg) printConfig() {

	fmt.Printf("Host: %s Port: %d Version: %s\n", c.Host, c.Port, c.SnmpVersion)
	fmt.Printf("----------------------------------------------\n")
	for _, vM := range c.InfmeasArray {
		fmt.Printf(" Measurement : %s\n", vM.cfg.id)
		fmt.Printf(" ----------------------------------------------------------\n")
		vM.printConfig()
	}
}

func (c *SnmpDeviceCfg) incRequests() {
	atomic.AddInt64(&c.Requests, 1)
}

func (c *SnmpDeviceCfg) addRequests(n int64) {
	atomic.AddInt64(&c.Requests, n)
}
func (c *SnmpDeviceCfg) incGets() {
	atomic.AddInt64(&c.Gets, 1)
}
func (c *SnmpDeviceCfg) addGets(n int64) {
	atomic.AddInt64(&c.Gets, 1)
}

func (c *SnmpDeviceCfg) incErrors() {
	atomic.AddInt64(&c.Errors, 1)
}

func (c *SnmpDeviceCfg) addErrors(n int64) {
	atomic.AddInt64(&c.Errors, n)
}

// DebugLog returns a logger handler for snmp debug data
func (c *SnmpDeviceCfg) DebugLog() *olog.Logger {
	name := filepath.Join(logDir, "snmpdebug_"+strings.Replace(c.ID, ".", "-", -1)+".log")
	if l, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644); err == nil {
		return olog.New(l, "", 0)
	} else {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
}

//Gather Main GoRutine method to begin snmp data collecting
func (c *SnmpDeviceCfg) Gather(wg *sync.WaitGroup) {
	//client := c.snmpClient
	//debug := false

	c.log.Infof("Beginning gather process for device %s (%s)", c.ID, c.Host)

	s := time.Tick(time.Duration(c.Freq) * time.Second)
	for {
		//if active
		if c.deviceActive {
			//check if device is online
			if c.snmpClient == nil {
				client, err := snmpClient(c)
				if err != nil {
					c.log.Errorf("Client connect error to device: %s  error :%s", c.ID, err)
				} else {
					c.snmpClient = client
					c.log.Infof("SNMP connection stablishedm initializing SnmpDevice")
					c.InitDevSnmpInfo()
					//device not initialized
				}
			} else {
				//TODO: review if necesary this Sleep and what is the exact goal for the Timeout
				//time.Sleep(time.Duration(c.Timeout) * time.Second)

				bpts := c.Influx.BP()
				startSnmp := time.Now()
				for _, m := range c.InfmeasArray {
					c.log.Debugf("----------------Processing measurement : %s", m.cfg.id)
					var nGets, nErrors int64
					if m.cfg.GetMode == "value" || c.SnmpVersion == "1" {
						nGets, nErrors, _ = m.SnmpGetData(c.snmpClient)
					} else {
						nGets, nErrors, _ = m.SnmpBulkData(c.snmpClient)
					}
					if nGets > 0 {
						c.addGets(nGets)
					}
					if nErrors > 0 {
						c.addErrors(nErrors)
					}
					//prepare batchpoint and
					points := m.GetInfluxPoint( /*c.Host,*/ c.TagMap)
					(*bpts).AddPoints(points)

				}
				elapsedSnmp := time.Since(startSnmp)
				c.log.Infof("snmpdevice [%s] snmp poolling took [%s] ", c.ID, elapsedSnmp)
				startInflux := time.Now()
				c.Influx.Send(bpts)
				elapsedInflux := time.Since(startInflux)
				c.log.Infof("snmpdevice [%s] influx send took [%s]", c.ID, elapsedInflux)
				// pause for interval period and have optional debug toggling
			}
		}
	LOOP:
		for {
			select {
			case <-s:
				break LOOP
			case debug := <-c.chDebug:
				c.stateDebug = debug
				c.log.Infof("DEBUG  ACTIVE %s [%t] ", c.ID, debug)
				if debug && c.snmpClient.Logger == nil {
					c.snmpClient.Logger = c.DebugLog()
				} else {
					c.snmpClient.Logger = nil
				}
			case status := <-c.chEnabled:
				c.deviceActive = status
				c.log.Printf("STATUS  ACTIVE %s [%t] ", c.ID, status)
			}
		}
	}
	wg.Done()
}
