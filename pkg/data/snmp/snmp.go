package snmp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

var (
	mainlog utils.Logger
	logDir  string
	// seclpmap define the allowed SecLevels for SNMP v3
	seclpmap = map[string]gosnmp.SnmpV3MsgFlags{
		"NoAuthNoPriv": gosnmp.NoAuthNoPriv,
		"AuthNoPriv":   gosnmp.AuthNoPriv,
		"AuthPriv":     gosnmp.AuthPriv,
	}
	// authpmap define the allowed auth proto for SNMP v3
	authpmap = map[string]gosnmp.SnmpV3AuthProtocol{
		"NoAuth": gosnmp.NoAuth,
		"MD5":    gosnmp.MD5,
		"SHA":    gosnmp.SHA,
		"SHA224": gosnmp.SHA224,
		"SHA256": gosnmp.SHA256,
		"SHA384": gosnmp.SHA384,
		"SHA512": gosnmp.SHA512,
	}
	// privpmap define the allowed priv proto for SNMP v3
	privpmap = map[string]gosnmp.SnmpV3PrivProtocol{
		"NoPriv":  gosnmp.NoPriv,
		"DES":     gosnmp.DES,
		"AES":     gosnmp.AES,
		"AES192":  gosnmp.AES192,
		"AES192C": gosnmp.AES192C,
		"AES256":  gosnmp.AES256,
		"AES256C": gosnmp.AES256C,
	}
)

// SetLogger xx
func SetLogger(l utils.Logger) {
	mainlog = l
}

// SetLogDir xx
func SetLogDir(dir string) {
	logDir = dir
}

// SysInfo Info basic information for any SNMP based MIB-2 System
type SysInfo struct {
	SysDescr    string
	SysUptime   time.Duration
	SysContact  string
	SysName     string
	SysLocation string
}

// PduVal2BoolArray get boolean value from PDU
func PduVal2BoolArray(pdu gosnmp.SnmpPDU) []bool {
	data := pdu.Value.([]byte)
	// mainlog.Errorf("PduVal2BoolArray: %+v\n", data)
	barray := make([]bool, len(data)*8)
	cnt := 0
	for _, d := range data {
		for i := 0; i < 8; i++ {
			if (d & 0x80) == 0x80 {
				barray[cnt] = true
			} else {
				barray[cnt] = false
			}
			d <<= 1
			cnt++
		}
	}
	return barray
}

// PduVal2Cooked to get data from any defined type in gosnmp
func PduVal2Cooked(pdu gosnmp.SnmpPDU) interface{} {
	switch pdu.Type {
	case gosnmp.EndOfContents:
		return pdu.Value
	case gosnmp.Boolean:
		return pdu.Value
	case gosnmp.Integer:
		return PduVal2Int64(pdu)
	case gosnmp.BitString:
		return PduVal2str(pdu)
	case gosnmp.OctetString:
		return PduVal2str(pdu)
	case gosnmp.Null:
		return pdu.Value
	case gosnmp.ObjectIdentifier:
		//	log.Debugf("DEBUG ObjectIdentifier :%s", pdu.Value)
		return pdu.Value
	case gosnmp.ObjectDescription:
		return PduVal2str(pdu)
	case gosnmp.IPAddress:
		return PduVal2str(pdu)
	case gosnmp.Counter32:
		return PduVal2Int64(pdu)
	case gosnmp.Gauge32:
		return PduVal2Int64(pdu)
	case gosnmp.TimeTicks:
		return PduVal2Int64(pdu)
	case gosnmp.Opaque:
		return pdu.Value
	case gosnmp.NsapAddress:
		return PduVal2str(pdu)
	case gosnmp.Counter64:
		return PduVal2Int64(pdu)
	case gosnmp.Uinteger32:
		return PduVal2Int64(pdu)
	case gosnmp.NoSuchObject:
		return pdu.Value
	case gosnmp.NoSuchInstance:
		return "No Such Instance currently exists at this OID"
	case gosnmp.EndOfMibView:
		return pdu.Value
	default:
		return "--"
	}
}

