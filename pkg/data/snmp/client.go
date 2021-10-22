package snmp

import (
	"fmt"
	"strings"

	"github.com/gosnmp/gosnmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
)

// SNMPV3Params store SNMP configuration related only with protocol V3
type V3Params struct {
	SecLevel        string
	AuthUser        string
	AuthPass        string
	PrivPass        string
	PrivProt        string
	AuthProt        string
	ContextName     string
	ContextEngineID string
}

// ConnectionParams store all needed information to create a new snmpGo client
type ConnectionParams struct {
	// Host is the hostname or IP of the device
	Host string
	// Port where the SNMP connection should be established
	Port int
	// Timeout is the timeout for one SNMP request/response, in seconds
	Timeout int
	// Retries is the number of retries to attempt.
	Retries int
	// SnmpVersion which SNMP protocol version use. Allowed values: 1, 2c or 3
	SnmpVersion string
	// Community is an SNMP Community string. Optinal in v1, mandatory in v2c, not used in v3.
	Community string
	// MaxRepetitions sets the GETBULK max-repetitions used by BulkWalk*. Only for v2c or v3.
	MaxRepetitions uint8
	// MaxOids is the number of OIDs returned in a single Get call to the goSNMP client
	MaxOids int
	// Debug if enabled, log all SNMP operations to a new file.
	Debug bool
	// V3Params store specific values only for SNMP v3
	V3Params V3Params
}

type Client struct {
	snmpClient *gosnmp.GoSNMP
	// Log is the logger used to trace this client
	Log utils.Logger
	// ID used to generate the debug file name
	ID          string
	DisableBulk bool
	// Walk        func(string, gosnmp.WalkFunc) error // TODO usar un condicional cuando haga falta usar Walk en vez de definir esta func
	// Info        *SysInfo // TODO esto que se devuelva, pero no que se almacene
	// Connected define if the this client is considered Connected
	Connected        bool
	ConnectionParams ConnectionParams
}

// Validation check if SNMP parameters are valid to establish a SNMP connection.
func (c ConnectionParams) Validation() error {
	if c.SnmpVersion != "1" && c.SnmpVersion != "2c" && c.SnmpVersion != "3" {
		return fmt.Errorf("invalid snmp version: %s", c.SnmpVersion)
	}

	if c.SnmpVersion == "2c" {
		if c.Community == "" {
			return fmt.Errorf("snmp community its mandatory for v2c clients")
		}
	} else if c.SnmpVersion == "3" {

		if c.V3Params.AuthUser == "" {
			return fmt.Errorf("AuthUser is mandatory for v3 clients")
		}

		// Check if sec level is one the allowed values
		if _, ok := seclpmap[c.V3Params.SecLevel]; !ok {
			return fmt.Errorf("unknown SecLevel for SNMP v3: %v", c.V3Params.SecLevel)
		}

		if c.V3Params.SecLevel == "AuthNoPriv" || c.V3Params.SecLevel == "AuthPriv" {
			// AuthPass should be defined for Auth* level
			if c.V3Params.AuthPass == "" {
				return fmt.Errorf("AuthPass is mandatory for v3 clients with SecLevel AuthPriv or AuthNoPriv")
			}

			// Check if auth protocol is one the allowed values
			if _, ok := authpmap[c.V3Params.AuthProt]; !ok {
				return fmt.Errorf("unknown AuthProt for SNMP v3: %v", c.V3Params.AuthProt)
			}
		}

		if c.V3Params.SecLevel == "AuthPriv" {
			if c.V3Params.PrivPass == "" {
				return fmt.Errorf("PrivPass is mandatory for v3 clients with SecLevel AuthPriv")
			}

			// Check if priv protocol is one the allowed values
			if _, ok := privpmap[c.V3Params.PrivProt]; !ok {
				return fmt.Errorf("unknown PrivProt for SNMP v3: %v", c.V3Params.PrivProt)
			}
		}
	}

	return nil
}

