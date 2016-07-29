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

// SnmpDevice contains all runtime device related device configu ns and state
type SnmpDevice struct {
	cfg *SnmpDeviceCfg
	log *logrus.Logger
	//runtime built TagMap
	TagMap map[string]string
	//Measurments array

	InfmeasArray []*InfluxMeasurement

	//SNMP and Influx Clients config
	snmpClient *gosnmp.GoSNMP
	Influx     *InfluxDB
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
func (d *SnmpDevice) GetSysInfo(client *gosnmp.GoSNMP) (SysInfo, error) {
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
		d.log.Errorf("Error on getting initial basic system, Info to device %s: %s", d.cfg.Host, err)
		return info, err
	}

	for idx, pdu := range pkt.Variables {
		d.log.Debugf("DEBUG pdu:%+v", pdu)
		if pdu.Value == nil {
			continue
		}
		switch idx {
		case 0: // sysDescr     .1.3.6.1.2.1.1.1.0
			if pdu.Type == gosnmp.OctetString {
				info.sysDescr = string(pdu.Value.([]byte))
			} else {
				d.log.Warnf("Error on getting system %s sysDescr return data of type %v", d.cfg.Host, pdu.Type)
			}
		case 1: // sysUpTime    .1.3.6.1.2.1.1.3.0
			if pdu.Type == gosnmp.TimeTicks {
				seconds := uint32(pdu.Value.(int)) / 100
				info.sysUptime = time.Duration(seconds) * time.Second
			} else {
				d.log.Warnf("Error on getting system %s sysDescr return data of type %v", d.cfg.Host, pdu.Type)
			}
		case 2: // sysContact   .1.3.6.1.2.1.1.4.0
			if pdu.Type == gosnmp.OctetString {
				info.sysContact = string(pdu.Value.([]byte))
			} else {
				d.log.Warnf("Error on getting system %s sysContact return data of type %v", d.cfg.Host, pdu.Type)
			}
		case 3: // sysName      .1.3.6.1.2.1.1.5.0
			if pdu.Type == gosnmp.OctetString {
				info.sysName = string(pdu.Value.([]byte))
			} else {
				d.log.Warnf("Error on getting system %s sysName return data of type %v", d.cfg.Host, pdu.Type)
			}
		case 4: // sysLocation  .1.3.6.1.2.1.1.6.0
			if pdu.Type == gosnmp.OctetString {
				info.sysDescr = string(pdu.Value.([]byte))
			} else {
				d.log.Warnf("Error on getting system %s sysLocation return data of type %v", d.cfg.Host, pdu.Type)
			}
		}
	}
	return info, nil
}