// PduType2Str  type to string
func PduType2Str(pdutype gosnmp.Asn1BER) string {
	switch pdutype {
	case gosnmp.EndOfContents: // 	case gosnmp.UnknownType
		return "EndOfContents"
	case gosnmp.Boolean:
		return "Boolean"
	case gosnmp.Integer:
		return "Integer"
	case gosnmp.BitString:
		return "BitString"
	case gosnmp.OctetString:
		return "OctetString"
	case gosnmp.Null:
		return "Null"
	case gosnmp.ObjectIdentifier:
		return "ObjectIdentifier"
	case gosnmp.ObjectDescription:
		return "ObjectDescription"
	case gosnmp.IPAddress:
		return "IPaddress"
	case gosnmp.Counter32:
		return "Counter32"
	case gosnmp.Gauge32:
		return "Gauge32"
	case gosnmp.TimeTicks:
		return "TimeTicks"
	case gosnmp.Opaque:
		return "Opaque"
	case gosnmp.NsapAddress:
		return "NsapAddress"
	case gosnmp.Counter64:
		return "Counter64"
	case gosnmp.Uinteger32:
		return "Uinteger32"
	case gosnmp.NoSuchObject:
		return "NoSuchObject"
	case gosnmp.NoSuchInstance:
		return "NoSuchInstance"
	case gosnmp.EndOfMibView:
		return "EnvOfMibView"
	default:
		return "--"
	}
}

// EasyPDU enable user interface Info for OID data
type EasyPDU struct {
	Name  string
	Type  string
	Value interface{}
}

// Query enable arbitrary SNMP querys over the client
func Query(client *gosnmp.GoSNMP, mode string, oid string) ([]EasyPDU, error) {
	var result []EasyPDU
	switch mode {
	case "get":
		pkt, err := client.Get([]string{oid})
		if err != nil {
			mainlog.Errorf("SNMP (%s) for OIDs get error: %s\n", client.Target, err)
			return result, err
		}
		for _, pdu := range pkt.Variables {
			result = append(result, EasyPDU{Name: pdu.Name, Type: PduType2Str(pdu.Type), Value: PduVal2Cooked(pdu)})
		}
	case "walk":
		setRawData := func(pdu gosnmp.SnmpPDU) error {
			if pdu.Value == nil {
				mainlog.Warnf("no value retured by pdu :%+v", pdu)
				return nil // if error return the bulk process will stop
			}
			result = append(result, EasyPDU{Name: pdu.Name, Type: PduType2Str(pdu.Type), Value: PduVal2Cooked(pdu)})
			return nil
		}
		err := client.Walk(oid, setRawData)
		if err != nil {
			mainlog.Errorf("SNMP WALK error: %s", err)
			return result, err
		}
	default:
		return result, fmt.Errorf("error on getmode parameter [%s] not supported", mode)
	}
	if len(result) == 0 {
		result = append(result, EasyPDU{Name: oid, Type: "ERROR", Value: "No Such Instance currently exists at this OID"})
	}
	return result, nil
}