// Release release the GoSNMP object
func (c *Client) Release() error {
	c.Connected = false
	if c.snmpClient != nil {
		return c.snmpClient.Conn.Close()
	}
	return nil
}

// Connect using the info stored in the struct, generate the goSNMP client and make the first connection to the
// device to check if it works.
// It also try to obtain some basic OIDs to check if everything works.
func (c *Client) Connect(systemOIDs []string) (*SysInfo, error) {
	goSNMPClient, err := GetClient(c.ConnectionParams, c.Log)
	if err != nil {
		return nil, fmt.Errorf("initializating the goSNMP client: %v", err)
	}

	c.snmpClient = goSNMPClient

	sysinfo, err := c.SysInfoQuery(systemOIDs) // TODO meter los systemOIDs
	if err != nil {
		return nil, fmt.Errorf("obtaining the sysInfo: %v", err)
	}

	return sysinfo, nil
}

// SysInfoQuery return the SysInfo from the device.
// It is used as a test to check if the device is working.
// It will query a set of default basic OIDs (check GetSysInfo), or the defined in the parameter if it
// is not empty and its first element is not the string "null"
func (c *Client) SysInfoQuery(systemOIDs []string) (*SysInfo, error) {
	if len(systemOIDs) > 0 && systemOIDs[0] != "" && systemOIDs[0] != "null" {
		c.Log.Infof("Detected alternate %d SystemOID's ", len(systemOIDs))
		// this device has an alternate System Description (Non MIB-2 based systems)
		si, err := GetAlternateSysInfo(c.snmpClient, c.Log, systemOIDs)
		if err != nil {
			c.Log.Errorf("error on get Alternate System Info ERROR [%s] for OID's [%s] ", err, strings.Join(systemOIDs[:], ","))
			return nil, err
		}
		c.Log.Infof("Got basic system info %#v ", si)
		return &si, err
	}
	// For most devices System Description could be got with MIB-2::System base OID's
	si, err := GetSysInfo(c.snmpClient, c.Log)
	if err != nil {
		c.Log.Errorf("error on get System Info %s", err)
		return nil, err
	}
	c.Log.Infof("Got basic system info %#v ", si)

	return &si, err
}

func (c *Client) Target() string {
	return c.snmpClient.Target
}

// Walk selects how to gather data based on the SNMP version and a custom flag
func (c *Client) Walk(rootOid string, walkFn gosnmp.WalkFunc) error {
	if c.snmpClient.Version == gosnmp.Version1 || c.DisableBulk {
		return c.snmpClient.Walk(rootOid, walkFn)
	}
	return c.snmpClient.BulkWalk(rootOid, walkFn)
}

// SetSnmpClient get the values of the list of OIDs in groups of c.MaxOids.
// Send each value to the walkfunc (second parameter).
func (c *Client) Get(oids []string, walkFunc gosnmp.WalkFunc) error {
	l := len(oids)
	c.Log.Infof("LEN %d : %+v | client : %+v", l, oids, c)

	// Get values in groups of c.MaxOids
	for i := 0; i < l; i += c.snmpClient.MaxOids {
		end := i + c.snmpClient.MaxOids
		if end > l {
			end = len(oids)
		}
		c.Log.Infof("Getting snmp data from %d to %d", i, end)
		pkt, err := c.snmpClient.Get(oids[i:end])
		if err != nil {
			c.Log.Debugf("selected OIDS %+v", oids[i:end])
			c.Log.Errorf("SNMP (%s) for OIDs (%d/%d) get error: %s\n", c.snmpClient.Target, i, end, err)
			continue
		}

		for _, pdu := range pkt.Variables {
			c.Log.Debugf("DEBUG pdu [%+v] || Value type %T [%x] ", pdu, pdu.Value, pdu.Type)
			if pdu.Value == nil {
				continue
			}
			if err := walkFunc(pdu); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) Query(mode string, oid string) ([]EasyPDU, error) {
	return Query(c.snmpClient, mode, oid)
}
