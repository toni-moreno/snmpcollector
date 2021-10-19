package snmp

import (
	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
)

type Client struct {
	snmpClient  *gosnmp.GoSNMP
	log         *logrus.Logger
	DisableBulk bool
	Walk        func(string, gosnmp.WalkFunc) error
	MaxOids     int
	Info        *SysInfo
}

// New create the snmp client and define some options for future queries
// While creating the snmp client (GetClient)
func New(
	host string,
	maxRepetitions uint8,
	snmpVersion string,
	community string,
	port int,
	timeout int,
	retries int,
	v3AuthUser string,
	v3SecLevel string,
	v3AuthPass string,
	v3PrivPass string,
	v3PrivProt string,
	v3AuthProt string,
	v3ContextName string,
	v3ContextEngineID string,
	id string,
	systemOIDs []string,
	maxOids int,
	disableBulk bool,
	l *logrus.Logger,
	mkey string,
	debug bool,
	maxrep uint8,
) (*Client, error) {
	cli := &Client{}
	c, s, err := GetClient(
		host,
		maxRepetitions,
		snmpVersion,
		community,
		port,
		timeout,
		retries,
		v3AuthUser,
		v3SecLevel,
		v3AuthPass,
		v3PrivPass,
		v3PrivProt,
		v3AuthProt,
		v3ContextName,
		v3ContextEngineID,
		id,
		systemOIDs,

		l,
		mkey,
		debug,
		maxrep,
	)
	if err != nil {
		return nil, err
	}
	cli.log = l
	cli.snmpClient = c
	cli.Info = s
	cli.DisableBulk = disableBulk
	if maxOids > 0 {
		cli.MaxOids = maxOids
	} else {
		cli.MaxOids = 60
	}

	switch {
	case cli.snmpClient.Version == gosnmp.Version1 || cli.DisableBulk:
		cli.Walk = cli.snmpClient.Walk
	default:
		cli.Walk = cli.snmpClient.BulkWalk
	}
	return cli, nil
}

// Release release the GoSNMP object
func (c *Client) Release() error {
	if c.snmpClient != nil {
		return c.snmpClient.Conn.Close()
	}
	return nil
}

func (c *Client) Connect() error {
	if c.snmpClient != nil {
		return c.snmpClient.Connect()
	}
	return nil
}

func (c *Client) Target() string {
	return c.snmpClient.Target
}

// SetSnmpClient get the values of the list of OIDs in groups of c.MaxOids.
// Send each value to the walkfunc (second parameter).
func (c *Client) Get(oids []string, walkFunc gosnmp.WalkFunc) error {
	l := len(oids)
	c.log.Infof("LEN %d : %+v | client : %+v", l, oids, c)

	// Get values in groups of c.MaxOids
	for i := 0; i < l; i += c.MaxOids {
		end := i + c.MaxOids
		if end > l {
			end = len(oids)
		}
		c.log.Infof("Getting snmp data from %d to %d", i, end)
		pkt, err := c.snmpClient.Get(oids[i:end])
		if err != nil {
			c.log.Debugf("selected OIDS %+v", oids[i:end])
			c.log.Errorf("SNMP (%s) for OIDs (%d/%d) get error: %s\n", c.snmpClient.Target, i, end, err)
			continue
		}

		for _, pdu := range pkt.Variables {
			c.log.Debugf("DEBUG pdu [%+v] || Value type %T [%x] ", pdu, pdu.Value, pdu.Type)
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
