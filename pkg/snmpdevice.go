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
	//basic sistem info
	sysInfo SysInfo
	//runtime built TagMap
	TagMap map[string]string
	//Measurements array

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
	deviceActive bool
	stateDebug   bool

	chDebug   chan bool
	chEnabled chan bool
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

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key, _ := range encountered {
		result = append(result, key)
	}
	return result
}

//InitDevSnmpInfo generte all internal structs from SNMP device
func (d *SnmpDevice) InitDevSnmpInfo() {

	//Alloc array
	d.InfmeasArray = make([]*InfluxMeasurement, 0, 0)
	d.log.Debugf("-----------------Init device %s------------------", d.cfg.Host)
	//for this device get MeasurementGroups and search all measurements

	for _, devMeas := range d.cfg.MeasurementGroups {

		//Selecting all Metric Groups that matches with device.MeasurementGroups
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

				//creating a new measurement runtime object and asigning to array
				d.InfmeasArray = append(d.InfmeasArray, &InfluxMeasurement{ID: mVal.ID, cfg: mVal, log: d.log, snmpClient: d.snmpClient})
			}
		}
	}

	/*For each  measurement look for filters and Initialize Measurement with this Filter 	*/

	for _, m := range d.InfmeasArray {
		//check for filters asociated with this measurement
		var mfilter *MeasFilterCfg
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
			log.Debugf("filters %s found for device %s and measurment %s ", mfilter.ID, d.cfg.ID, m.cfg.ID)

		} else {
			log.Debugf("no filters found for device %s and measurment %s", d.cfg.ID, m.cfg.ID)
		}
		err := m.Init(mfilter)
		if err != nil {
			log.Errorf("Error on initialize Measurement %s , Error:%s no data will be gathered for this measurement", m.cfg.ID, err)
			//d.InfmeasArray = append(d.InfmeasArray[:i], d.InfmeasArray[i+1:]...)
		}

	}
	//Initialize all snmpMetrics  objects and OID array
	//get data first time
	// useful to inicialize counter all value and test device snmp availability
	d.log.Debugf("SNMP Info: %+v", d.snmpClient)
	for _, m := range d.InfmeasArray {
		//if m.cfg.GetMode == "value" || d.cfg.SnmpVersion == "1" {
		if m.cfg.GetMode == "value" {
			_, _, err := m.SnmpGetData()
			if err != nil {
				d.log.Errorf("SNMP First Get Data error for host: %s", d.cfg.Host)
			}
		} else {
			_, _, err := m.SnmpWalkData()
			if err != nil {
				d.log.Errorf("SNMP First Get Data error for host: %s", d.cfg.Host)
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
					//if m.cfg.GetMode == "value" || d.cfg.SnmpVersion == "1" {
					if m.cfg.GetMode == "value" {
						nGets, nErrors, _ = m.SnmpGetData()
					} else {
						nGets, nErrors, _ = m.SnmpWalkData()
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