// GetAlternateSysInfo got system basic info from a snmp client when sysinfo should be take from specified OID's
func GetAlternateSysInfo(client *gosnmp.GoSNMP, l utils.Logger, SystemOIDs []string) (SysInfo, error) {
	// Get System Info from Alternate SystemOIDs
	sysOids := []string{}
	sysOidsiMap := make(map[string]string) // inverse map to get Key name from OID

	info := SysInfo{SysDescr: "", SysUptime: time.Duration(0), SysContact: "", SysName: "", SysLocation: ""}

	for _, v := range SystemOIDs {
		s := strings.Split(v, "=")
		if len(s) == 2 {
			key, value := s[0], s[1]
			// add initial dot to the OID if it has not.
			if strings.HasPrefix(value, ".") == false {
				value = "." + value
			}
			sysOids = append(sysOids, value)
			sysOidsiMap[value] = key
		} else {
			return info, fmt.Errorf("Error on AlternateSystemInfo OID definition TAG=OID [ %s ]", v)
		}
	}

	pkt, err := client.Get(sysOids)
	if err != nil {
		l.Errorf("Error on getting initial basic system info: %s", err)
		return info, err
	}

	tmpDesc := []string{}

	for _, pdu := range pkt.Variables {
		l.Debugf("DEBUG pdu:%+v", pdu)
		if pdu.Value == nil {
			continue
		}
		oidname := ""
		if val, ok := sysOidsiMap[pdu.Name]; ok {
			oidname = val
		}
		switch pdu.Type {
		case gosnmp.OctetString: //  like SysDescr
			value := fmt.Sprintf("%s = %s", oidname, string(pdu.Value.([]byte)))
			tmpDesc = append(tmpDesc, value)
		case gosnmp.TimeTicks: // like sysUpTime
			seconds := uint32(pdu.Value.(uint32)) / 100
			value := fmt.Sprintf("%s = %d seconds", oidname, seconds)
			tmpDesc = append(tmpDesc, value)
		}
	}
	info.SysDescr = strings.Join(tmpDesc[:], " | ")

	// sometimes (authenticacion error on v3) client.get doesn't return error but the connection is not still available
	if info.SysDescr == "" {
		return info, fmt.Errorf("Some Error happened while getting alternate system info")
	}
	return info, nil
}

// Mejor, retornar los errores y logar aguas arriba.
// GetSysInfo got system basic info from a snmp client
func GetSysInfo(client *gosnmp.GoSNMP, l utils.Logger) (SysInfo, error) {
	// Get Basic System Info
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
		".1.3.6.1.2.1.1.6.0",
	}

	info := SysInfo{SysDescr: "", SysUptime: time.Duration(0), SysContact: "", SysName: "", SysLocation: ""}
	pkt, err := client.Get(sysOids)
	if err != nil {
		return info, fmt.Errorf("client get: %s", err)
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
				l.Warnf("Error on getting SysDescr, return data of type %v", pdu.Type)
			}
		case 1: // sysUpTime    .1.3.6.1.2.1.1.3.0
			if pdu.Type == gosnmp.TimeTicks {
				seconds := uint32(pdu.Value.(uint32)) / 100
				info.SysUptime = time.Duration(seconds) * time.Second
			} else {
				l.Warnf("Error on getting SysUptime, return data of type %v", pdu.Type)
			}
		case 2: // SysContact   .1.3.6.1.2.1.1.4.0
			if pdu.Type == gosnmp.OctetString {
				info.SysContact = string(pdu.Value.([]byte))
			} else {
				l.Warnf("Error on getting SysContact, return data of type %v", pdu.Type)
			}
		case 3: // SysName      .1.3.6.1.2.1.1.5.0
			if pdu.Type == gosnmp.OctetString {
				info.SysName = string(pdu.Value.([]byte))
			} else {
				l.Warnf("Error on getting SysName, return data of type %v", pdu.Type)
			}
		case 4: // SysLocation  .1.3.6.1.2.1.1.6.0
			if pdu.Type == gosnmp.OctetString {
				info.SysLocation = string(pdu.Value.([]byte))
			} else {
				l.Warnf("Error on getting SysLocation, return data of type %v", pdu.Type)
			}
		}
	}
	// sometimes (authenticacion error on v3) client.get doesn't return error but the connection is not still available
	if info.SysDescr == "" && info.SysUptime == 0 {
		return info, fmt.Errorf("Some error happened while getting system info")
	}
	return info, nil
}