//InitDevSnmpInfo generte all internal structs from SNMP device
func (d *SnmpDevice) InitDevSnmpInfo() {

	//Alloc array
	d.InfmeasArray = make([]*InfluxMeasurement, 0, 0)
	d.log.Debugf("-----------------Init device %s------------------", d.cfg.Host)
	//for this device get MetricGroups and search all measurements

	//log.Printf("SNMPDEV: %+v", c)
	for _, devMeas := range d.cfg.MetricGroups {

		//Selecting all Metric Groups that matches with device.MetricGroups

		selGroups := make(map[string]*MGroupsCfg, 0)
		var RegExp = regexp.MustCompile(devMeas)
		for key, val := range cfg.GetGroups {
			if RegExp.MatchString(key) {
				selGroups[key] = val
			}

		}

		d.log.Debugf("SNMP device %s has this SELECTED GROUPS: %+v", d.cfg.ID, selGroups)

		//Only For selected Groups we will get all selected measurements and we will remove repeated values

		var selMeas []string
		for key, val := range selGroups {
			d.log.Debugln("Selecting from group", key)
			for _, item := range val.Measurements {
				d.log.Debugln("Selecting measurements", item, "from group", key)
				selMeas = append(selMeas, item)
			}
		}

		selMeasUniq := removeDuplicatesUnordered(selMeas)

		//Now we know what measurements names  will send influx from this device

		d.log.Debugln("DEVICE MEASUREMENT: ", devMeas, "HOST: ", d.cfg.Host)
		for _, val := range selMeasUniq {
			//check if measurement exist
			if mVal, ok := cfg.Measurements[val]; !ok {
				d.log.Warnln("no measurement configured with name ", val, "in host :", d.cfg.Host)
			} else {
				d.log.Debugln("MEASUREMENT CFG KEY:", val, " VALUE ", mVal.Name)
				imeas := &InfluxMeasurement{
					cfg: mVal,
					log: d.log,
				}
				d.InfmeasArray = append(d.InfmeasArray, imeas)
			}
		}
	}

	/*For each Indexed measurement
	a) LoadLabels for all device available tags
	b) apply filters , and get list of names Indexed tames for add to IndexTAG
	*/

	for _, m := range d.InfmeasArray {
		//loading all posible values.
		if m.cfg.GetMode == "indexed" {
			d.log.Infof("Loading Indexed values in : %s", m.cfg.ID)
			m.loadIndexedLabels(d)
		}
		//loading filters
		d.log.Debugf("Looking for filters set to: %s ", m.cfg.ID)
		var ftype string
		for _, f := range d.cfg.MeasFilters {
			if filter, ok := cfg.MFilters[f]; ok {
				if filter.IDMeasurementCfg == m.cfg.ID {
					m.Filter = filter
					d.log.Debugf("filter Found  %s for measurement %s  (type %s)", f, m.cfg.ID, m.Filter.FType)
					if m.cfg.GetMode == "indexed" {
						ftype = m.Filter.FType
						switch m.Filter.FType {
						case "file":
							m.applyFileFilter(m.Filter.FileName,
								m.Filter.EnableAlias)
						case "OIDCondition":
							m.applyOIDCondFilter(d,
								m.Filter.OIDCond,
								m.Filter.CondType,
								m.Filter.CondValue)
						default:
							d.log.Errorf("Invalid Filter Type %s for measurement: %s", m.Filter.FType, m.cfg.ID)
						}
					} else {
						//no filters enabled  on not indexed measurements
						d.log.Debugf("Filters %s not match with indexed measurements: %s", f, m.cfg.ID)
					}
				} //if filter.IDMeasurementCfg == m.cfg.ID
			} else {
				d.log.Debugf("Filters %s  found in device but not in configured filters", f)
			} //ok
		} //for
		//Loading final Values to query with snmp

		if len(d.cfg.MeasFilters) > 0 {
			//FIXME: this final indexation is done only with the last type !!! even if more than one filter exist with different type
			m.filterIndexedLabels(ftype)
		} else {
			log.Debugf("no filters found for device %s", d.cfg.ID)
			m.IndexedLabels()
		}
		d.log.Debugf("MEASUREMENT HOST:%s | %s | %+v\n", d.cfg.Host, m.cfg.ID, m)
	}
	//Initialize all snmpMetrics  objects and OID array
	for _, m := range d.InfmeasArray {
		d.log.Debug("Initialize OID array")
		m.values = make(map[string]map[string]*SnmpMetric)

		//create metrics.
		switch m.cfg.GetMode {
		case "value":
			//for each field
			idx := make(map[string]*SnmpMetric)
			for _, smcfg := range m.cfg.fieldMetric {
				d.log.Debugf("initializing [value]metric cfgi %s", smcfg.ID)
				metric := &SnmpMetric{cfg: smcfg, realOID: smcfg.BaseOID}
				metric.Init()
				idx[smcfg.ID] = metric
			}
			m.values["0"] = idx

		case "indexed":
			//for each field an each index (previously initialized)
			for key, label := range m.CurIndexedLabels {
				idx := make(map[string]*SnmpMetric)
				d.log.Debugf("initializing [indexed] metric cfg for [%s/%s]", key, label)
				for _, smcfg := range m.cfg.fieldMetric {
					metric := &SnmpMetric{cfg: smcfg, realOID: smcfg.BaseOID + "." + key}
					metric.Init()
					idx[smcfg.ID] = metric
				}
				m.values[label] = idx
			}

		default:
			d.log.Errorf("Unknown Measurement GetMode Config :%s", m.cfg.GetMode)
		}
		d.log.Debugf("ARRAY VALUES for host %s :%s : %+v", d.cfg.Host, m.cfg.Name, m.values)
		//building real OID array for SNMPWALK and OID=> snmpMetric map to asign results to each object
		m.snmpOids = []string{}
		m.oidSnmpMap = make(map[string]*SnmpMetric)
		//metric level
		for kIdx, vIdx := range m.values {
			d.log.Debugf("KEY iDX %s", kIdx)
			//index level
			for kM, vM := range vIdx {
				d.log.Debugf("KEY METRIC %s OID %s", kM, vM.realOID)
				m.snmpOids = append(m.snmpOids, vM.realOID)
				m.oidSnmpMap[vM.realOID] = vM

			}
		}
		//		log.Printf("DEBUG oid map %+v", m.oidSnmpMap)

	}
	//get data first time
	// useful to inicialize counter all value and test device snmp availability
	for _, m := range d.InfmeasArray {
		if m.cfg.GetMode == "value" || d.cfg.SnmpVersion == "1" {
			_, _, err := m.SnmpGetData(d.snmpClient)
			if err != nil {
				d.log.Fatalf("SNMP First Get Data error for host: %s", d.cfg.Host)
			}
		} else {
			_, _, err := m.SnmpBulkData(d.snmpClient)
			if err != nil {
				d.log.Fatalf("SNMP First Get Data error for host: %s", d.cfg.Host)
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
func (d *SnmpDevice) Init(name string) {
	log.Infof("Initializing device %s\n", name)
	//Init id
	d.cfg.ID = name

	//Init Logger

	if len(d.cfg.LogFile) == 0 {
		d.cfg.LogFile = cfg.General.LogDir + "/" + name + ".log"

	}
	if len(d.cfg.LogLevel) == 0 {
		d.cfg.LogLevel = "info"
	}

	f, err := os.OpenFile(d.cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	d.log = logrus.New()
	d.log.Out = f
	l, _ := logrus.ParseLevel(d.cfg.LogLevel)
	d.log.Level = l

	//Init channels
	d.chDebug = make(chan bool)
	d.chEnabled = make(chan bool)
	d.deviceActive = true

	//Init Device Tags

	d.TagMap = make(map[string]string)
	if len(d.cfg.DeviceTagName) == 0 {
		d.cfg.DeviceTagName = "device"
	}

	var val string

	switch d.cfg.DeviceTagValue {
	case "id":
		val = d.cfg.ID
	case "host":
		val = d.cfg.Host
	default:
		val = d.cfg.ID
		d.log.Warnf("Unkwnown DeviceTagValue %s set ID (%s) as value", d.cfg.DeviceTagValue, val)
	}

	d.TagMap[d.cfg.DeviceTagName] = val

	if len(d.cfg.ExtraTags) > 0 {
		for _, tag := range d.cfg.ExtraTags {
			s := strings.Split(tag, "=")
			key, value := s[0], s[1]
			d.TagMap[key] = value
		}
	} else {
		d.log.Warnf("No map detected in device %s\n", name)
	}

	//Init SNMP client device

	client, err := snmpClient(d)
	if err != nil {
		d.log.Errorf("Client connect error to device: %s  error :%s", d.cfg.ID, err)
		d.snmpClient = nil
		return
	}
	d.snmpClient = client
	d.InitDevSnmpInfo()
}

func (d *SnmpDevice) printConfig() {

	fmt.Printf("Host: %s Port: %d Version: %s\n", d.cfg.Host, d.cfg.Port, d.cfg.SnmpVersion)
	fmt.Printf("----------------------------------------------\n")
	for _, vM := range d.InfmeasArray {
		fmt.Printf(" Measurement : %s\n", vM.cfg.ID)
		fmt.Printf(" ----------------------------------------------------------\n")
		vM.printConfig()
	}
}

func (d *SnmpDevice) incRequests() {
	atomic.AddInt64(&d.Requests, 1)
}

func (d *SnmpDevice) addRequests(n int64) {
	atomic.AddInt64(&d.Requests, n)
}
func (d *SnmpDevice) incGets() {
	atomic.AddInt64(&d.Gets, 1)
}
func (d *SnmpDevice) addGets(n int64) {
	atomic.AddInt64(&d.Gets, 1)
}

func (d *SnmpDevice) incErrors() {
	atomic.AddInt64(&d.Errors, 1)
}

func (d *SnmpDevice) addErrors(n int64) {
	atomic.AddInt64(&d.Errors, n)
}

// DebugLog returns a logger handler for snmp debug data
func (d *SnmpDevice) DebugLog() *olog.Logger {
	name := filepath.Join(logDir, "snmpdebug_"+strings.Replace(d.cfg.ID, ".", "-", -1)+".log")
	if l, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644); err == nil {
		return olog.New(l, "", 0)
	} else {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
}

//Gather Main GoRutine method to begin snmp data collecting
func (d *SnmpDevice) Gather(wg *sync.WaitGroup) {
	//client := d.snmpClient
	//debug := false

	d.log.Infof("Beginning gather process for device %s (%s)", d.cfg.ID, d.cfg.Host)

	s := time.Tick(time.Duration(d.cfg.Freq) * time.Second)
	for {
		//if active
		if d.deviceActive {
			//check if device is online
			if d.snmpClient == nil {
				client, err := snmpClient(d)
				if err != nil {
					d.log.Errorf("Client connect error to device: %s  error :%s", d.cfg.ID, err)
				} else {
					d.snmpClient = client
					d.log.Infof("SNMP connection stablishedm initializing SnmpDevice")
					d.InitDevSnmpInfo()
					//device not initialized
				}
			} else {
				//TODO: review if necesary this Sleep and what is the exact goal for the Timeout
				//time.Sleep(time.Duration(d.cfg.Timeout) * time.Second)

				bpts := d.Influx.BP()
				startSnmp := time.Now()
				for _, m := range d.InfmeasArray {
					d.log.Debugf("----------------Processing measurement : %s", m.cfg.ID)
					var nGets, nErrors int64
					if m.cfg.GetMode == "value" || d.cfg.SnmpVersion == "1" {
						nGets, nErrors, _ = m.SnmpGetData(d.snmpClient)
					} else {
						nGets, nErrors, _ = m.SnmpBulkData(d.snmpClient)
					}
					if nGets > 0 {
						d.addGets(nGets)
					}
					if nErrors > 0 {
						d.addErrors(nErrors)
					}
					//prepare batchpoint and
					points := m.GetInfluxPoint( /*d.cfg.Host,*/ d.TagMap)
					(*bpts).AddPoints(points)

				}
				elapsedSnmp := time.Since(startSnmp)
				d.log.Infof("snmpdevice [%s] snmp poolling took [%s] ", d.cfg.ID, elapsedSnmp)
				startInflux := time.Now()
				d.Influx.Send(bpts)
				elapsedInflux := time.Since(startInflux)
				d.log.Infof("snmpdevice [%s] influx send took [%s]", d.cfg.ID, elapsedInflux)
				// pause for interval period and have optional debug toggling
			}
		}
	LOOP:
		for {
			select {
			case <-s:
				break LOOP
			case debug := <-d.chDebug:
				d.stateDebug = debug
				d.log.Infof("DEBUG  ACTIVE %s [%t] ", d.cfg.ID, debug)
				if debug && d.snmpClient.Logger == nil {
					d.snmpClient.Logger = d.DebugLog()
				} else {
					d.snmpClient.Logger = nil
				}
			case status := <-d.chEnabled:
				d.deviceActive = status
				d.log.Printf("STATUS  ACTIVE %s [%t] ", d.cfg.ID, status)
			}
		}
	}
	wg.Done()
}
