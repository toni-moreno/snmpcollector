package snmp

import (
	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/config"
)

type Client struct {
	snmpClient  *gosnmp.GoSNMP
	log         *logrus.Logger
	DisableBulk bool
	Walk        func(string, gosnmp.WalkFunc) error
	MaxOids     int
	Info        *SysInfo
}

func New(cfg *config.SnmpDeviceCfg, l *logrus.Logger, mkey string, debug bool, maxrep uint8) (*Client, error) {
	cli := &Client{}
	c, s, err := GetClient(cfg, l, mkey, debug, maxrep)
	if err != nil {
		return nil, err
	}
	cli.log = l
	cli.snmpClient = c
	cli.Info = s
	cli.DisableBulk = cfg.DisableBulk
	if cfg.MaxOids > 0 {
		cli.MaxOids = cfg.MaxOids
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

// SetSnmpClient set a GoSNMP client to the Measurement
func (c *Client) Get(oids []string, f gosnmp.WalkFunc) error {
	l := len(oids)
	c.log.Infof("LEN %d : %+v | client : %+v", l, oids, c)

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
			if err := f(pdu); err != nil {
				return err
			}
		}
	}

	return nil
}