// PduVal2str transform PDU data to string
func PduVal2str(pdu gosnmp.SnmpPDU) string {
	switch pdu.Type {
	case gosnmp.Integer:
		return strconv.FormatInt(PduVal2Int64(pdu), 10)
	case gosnmp.Counter32:
		return strconv.FormatInt(PduVal2Int64(pdu), 10)
	case gosnmp.Gauge32:
		return strconv.FormatInt(PduVal2Int64(pdu), 10)
	case gosnmp.TimeTicks:
		return strconv.FormatInt(PduVal2Int64(pdu), 10)
	case gosnmp.Counter64:
		return strconv.FormatInt(PduVal2Int64(pdu), 10)
	case gosnmp.Uinteger32:
		return strconv.FormatInt(PduVal2Int64(pdu), 10)
	case gosnmp.OctetString:
		return string(pdu.Value.([]byte))
	case gosnmp.ObjectIdentifier:
		return PduVal2OID(pdu)
	case gosnmp.IPAddress:
		ip, err := PduVal2IPaddr(pdu)
		if err != nil {
			mainlog.Errorf("Error on SNMP IPAddress decode on PDU[%#+v] Error: %s\n", pdu, err)
		}
		return ip
	default:
		return ""
	}
}

// PduValHexString2Uint transform PDU HexString to uint64
func PduValHexString2Uint(pdu gosnmp.SnmpPDU) (uint64, error) {
	value := pdu.Value
	if pdu.Type == gosnmp.OctetString {
		result, err := strconv.ParseUint(string(value.([]byte)), 16, 64)
		return result, err
	}
	return 0, fmt.Errorf("The PDU scanned is not and OctetString as expected")
}

// PduVal2OID transform PDU data to string
func PduVal2OID(pdu gosnmp.SnmpPDU) string {
	value := pdu.Value
	if pdu.Type == gosnmp.ObjectIdentifier {
		return value.(string)
	}
	return ""
}

