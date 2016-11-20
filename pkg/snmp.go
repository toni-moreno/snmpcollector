package main

import (
	ers "errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/soniah/gosnmp"
	olog "log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// SysInfo basic information for any SNMP device
type SysInfo struct {
	SysDescr    string
	SysUptime   time.Duration
	SysContact  string
	SysName     string
	SysLocation string
}

// SnmpDebugLog returns a logger handler for snmp debug data
func SnmpDebugLog(filename string) *olog.Logger {
	name := filepath.Join(logDir, "snmpdebug_"+strings.Replace(filename, ".", "-", -1)+".log")
	if l, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644); err == nil {
		return olog.New(l, "", 0)
	} else {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
}

// SnmpGetSysInfo got system basic info from a snmp client
func SnmpGetSysInfo(id string, client *gosnmp.GoSNMP, l *logrus.Logger) (SysInfo, error) {
	//Get Basic System Info
	// SysDescr     .1.3.6.1.2.1.1.1.0
	// sysUpTime    .1.3.6.1.2.1.1.3.0
	// SysContact   .1.3.6.1.2.1.1.4.0
	// SysName      .1.3.6.1.2.1.1.5.0
	// SysLocation  .1.3.6.1.2.1.1.6.0
	sysOids := []string{
		".1.3.6.1.2.1.1.1.0",
		".1.3.6.1.2.1.1.3.0",
		".1.3.6.1.2.1.1.4.0",
		".1.3.6.1.2.1.1.5.0",
		".1.3.6.1.2.1.1.6.0"}

	info := SysInfo{SysDescr: "", SysUptime: time.Duration(0), SysContact: "", SysName: "", SysLocation: ""}
	pkt, err := client.Get(sysOids)

	if err != nil {
		l.Errorf("Error on getting initial basic system, Info to device %s: %s", id, err)
		return info, err
	}

	for idx, pdu := range pkt.Variables {
		l.Debugf("DEBUG pdu:%+v", pdu)
		if pdu.Value == nil {
			continue
		}
		switch idx {
		case 0: // SysDescr     .1.3.6.1.2.1.1.1.0
			if pdu.Type == gosnmp.OctetString {
				info.SysDescr = string(pdu.Value.([]byte))
			} else {
				l.Warnf("Error on getting system %s SysDescr return data of type %v", id, pdu.Type)
			}
		case 1: // sysUpTime    .1.3.6.1.2.1.1.3.0
			if pdu.Type == gosnmp.TimeTicks {
				seconds := uint32(pdu.Value.(int)) / 100
				info.SysUptime = time.Duration(seconds) * time.Second
			} else {
				l.Warnf("Error on getting system %s SysDescr return data of type %v", id, pdu.Type)
			}
		case 2: // SysContact   .1.3.6.1.2.1.1.4.0
			if pdu.Type == gosnmp.OctetString {
				info.SysContact = string(pdu.Value.([]byte))
			} else {
				l.Warnf("Error on getting system %s SysContact return data of type %v", id, pdu.Type)
			}
		case 3: // SysName      .1.3.6.1.2.1.1.5.0
			if pdu.Type == gosnmp.OctetString {
				info.SysName = string(pdu.Value.([]byte))
			} else {
				l.Warnf("Error on getting system %s SysName return data of type %v", id, pdu.Type)
			}
		case 4: // SysLocation  .1.3.6.1.2.1.1.6.0
			if pdu.Type == gosnmp.OctetString {
				info.SysLocation = string(pdu.Value.([]byte))
			} else {
				l.Warnf("Error on getting system %s SysLocation return data of type %v", id, pdu.Type)
			}
		}
	}
	//sometimes (authenticacion error on v3) client.get doesn't return error but the connection is not still available
	if len(info.SysDescr) == 0 && info.SysUptime == 0 {
		return info, fmt.Errorf("Some Error happened while getting system info for device %s", id)
	}
	return info, nil
}

func pduVal2str(pdu gosnmp.SnmpPDU) string {
	value := pdu.Value
	if pdu.Type == gosnmp.OctetString {
		return string(value.([]byte))
	} else {
		return ""
	}
}

func pduVal2Int64(pdu gosnmp.SnmpPDU) int64 {
	value := pdu.Value
	var val int64
	//revisar esta asignación
	switch value := value.(type) { // shadow
	case int:
		val = int64(value)
	case int8:
		val = int64(value)
	case int16:
		val = int64(value)
	case int32:
		val = int64(value)
	case int64:
		val = int64(value)
	case uint:
		val = int64(value)
	case uint8:
		val = int64(value)
	case uint16:
		val = int64(value)
	case uint32:
		val = int64(value)
	case uint64:
		val = int64(value)
	case string:
		// for testing and other apps - numbers may appear as strings
		var err error
		if val, err = strconv.ParseInt(value, 10, 64); err != nil {
			return val
		}
	default:
		return 0
	}
	return val
}

func pduVal2Hwaddr(pdu gosnmp.SnmpPDU) (string, error) {
	value := pdu.Value
	switch vt := value.(type) {
	case string:
		value = net.HardwareAddr(vt).String()
	case []byte:
		value = net.HardwareAddr(vt).String()
	default:
		return "", fmt.Errorf("invalid type (%T) for hwaddr conversion", value)
	}
	return string(value.([]byte)), nil
}

func pduVal2IPaddr(pdu gosnmp.SnmpPDU) (string, error) {
	var ipbs []byte
	value := pdu.Value
	switch vt := value.(type) {
	case string:
		ipbs = []byte(vt)
	case []byte:
		ipbs = vt
	default:
		return "", fmt.Errorf("invalid type (%T) for ipaddr conversion", value)
	}

	switch len(ipbs) {
	case 4, 16:
		value = net.IP(ipbs).String()
	default:
		return "", fmt.Errorf("invalid length (%d) for ipaddr conversion", len(ipbs))
	}

	return string(value.([]byte)), nil
}

const (
	maxOids = 60 // const in gosnmp
)

func SnmpClient(s *SnmpDeviceCfg, l *logrus.Logger) (*gosnmp.GoSNMP, *SysInfo, error) {
	var client *gosnmp.GoSNMP
	hostIPs, err := net.LookupHost(s.Host)
	if err != nil {
		l.Errorf("Error on Name Lookup for host: %s  ERROR: %s", s.Host, err)
		return nil, nil, err
	}
	if len(hostIPs) == 0 {
		l.Errorf("Error on Name Lookup for host: %s ", s.Host)
		return nil, nil, ers.New("Error on Name Lookup for host :" + s.Host)
	}
	if len(hostIPs) > 1 {
		l.Warnf("Lookup for %s host has more than one IP: %v => Finally used first IP %s", s.Host, hostIPs, hostIPs[0])
	}
	switch s.SnmpVersion {
	case "1":
		client = &gosnmp.GoSNMP{
			Target:  hostIPs[0],
			Port:    uint16(s.Port),
			Version: gosnmp.Version1,
			Timeout: time.Duration(s.Timeout) * time.Second,
			Retries: s.Retries,
		}
	case "2c":
		//validate community
		if len(s.Community) < 1 {
			l.Errorf("Error no community found %s in host %s", s.Community, s.Host)
			return nil, nil, ers.New("Error on snmp community")
		}

		client = &gosnmp.GoSNMP{
			Target:    hostIPs[0],
			Port:      uint16(s.Port),
			Community: s.Community,
			Version:   gosnmp.Version2c,
			Timeout:   time.Duration(s.Timeout) * time.Second,
			Retries:   s.Retries,
		}
	case "3":
		seclpmap := map[string]gosnmp.SnmpV3MsgFlags{
			"NoAuthNoPriv": gosnmp.NoAuthNoPriv,
			"AuthNoPriv":   gosnmp.AuthNoPriv,
			"AuthPriv":     gosnmp.AuthPriv,
		}
		authpmap := map[string]gosnmp.SnmpV3AuthProtocol{
			"NoAuth": gosnmp.NoAuth,
			"MD5":    gosnmp.MD5,
			"SHA":    gosnmp.SHA,
		}
		privpmap := map[string]gosnmp.SnmpV3PrivProtocol{
			"NoPriv": gosnmp.NoPriv,
			"DES":    gosnmp.DES,
			"AES":    gosnmp.AES,
		}
		UsmParams := new(gosnmp.UsmSecurityParameters)

		if len(s.V3AuthUser) < 1 {
			l.Errorf("Error username not found in snmpv3 %s in host %s", s.V3AuthUser, s.Host)
			return nil, nil, ers.New("Error on snmp v3 user")
		}

		switch s.V3SecLevel {

		case "NoAuthNoPriv":
			UsmParams = &gosnmp.UsmSecurityParameters{
				UserName:               s.V3AuthUser,
				AuthenticationProtocol: gosnmp.NoAuth,
				PrivacyProtocol:        gosnmp.NoPriv,
			}
		case "AuthNoPriv":
			if len(s.V3AuthPass) < 1 {
				l.Errorf("Error password not found in snmpv3 %s in host %s", s.V3AuthUser, s.Host)
				return nil, nil, ers.New("Error on snmp v3 AuthPass")
			}

			//validate correct s.authuser

			if val, ok := authpmap[s.V3AuthProt]; !ok {
				l.Errorf("Error in Auth Protocol %v | %v  in host %s", s.V3AuthProt, val, s.Host)
				return nil, nil, ers.New("Error on snmp v3 AuthProt")
			}

			//validate s.authpass s.authprot
			UsmParams = &gosnmp.UsmSecurityParameters{
				UserName:                 s.V3AuthUser,
				AuthenticationProtocol:   authpmap[s.V3AuthProt],
				AuthenticationPassphrase: s.V3AuthPass,
				PrivacyProtocol:          gosnmp.NoPriv,
			}
		case "AuthPriv":
			//validate s.authpass s.authprot

			if len(s.V3AuthPass) < 1 {
				l.Errorf("Error password not found in snmpv3 %s in host %s", s.V3AuthUser, s.Host)
				return nil, nil, ers.New("Error on snmp v3 AuthPass")
			}

			if val, ok := authpmap[s.V3AuthProt]; !ok {
				l.Errorf("Error in Auth Protocol %v | %v  in host %s", s.V3AuthProt, val, s.Host)
				return nil, nil, ers.New("Error on snmp v3 AuthProt")
			}

			//validate s.privpass s.privprot

			if len(s.V3PrivPass) < 1 {
				l.Errorf("Error privPass not found in snmpv3 %s in host %s", s.V3AuthUser, s.Host)
				//		log.Printf("DEBUG SNMP: %+v", *s)
				return nil, nil, ers.New("Error on snmp v3 PrivPAss")
			}

			if val, ok := privpmap[s.V3PrivProt]; !ok {
				l.Errorf("Error in Priv Protocol %v | %v  in host %s", s.V3PrivProt, val, s.Host)
				return nil, nil, ers.New("Error on snmp v3 AuthPass")
			}

			UsmParams = &gosnmp.UsmSecurityParameters{
				UserName:                 s.V3AuthUser,
				AuthenticationProtocol:   authpmap[s.V3AuthProt],
				AuthenticationPassphrase: s.V3AuthPass,
				PrivacyProtocol:          privpmap[s.V3PrivProt],
				PrivacyPassphrase:        s.V3PrivPass,
			}
		default:
			l.Errorf("Error no Security Level found %s in host %s", s.V3SecLevel, s.Host)
			return nil, nil, ers.New("Error on snmp Security Level")

		}
		client = &gosnmp.GoSNMP{
			Target:             hostIPs[0],
			Port:               uint16(s.Port),
			Version:            gosnmp.Version3,
			Timeout:            time.Duration(s.Timeout) * time.Second,
			Retries:            s.Retries,
			SecurityModel:      gosnmp.UserSecurityModel,
			MsgFlags:           seclpmap[s.V3SecLevel],
			SecurityParameters: UsmParams,
		}
	default:
		l.Errorf("Error no snmpversion found %s in host %s", s.SnmpVersion, s.Host)
		return nil, nil, ers.New("Error on snmp Version")
	}
	if s.SnmpDebug {
		client.Logger = SnmpDebugLog(s.ID)
	}
	//first connect
	err = client.Connect()
	if err != nil {
		l.Errorf("error on first connect %s", err)
		return nil, nil, err
	} else {
		l.Infof("First SNMP connection to host  %s stablished", s.Host)
	}
	//first snmp query
	si, err := SnmpGetSysInfo(s.ID, client, l)
	if err != nil {
		l.Errorf("error on get System Info %s", err)
		return nil, nil, err
	} else {
		l.Infof("Got basic system info %#v ", si)
	}
	return client, &si, err
}
