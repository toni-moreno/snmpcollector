package mock

import (
	"fmt"
	"time"

	c "github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
)

func ExampleServerClientGet() {
	var err error
	log = logrus.New()
	log.Level = logrus.DebugLevel

	s := &SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []c.SnmpPDU{
			{Name: ".1.3.6.1.2.1.1.4.0", Type: c.Integer, Value: int(55)},
			{Name: ".1.3.6.1.2.1.1.7.0", Type: c.Integer, Value: int(56)},
		},
	}

	err = s.Start()
	if err != nil {
		log.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	c.Default.Version = c.Version2c
	c.Default.Community = "public"
	c.Default.Target = "127.0.0.1"
	c.Default.Port = 1161
	c.Default.Timeout = 5 * time.Second
	c.Default.Retries = 0
	c.Default.Logger = log
	err = c.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer c.Default.Conn.Close()

	oids := []string{"1.3.6.1.2.1.1.4.0", "1.3.6.1.2.1.1.7.0"}
	result, err2 := c.Default.Get(oids) // Get() accepts up to g.MAX_OIDS
	if err2 != nil {
		log.Errorf("Get() err: %v", err2)
		return
	}
	for _, v := range result.Variables {
		fmt.Printf("Result for [%s]:  %+v\n", v.Name, v.Value)
	}

	// Output:
	// Result for [.1.3.6.1.2.1.1.4.0]:  55
	// Result for [.1.3.6.1.2.1.1.7.0]:  56
}

func ExampleServerClientWalk() {
	var err error
	log = logrus.New()
	log.Level = logrus.DebugLevel

	s := &SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []c.SnmpPDU{
			{Name: ".1.1.1", Type: c.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: c.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: c.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: c.Integer, Value: int(54)},
			{Name: ".1.2.1", Type: c.OctetString, Value: "eth0"},
			{Name: ".1.2.2", Type: c.OctetString, Value: "eth1"},
			{Name: ".1.2.3", Type: c.OctetString, Value: "eth2"},
			{Name: ".1.2.4", Type: c.OctetString, Value: "eth3"},
		},
	}

	err = s.Start()
	if err != nil {
		log.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	c.Default.Version = c.Version2c
	c.Default.Community = "public"
	c.Default.Target = "127.0.0.1"
	c.Default.Port = 1161
	c.Default.Timeout = 5 * time.Second
	c.Default.Retries = 0
	c.Default.Logger = log
	err = c.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer c.Default.Conn.Close()

	oid := "1.2"

	var results []c.SnmpPDU

	err = c.Default.Walk(oid, func(pdu c.SnmpPDU) error {
		results = append(results, pdu)
		return nil
	})
	if err != nil {
		fmt.Printf("Walk Error: %v\n", err)
	}

	for _, v := range results {
		if v.Type == c.OctetString {
			fmt.Printf("Result for [%s]:  %+v\n", v.Name, string(v.Value.([]byte)))
			continue
		}
		fmt.Printf("Result for [%s]:  %+v\n", v.Name, v.Value)
	}

	// Output:
	// Result for [.1.2.1]:  eth0
	// Result for [.1.2.2]:  eth1
	// Result for [.1.2.3]:  eth2
	// Result for [.1.2.4]:  eth3
}

func ExampleServerClientBulkWalk() {
	var err error
	log = logrus.New()
	log.Level = logrus.DebugLevel

	s := &SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []c.SnmpPDU{
			{Name: ".1.1.1", Type: c.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: c.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: c.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: c.Integer, Value: int(54)},
			{Name: ".1.2.1", Type: c.OctetString, Value: "eth0"},
			{Name: ".1.2.2", Type: c.OctetString, Value: "eth1"},
			{Name: ".1.2.3", Type: c.OctetString, Value: "eth2"},
			{Name: ".1.2.4", Type: c.OctetString, Value: "eth3"},
		},
	}

	err = s.Start()
	if err != nil {
		log.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	c.Default.Version = c.Version2c
	c.Default.Community = "public"
	c.Default.Target = "127.0.0.1"
	c.Default.Port = 1161
	c.Default.Timeout = 5 * time.Second
	c.Default.Retries = 0
	c.Default.Logger = log
	err = c.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer c.Default.Conn.Close()

	oid := "1.1"

	var results []c.SnmpPDU

	err = c.Default.BulkWalk(oid, func(pdu c.SnmpPDU) error {
		results = append(results, pdu)
		return nil
	})
	if err != nil {
		fmt.Printf("Walk Error: %v\n", err)
	}

	for _, v := range results {
		fmt.Printf("Result for [%s]:  %+v\n", v.Name, v.Value)
	}

	// Output:
	// Result for [.1.1.1]:  51
	// Result for [.1.1.2]:  52
	// Result for [.1.1.3]:  53
	// Result for [.1.1.4]:  54
}