// PduVal2Int64 transform PDU data to Int64
func PduVal2Int64(pdu gosnmp.SnmpPDU) int64 {
	value := pdu.Value
	var val int64
	// revisar esta asignación
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

// PduVal2UInt64 transform data to Uint64
func PduVal2UInt64(pdu gosnmp.SnmpPDU) uint64 {
	value := pdu.Value
	var val uint64
	// revisar esta asignación
	switch value := value.(type) { // shadow
	case int:
		val = uint64(value)
	case int8:
		val = uint64(value)
	case int16:
		val = uint64(value)
	case int32:
		val = uint64(value)
	case int64:
		val = uint64(value)
	case uint:
		val = uint64(value)
	case uint8:
		val = uint64(value)
	case uint16:
		val = uint64(value)
	case uint32:
		val = uint64(value)
	case uint64:
		val = uint64(value)
	case string:
		// for testing and other apps - numbers may appear as strings
		var err error
		if val, err = strconv.ParseUint(value, 10, 64); err != nil {
			return val
		}
	default:
		return 0
	}
	return val
}

// PduVal2Hwaddr transform data to MAC address
func PduVal2Hwaddr(pdu gosnmp.SnmpPDU) (string, error) {
	value := pdu.Value
	switch vt := value.(type) {
	case string:
		value = net.HardwareAddr(vt).String()
	case []byte:
		value = net.HardwareAddr(vt).String()
	default:
		return "", fmt.Errorf("invalid type (%T) for hwaddr conversion", value)
	}
	return value.(string), nil
}

// PduVal2IPaddr transform data o IP address
func PduVal2IPaddr(pdu gosnmp.SnmpPDU) (string, error) {
	var ipbs []byte
	value := pdu.Value
	switch vt := value.(type) {
	case string:
		ipbs = []byte(vt)
		return string(ipbs), nil
	case []byte:
		ipbs = vt
		var ipvalue string
		switch len(ipbs) {
		case 4, 16:
			ipvalue = net.IP(ipbs).String()
		default:
			return "", fmt.Errorf("invalid length (%d) for ipaddr conversion", len(ipbs))
		}
		return ipvalue, nil
	default:
		return "", fmt.Errorf("invalid type (%T) for ipaddr conversion", value)
	}
}

// Release release the GoSNMP object
func Release(client *gosnmp.GoSNMP) {
	if client != nil {
		client.Conn.Close()
	}
}

// GetClient return the gosnmp client configured.
// To connect, the host is resolved and the first IP is used.
func GetClient(connectionParams ConnectionParams, l utils.Logger) (*gosnmp.GoSNMP, error) {
	hostIPs, err := net.LookupHost(connectionParams.Host)
	if err != nil {
		return nil, fmt.Errorf("name lookup: %v", err)
	}
	if len(hostIPs) == 0 {
		return nil, fmt.Errorf("empty name lookup response")
	}

	if len(hostIPs) > 1 {
		l.Warnf("Lookup has more than one IP: %v. Using the first IP: %s", hostIPs, hostIPs[0])
	}

	// Common options
	client := &gosnmp.GoSNMP{
		Target:  hostIPs[0],
		Port:    uint16(connectionParams.Port),
		Timeout: time.Duration(connectionParams.Timeout) * time.Second,
		Retries: connectionParams.Retries,
		MaxOids: connectionParams.MaxOids,
	}

	switch connectionParams.SnmpVersion {
	case "1":
		client.Version = gosnmp.Version1
		client.Community = connectionParams.Community
	case "2c":
		client.Version = gosnmp.Version2c
		client.Community = connectionParams.Community
		client.MaxRepetitions = uint32(connectionParams.MaxRepetitions)
	case "3":
		client.Version = gosnmp.Version3
		UsmParams := new(gosnmp.UsmSecurityParameters)

		switch connectionParams.V3Params.SecLevel {
		case "NoAuthNoPriv":
			UsmParams = &gosnmp.UsmSecurityParameters{
				UserName:               connectionParams.V3Params.AuthUser,
				AuthenticationProtocol: gosnmp.NoAuth,
				PrivacyProtocol:        gosnmp.NoPriv,
			}
		case "AuthNoPriv":
			UsmParams = &gosnmp.UsmSecurityParameters{
				UserName:                 connectionParams.V3Params.AuthUser,
				AuthenticationProtocol:   authpmap[connectionParams.V3Params.AuthProt],
				AuthenticationPassphrase: connectionParams.V3Params.AuthPass,
				PrivacyProtocol:          gosnmp.NoPriv,
			}
		case "AuthPriv":
			UsmParams = &gosnmp.UsmSecurityParameters{
				UserName:                 connectionParams.V3Params.AuthUser,
				AuthenticationProtocol:   authpmap[connectionParams.V3Params.AuthProt],
				AuthenticationPassphrase: connectionParams.V3Params.AuthPass,
				PrivacyProtocol:          privpmap[connectionParams.V3Params.PrivProt],
				PrivacyPassphrase:        connectionParams.V3Params.PrivPass,
			}
		default:
			panic("Invalid SNMP v3 SecLevel. Code should never reach here. Validation should control it")
		}

		client.MaxRepetitions = uint32(connectionParams.MaxRepetitions)
		client.SecurityModel = gosnmp.UserSecurityModel
		client.MsgFlags = seclpmap[connectionParams.V3Params.SecLevel]
		client.SecurityParameters = UsmParams
		client.ContextName = connectionParams.V3Params.ContextName
		client.ContextEngineID = connectionParams.V3Params.ContextEngineID
	default:
		panic("Invalid SNMP version. Code should never reach here. Validation should control it")
	}

	// first connect
	err = client.Connect()
	if err != nil {
		return nil, fmt.Errorf("goSNMP unable to connect: %v", err)
	}
	l.Infof("First SNMP connection stablished with MaxRepetitions set to %d", connectionParams.MaxRepetitions)

	return client, nil
}
