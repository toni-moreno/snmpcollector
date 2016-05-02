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

type SnmpDeviceCfg struct {
	id string
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
	DeviceTagName  string   `toml:devicetagname`
	DeviceTagValue string   `toml:devicetagvalue`
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
	debugging chan bool
	enabled   chan chan bool
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
	c.id = name
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

	//Init Device Tags

	c.TagMap = make(map[string]string)
	if len(c.DeviceTagName) == 0 {
		c.DeviceTagName = "device"
	}

	var val string

	switch c.DeviceTagValue {
	case "id":
		val = c.id
	case "host":
		val = c.Host
	default:
		val = c.id
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
		fatal("Client connect error:", err)
	}
	c.snmpClient = client

	//Alloc array
	c.InfmeasArray = make([]*InfluxMeasurement, 0, 0)
	c.log.Debugf("-----------------Init device %s------------------", c.Host)
	//for this device get MetricGroups and search all measurements

	//log.Printf("SNMPDEV: %+v", c)
	for _, dev_meas := range c.MetricGroups {

		//Selecting all Metric Groups that matches with device.MetricGroups

		selGroups := make(map[string]*MGroupsCfg, 0)
		var RegExp = regexp.MustCompile(dev_meas)
		for key, val := range cfg.GetGroups {
			if RegExp.MatchString(key) {
				selGroups[key] = val
			}

		}

		c.log.Debugf("SNMP device %s has this SELECTED GROUPS: %+v", c.id, selGroups)

		//Only For selected Groups we will get all selected measurements and we will remove repeated values

		selMeas := make([]string, 0)
		for key, val := range selGroups {
			c.log.Debugln("Selecting from group", key)
			for _, item := range val.Measurements {
				c.log.Debugln("Selecting measurements", item, "from group", key)
				selMeas = append(selMeas, item)
			}
		}

		selMeasUniq := removeDuplicatesUnordered(selMeas)

		//Now we know what measurements names  will send influx from this device

		c.log.Debugln("DEVICE MEASUREMENT: ", dev_meas, "HOST: ", c.Host)
		for _, val := range selMeasUniq {
			//check if measurement exist
			if m_val, ok := cfg.Measurements[val]; !ok {
				c.log.Warnln("no measurement configured with name ", val, "in host :", c.Host)
			} else {
				c.log.Debugln("MEASUREMENT CFG KEY:", val, " VALUE ", m_val.Name)
				imeas := &InfluxMeasurement{
					cfg: m_val,
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
		for k_idx, v_idx := range m.values {
			c.log.Debugf("KEY iDX %s", k_idx)
			//index level
			for k_m, v_m := range v_idx {
				c.log.Debugf("KEY METRIC %s OID %s", k_m, v_m.realOID)
				m.snmpOids = append(m.snmpOids, v_m.realOID)
				m.oidSnmpMap[v_m.realOID] = v_m

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

func (c *SnmpDeviceCfg) printConfig() {

	fmt.Printf("Host: %s Port: %d Version: %s\n", c.Host, c.Port, c.SnmpVersion)
	fmt.Printf("----------------------------------------------\n")
	for _, v_m := range c.InfmeasArray {
		fmt.Printf(" Measurement : %s\n", v_m.cfg.id)
		fmt.Printf(" ----------------------------------------------------------\n")
		v_m.printConfig()
	}
}

func (c *SnmpDeviceCfg) DebugAction() string {
	debug := make(chan bool)
	c.enabled <- debug
	if <-debug {
		return "disable"
	}
	return "enable"
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

func (c *SnmpDeviceCfg) DebugLog() *olog.Logger {
	name := filepath.Join(logDir, "snmpdebug_"+strings.Replace(c.id, ".", "-", -1)+".log")
	if l, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644); err == nil {
		return olog.New(l, "", 0)
	} else {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
}

func (c *SnmpDeviceCfg) Gather(wg *sync.WaitGroup) {
	client := c.snmpClient
	debug := false

	c.log.Infof("Beginning gather process for device %s (%s)", c.id, c.Host)

	client = c.snmpClient
	s := time.Tick(time.Duration(c.Freq) * time.Second)
	for {
		time.Sleep(time.Duration(c.Timeout) * time.Second)
		bpts := c.Influx.BP()

		for _, m := range c.InfmeasArray {
			c.log.Debugf("----------------Processing measurement : %s", m.cfg.id)
			if m.cfg.GetMode == "value" || c.SnmpVersion == "1" {
				n_gets, n_errors, _ := m.SnmpGetData(client)
				if n_gets > 0 {
					c.addGets(n_gets)
				}
				if n_errors > 0 {
					c.addErrors(n_errors)
				}
			} else {
				n_gets, n_errors, _ := m.SnmpBulkData(client)
				if n_gets > 0 {
					c.addGets(n_gets)
				}
				if n_errors > 0 {
					c.addErrors(n_errors)
				}

			}
			//prepare batchpoint and
			points := m.GetInfluxPoint( /*c.Host,*/ c.TagMap)
			(*bpts).AddPoints(points)

		}
		c.Influx.Send(bpts)

		// pause for interval period and have optional debug toggling
	LOOP:
		for {
			select {
			case <-s:
				break LOOP
			case debug := <-c.debugging:
				c.log.Infof("debugging: %s", debug)
				if debug && client.Logger == nil {
					client.Logger = c.DebugLog()
				} else {
					client.Logger = nil
				}
			case status := <-c.enabled:
				status <- debug
			}
		}
	}
	wg.Done()
}
