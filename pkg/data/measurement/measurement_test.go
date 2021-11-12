package measurement

import (
	"bytes"
	"fmt"
	"sort"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/metric"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/mock"
)

func ProcessMeasurementFull(m *Measurement, varmap map[string]interface{}) error {
	err := m.Init()
	if err != nil {
		return fmt.Errorf("Can not initialize measurement %s", err)
	}

	m.InitBuildRuntime()

	nGets, nProcs, nErrs := m.GetData()
	m.ComputeOidConditionalMetrics()
	m.ComputeEvaluatedMetrics(varmap)

	m.Log.Infof("GETS: %d,NPROCS: %d ,NERRS %d", nGets, nProcs, nErrs)
	m.Log.Infof("GOT CURR INDEXED VALUES --> %+v ", m.CurIndexedLabels)
	return nil
}

func OrderMapByKey(m map[string]string) string {
	var buffer bytes.Buffer

	buffer.WriteString("{")

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		buffer.WriteString(" ")
		buffer.WriteString(k)
		buffer.WriteString(":")
		buffer.WriteString(m[k])
		if i < len(keys)-1 {
			buffer.WriteString(",")
		} else {
			buffer.WriteString(" ")
		}
	}

	buffer.WriteString("}")

	return buffer.String()
}

func GetOutputInfluxMetrics(m *Measurement) {
	m.Log.Infof("GOT MEAS --> %+v", m)

	metSent, metError, measSent, measError, ptarray := m.GetInfluxPoint(map[string]string{})

	m.Log.Infof("METRIC SENT[%d],METRIC ERROR[%d],MEAS SENT[%d], MEAS ERROR[%d]", metSent, metError, measSent, measError)

	for _, v := range ptarray {
		m.Log.Infof("GOT V %+v", v)
		fields, _ := v.Fields()
		tags := OrderMapByKey(v.Tags())

		for fname, fvalue := range fields {
			fmt.Printf("Measurement:%s Tags:%s Field:%s ValueType:%T  Value:%+v\n", v.Name(), tags, fname, fvalue, fvalue)
		}
	}
}

// 1.- SETUP LOGGER
// 2.- MOCK SERVER SETUP
// 3.- SNMP CLIENT SETUP
// 4.- METRICMAP SETUP
// 5.- MEASUREMENT CONFIG SETUP

func Example_Measurement_GetMode_Value() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.2.1", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.2.2", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.2.3", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.2.4", Type: gosnmp.OctetString, Value: "eth4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"metric_1": {
			ID:          "metric_1",
			FieldName:   "metric_1_name",
			Description: "",
			BaseOID:     ".1.1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"metric_2": {
			ID:          "metric_2",
			FieldName:   "metric_2_name",
			Description: "",
			BaseOID:     ".1.1.2",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:             "test_id",
		Name:           "test_name",
		GetMode:        "value",
		IndexOID:       "",
		TagOID:         "",
		IndexTag:       "",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "metric_1", Report: metric.AlwaysReport},
			{ID: "metric_2", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP
	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return

	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:test_name Tags:{} Field:metric_2_name ValueType:int64  Value:52
	// Measurement:test_name Tags:{} Field:metric_1_name ValueType:int64  Value:51
}

func TestSmallMaxOIDValue(t *testing.T) {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(1)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(2)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(3)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(4)},
			{Name: ".1.1.5", Type: gosnmp.Integer, Value: int(5)},
			{Name: ".1.1.6", Type: gosnmp.Integer, Value: int(6)},
			{Name: ".1.1.7", Type: gosnmp.Integer, Value: int(7)},
			{Name: ".1.1.8", Type: gosnmp.Integer, Value: int(8)},
			{Name: ".1.1.9", Type: gosnmp.Integer, Value: int(9)},
			{Name: ".1.1.10", Type: gosnmp.Integer, Value: int(10)},
			{Name: ".1.1.11", Type: gosnmp.Integer, Value: int(11)},
			{Name: ".1.1.12", Type: gosnmp.Integer, Value: int(12)},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
		MaxOids:     2,
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{
		"foo1=.1.1.1",
		"foo2=.1.1.2",
		"foo3=.1.1.3",
	})
	if err == nil {
		t.Fatal("Connect should fail because MaxOids is less than the number of OIDs requested while getting sys info")
	}
}

func Example_Measurement_GetMode_Indexed() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.2.1", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.2.2", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.2.3", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.2.4", Type: gosnmp.OctetString, Value: "eth4"},
			{Name: ".1.3.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4", Type: gosnmp.Integer, Value: int(24)},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:             "interfaces_data",
		Name:           "interfaces_data",
		GetMode:        "indexed",
		IndexOID:       ".1.2",
		TagOID:         "",
		IndexTag:       "portName",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:output ValueType:int64  Value:24
}

func Example_Measurement_GetMode_Indexed_Indirect() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4", Type: gosnmp.Integer, Value: int(24)},
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.2.3", Type: gosnmp.Integer, Value: int(92)},
			{Name: ".1.2.4", Type: gosnmp.Integer, Value: int(93)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.4.92", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.4.93", Type: gosnmp.OctetString, Value: "eth4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP

	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:             "interfaces_data",
		Name:           "interfaces_data",
		GetMode:        "indexed_it",
		IndexOID:       ".1.2",
		TagOID:         ".1.4",
		IndexTag:       "portName",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:output ValueType:int64  Value:24
}

func Example_Measurement_GetMode_Indexed_Multi_Indirect() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	lev, _ := logrus.ParseLevel("debug")

	l.SetLevel(lev)

	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.1", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.1", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.1", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.1", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.1", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.1", Type: gosnmp.Integer, Value: int(24)},
			// Indirect Table
			{Name: ".1.2.1.1", Type: gosnmp.Integer, Value: int(2)},
			{Name: ".1.2.2.1", Type: gosnmp.Integer, Value: int(2)},
			{Name: ".1.2.3.1", Type: gosnmp.Integer, Value: int(2)},
			{Name: ".1.2.4.1", Type: gosnmp.Integer, Value: int(2)},

			{Name: ".1.10.1.2", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.10.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.10.3.2", Type: gosnmp.Integer, Value: int(92)},
			{Name: ".1.10.4.2", Type: gosnmp.Integer, Value: int(93)},

			{Name: ".1.4.90", Type: gosnmp.Integer, Value: int(94)},
			{Name: ".1.4.91", Type: gosnmp.Integer, Value: int(95)},
			{Name: ".1.4.92", Type: gosnmp.Integer, Value: int(96)},
			{Name: ".1.4.93", Type: gosnmp.Integer, Value: int(97)},

			{Name: ".1.5.94", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.5.95", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.5.96", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.5.97", Type: gosnmp.OctetString, Value: "eth4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP

	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:       "interfaces_data",
		Name:     "interfaces_data",
		GetMode:  "indexed_mit",
		IndexOID: ".1.2",
		MultiTagOID: []config.MultipleTagOID{
			{
				TagOID:      ".1.10",
				IndexFormat: "${IDX1|DOT[0:0]|STRING}.$VAL1",
			},
			{
				TagOID:      ".1.4",
				IndexFormat: "",
			},
			{
				TagOID:      ".1.5",
				IndexFormat: "",
			},
		},
		IndexTag:       "portName",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP
	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:output ValueType:int64  Value:24
}

func Example_Measurement_GetMode_Indexed_MultiIndex_() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4", Type: gosnmp.Integer, Value: int(24)},
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.2.3", Type: gosnmp.Integer, Value: int(92)},
			{Name: ".1.2.4", Type: gosnmp.Integer, Value: int(93)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.4.92", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.4.93", Type: gosnmp.OctetString, Value: "eth4"},
			{Name: ".1.6.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.4", Type: gosnmp.OctetString, Value: "port4"},
			{Name: ".1.7.1", Type: gosnmp.OctetString, Value: "myPort1"},
			{Name: ".1.7.2", Type: gosnmp.OctetString, Value: "myPort2"},
			{Name: ".1.7.3", Type: gosnmp.OctetString, Value: "myPort3"},
			{Name: ".1.7.4", Type: gosnmp.OctetString, Value: "myPort4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     "IDX{0}",
			},
			{
				Label:          "ifAlias",
				IndexOID:       ".1.7",
				IndexTag:       "portAlias",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     "IDX{1};DOT[0:0];FILL(Not defined)",
			},
		},
		MultiIndexResult: "IDX{2}",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portAlias:myPort2, portDesc:Not defined, portName:Not defined } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portAlias:myPort2, portDesc:Not defined, portName:Not defined } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portAlias:myPort3, portDesc:Not defined, portName:Not defined } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portAlias:myPort3, portDesc:Not defined, portName:Not defined } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portAlias:myPort4, portDesc:port4, portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portAlias:myPort4, portDesc:port4, portName:eth4 } Field:output ValueType:int64  Value:24
	// Measurement:interfaces_data Tags:{ portAlias:myPort1, portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portAlias:myPort1, portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
}

func Example_Measurement_GetMode_Indexed_MultiIndexDIM2() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.4", Type: gosnmp.Integer, Value: int(24)},
			// Measurement tables
			// Indirect Table
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.2.3", Type: gosnmp.Integer, Value: int(92)},
			{Name: ".1.2.4", Type: gosnmp.Integer, Value: int(93)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.4.92", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.4.93", Type: gosnmp.OctetString, Value: "eth4"},
			// Direct table
			{Name: ".1.6.1.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.2.2", Type: gosnmp.OctetString, Value: "port2"},
			{Name: ".1.6.3.3", Type: gosnmp.OctetString, Value: "port3"},
			{Name: ".1.6.4.4", Type: gosnmp.OctetString, Value: "port4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     "IDX{0};DOT[0:0];SKIP",
			},
		},
		MultiIndexResult: "IDX{1}",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portDesc:port3, portName:eth3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portDesc:port3, portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portDesc:port4, portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portDesc:port4, portName:eth4 } Field:output ValueType:int64  Value:24
}

func Example_Measurement_GetMode_Indexed_MultiIndex_DIM2_SKIP() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.4", Type: gosnmp.Integer, Value: int(24)},
			// Measurement tables
			// Indirect Table
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			// Direct table
			{Name: ".1.6.1.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.2.2", Type: gosnmp.OctetString, Value: "port2"},
			{Name: ".1.6.3.3", Type: gosnmp.OctetString, Value: "port3"},
			{Name: ".1.6.4.4", Type: gosnmp.OctetString, Value: "port4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];SKIP`,
			},
		},
		MultiIndexResult: "IDX{1}",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:output ValueType:int64  Value:22
}

func Example_Measurement_GetMode_Indexed_MultiIndex_DIM2_FILLNONE() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.4", Type: gosnmp.Integer, Value: int(24)},
			// Measurement tables
			// Indirect Table
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			// Direct table
			{Name: ".1.6.1.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.2.2", Type: gosnmp.OctetString, Value: "port2"},
			{Name: ".1.6.3.3", Type: gosnmp.OctetString, Value: "port3"},
			{Name: ".1.6.4.4", Type: gosnmp.OctetString, Value: "port4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];FILL()`,
			},
		},
		MultiIndexResult: "IDX{1}",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portDesc:port3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portDesc:port3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portDesc:port4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portDesc:port4 } Field:output ValueType:int64  Value:24
}

func Example_Measurement_GetMode_Indexed_MultiIndex_DIM2_FILLSTRING() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.4", Type: gosnmp.Integer, Value: int(24)},
			// Measurement tables
			// Indirect Table
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			// Direct table
			{Name: ".1.6.1.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.2.2", Type: gosnmp.OctetString, Value: "port2"},
			{Name: ".1.6.3.3", Type: gosnmp.OctetString, Value: "port3"},
			{Name: ".1.6.4.4", Type: gosnmp.OctetString, Value: "port4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];FILL(NotRelated)`,
			},
		},
		MultiIndexResult: "IDX{1}",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portDesc:port3, portName:NotRelated } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portDesc:port3, portName:NotRelated } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portDesc:port4, portName:NotRelated } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portDesc:port4, portName:NotRelated } Field:output ValueType:int64  Value:24
}

func Example_Measurement_GetMode_Indexed_MultiIndex_DIM2_SKIP_CUSTOMRESULT() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.2.1", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.3.1", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.4.1", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.2.1", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.3.1", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.4.1", Type: gosnmp.Integer, Value: int(24)},
			// Measurement tables
			// Indirect Table
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			// Direct table
			{Name: ".1.6.1.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.2.2", Type: gosnmp.OctetString, Value: "port2"},
			{Name: ".1.6.3.3", Type: gosnmp.OctetString, Value: "port3"},
			{Name: ".1.6.4.4", Type: gosnmp.OctetString, Value: "port4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];SKIP`,
			},
		},
		MultiIndexResult: "IDX{1}.1",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portDesc:port2, portName:eth2 } Field:output ValueType:int64  Value:22
}

func Example_Measurement_GetMode_Indexed_MultiIndex_DIM2_SKIP_CUSTOMRESULT_COMPLEX() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1.1.5", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.2.1.6", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.3.1.5", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.4.1.6", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1.1.5", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.2.1.6", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.3.1.5", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.4.1.6", Type: gosnmp.Integer, Value: int(24)},
			// Measurement tables
			// Indirect Table
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			// Direct table
			{Name: ".1.6.1.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.2.2", Type: gosnmp.OctetString, Value: "port2"},
			{Name: ".1.6.3.3", Type: gosnmp.OctetString, Value: "port3"},
			{Name: ".1.6.4.4", Type: gosnmp.OctetString, Value: "port4"},
			// Direct table
			{Name: ".1.7.5", Type: gosnmp.OctetString, Value: "5"},
			{Name: ".1.7.6", Type: gosnmp.OctetString, Value: "6"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];SKIP`,
			},
			{
				Label:          "ifType",
				IndexOID:       ".1.7",
				IndexTag:       "ifType",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     ``,
			},
		},
		MultiIndexResult: "IDX{1}.1.IDX{2}",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ ifType:5, portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ ifType:5, portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ ifType:6, portDesc:port2, portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ ifType:6, portDesc:port2, portName:eth2 } Field:output ValueType:int64  Value:22
}

func Example_Measurement_GetMode_Indexed_MultiIndex_DIM2_FILLNONE_CUSTOMRESULT_COMPLEX() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1.1.1.5", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2.2.1.6", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3.3.1.5", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4.4.1.6", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1.1.1.5", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2.2.1.6", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3.3.1.5", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4.4.1.6", Type: gosnmp.Integer, Value: int(24)},
			// Measurement tables
			// Indirect Table
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			// Direct table
			{Name: ".1.6.1.1", Type: gosnmp.OctetString, Value: "port1"},
			{Name: ".1.6.2.2", Type: gosnmp.OctetString, Value: "port2"},
			{Name: ".1.6.3.3", Type: gosnmp.OctetString, Value: "port3"},
			{Name: ".1.6.4.4", Type: gosnmp.OctetString, Value: "port4"},
			// Direct table
			{Name: ".1.7.5", Type: gosnmp.OctetString, Value: "5"},
			{Name: ".1.7.6", Type: gosnmp.OctetString, Value: "6"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "ifName",
				IndexOID:       ".1.2",
				TagOID:         ".1.4",
				IndexTag:       "portName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "ifDesc",
				IndexOID:       ".1.6",
				IndexTag:       "portDesc",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];FILL()`,
			},
			{
				Label:          "ifType",
				IndexOID:       ".1.7",
				IndexTag:       "ifType",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     ``,
			},
		},
		MultiIndexResult: "IDX{1}.1.IDX{2}",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ ifType:5, portDesc:port1, portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ ifType:5, portDesc:port1, portName:eth1 } Field:output ValueType:int64  Value:21
	// Measurement:interfaces_data Tags:{ ifType:6, portDesc:port2, portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ ifType:6, portDesc:port2, portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ ifType:5, portDesc:port3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ ifType:5, portDesc:port3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ ifType:6, portDesc:port4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ ifType:6, portDesc:port4 } Field:output ValueType:int64  Value:24
}

func Example_Measurement_GetMode_Indexed_MultiIndex_QOS_CMSTATS() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			// cbQosCMPrePolicyPkt64
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.3.1043.1045", Type: gosnmp.Counter64, Value: uint64(8)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.3.1043.1051", Type: gosnmp.Counter64, Value: uint64(1131)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.3.1099.1101", Type: gosnmp.Counter64, Value: uint64(281)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.3.1099.1107", Type: gosnmp.Counter64, Value: uint64(7016)},
			// cbQosCMPrePolicyByte64
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.6.1043.1045", Type: gosnmp.Counter64, Value: uint64(784)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.6.1043.1051", Type: gosnmp.Counter64, Value: uint64(114630)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.6.1099.1101", Type: gosnmp.Counter64, Value: uint64(69858)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.15.1.1.6.1099.1107", Type: gosnmp.Counter64, Value: uint64(658800)},
			//	cbQoSIfIndex
			//	- cbQoSIfIndex
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.4.1043", Type: gosnmp.Integer, Value: int(1)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.4.1099", Type: gosnmp.Integer, Value: int(1)},
			{Name: ".1.3.6.1.2.1.31.1.1.1.1.1", Type: gosnmp.OctetString, Value: "FastEthernet0/0"},
			//	cbQosPolicyDirection
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.3.1043", Type: gosnmp.Integer, Value: int(2)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.3.1099", Type: gosnmp.Integer, Value: int(1)}, // FastEthernet0/0 | input

			// cbQosPolicyMapName
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1043.1045", Type: gosnmp.Integer, Value: int(1043)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1043.1051", Type: gosnmp.Integer, Value: int(1043)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1099.1101", Type: gosnmp.Integer, Value: int(1099)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1099.1107", Type: gosnmp.Integer, Value: int(1099)},

			// cbQosParentObjectsIndex
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1043", Type: gosnmp.Integer, Value: int(1035)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1099", Type: gosnmp.Integer, Value: int(1063)},

			// - cbQosPolicyMapName
			{Name: ".1.3.6.1.4.1.9.9.166.1.6.1.1.1.1035", Type: gosnmp.OctetString, Value: "LAN_Out"},
			{Name: ".1.3.6.1.4.1.9.9.166.1.6.1.1.1.1063", Type: gosnmp.OctetString, Value: "CPP"},

			//	cbQosCMName
			//	- cbQosConfigIndex
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1045", Type: gosnmp.Integer, Value: int(1029)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1051", Type: gosnmp.Integer, Value: int(1025)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1101", Type: gosnmp.Integer, Value: int(1057)}, // FastEthernet0/0 | input
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1107", Type: gosnmp.Integer, Value: int(1025)}, // FastEthernet0/0 | input
			//- cbQosCMName
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.1.1025", Type: gosnmp.OctetString, Value: "class-default"}, // FastEthernet0/0 | output | class-default || //FastEthernet0/0 | input | class-default
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.1.1029", Type: gosnmp.OctetString, Value: "ICMP"},          // FastEthernet0/0 | input | ICMP
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.1.1057", Type: gosnmp.OctetString, Value: "NonLocal"},      // FastEthernet0/0 | input | NonLocal
			//- cbQosCMInfo
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.3.1025", Type: gosnmp.Integer, Value: int(3)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.3.1029", Type: gosnmp.Integer, Value: int(2)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.3.1057", Type: gosnmp.Integer, Value: int(3)},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"cisco_cbQosCMPrePolicyPkt64": {
			ID:          "cisco_cbQosCMPrePolicyPkt64",
			FieldName:   "cbQosCMPrePolicyPkt64",
			Description: "",
			BaseOID:     ".1.3.6.1.4.1.9.9.166.1.15.1.1.3",
			DataSrcType: "Counter64",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"cisco_cbQosCMPrePolicyByte64": {
			ID:          "cisco_cbQosCMPrePolicyByte64",
			FieldName:   "cbQosCMPrePolicyByte64",
			Description: "",
			BaseOID:     ".1.3.6.1.4.1.9.9.166.1.15.1.1.6",
			DataSrcType: "Counter64",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "cbQosIfIndex",
				Description:    "Index to retrieve interface name from cbQosIfIndex",
				IndexOID:       ".1.3.6.1.4.1.9.9.166.1.1.1.1.4",
				TagOID:         ".1.3.6.1.2.1.31.1.1.1.1",
				IndexTag:       "ifName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "cbQosPolicyDirection",
				Description:    "Index to retrieve interface Policy Direction",
				IndexOID:       ".1.3.6.1.4.1.9.9.166.1.1.1.1.3",
				IndexTag:       "policyDirection",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];SKIP`,
			},
			{
				Label:          "cbQosCMName",
				Description:    "Index to retrieve CMName",
				IndexOID:       ".1.3.6.1.4.1.9.9.166.1.5.1.1.2",
				TagOID:         ".1.3.6.1.4.1.9.9.166.1.7.1.1.1",
				IndexTag:       "cmName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     `IDX{1};DOT[0:0];SKIP`,
			},
			{
				Label:       "cbQosPolicyMapName",
				Description: "Index to retrieve cbQosPolicyMapName",
				IndexOID:    ".1.3.6.1.4.1.9.9.166.1.5.1.1.4",
				MultiTagOID: []config.MultipleTagOID{
					{TagOID: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2", IndexFormat: "${IDX1|DOT[0:0]|STRING}.$VAL1"},
					{TagOID: ".1.3.6.1.4.1.9.9.166.1.6.1.1.1", IndexFormat: ""},
				},
				IndexTag:       "policyMapName",
				IndexTagFormat: "",
				GetMode:        "indexed_mit",
				Dependency:     `IDX{2};DOT[0:1];SKIP`,
			},
			{
				Label:          "cbQosCMInfo",
				Description:    "Index to retrieve cbQosCMInfo",
				IndexOID:       ".1.3.6.1.4.1.9.9.166.1.5.1.1.2",
				TagOID:         ".1.3.6.1.4.1.9.9.166.1.7.1.1.3",
				IndexTag:       "cmInfo",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     `IDX{3};DOT[0:1];SKIP`,
			},
		},
		MultiIndexResult: "IDX{4}",
		Fields: []config.MeasurementFieldReport{
			{ID: "cisco_cbQosCMPrePolicyPkt64", Report: metric.AlwaysReport},
			{ID: "cisco_cbQosCMPrePolicyByte64", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ cmInfo:3, cmName:NonLocal, ifName:FastEthernet0/0, policyDirection:1, policyMapName:CPP } Field:cbQosCMPrePolicyByte64 ValueType:int64  Value:69858
	// Measurement:interfaces_data Tags:{ cmInfo:3, cmName:NonLocal, ifName:FastEthernet0/0, policyDirection:1, policyMapName:CPP } Field:cbQosCMPrePolicyPkt64 ValueType:int64  Value:281
	// Measurement:interfaces_data Tags:{ cmInfo:3, cmName:class-default, ifName:FastEthernet0/0, policyDirection:1, policyMapName:CPP } Field:cbQosCMPrePolicyByte64 ValueType:int64  Value:658800
	// Measurement:interfaces_data Tags:{ cmInfo:3, cmName:class-default, ifName:FastEthernet0/0, policyDirection:1, policyMapName:CPP } Field:cbQosCMPrePolicyPkt64 ValueType:int64  Value:7016
	// Measurement:interfaces_data Tags:{ cmInfo:2, cmName:ICMP, ifName:FastEthernet0/0, policyDirection:2, policyMapName:LAN_Out } Field:cbQosCMPrePolicyByte64 ValueType:int64  Value:784
	// Measurement:interfaces_data Tags:{ cmInfo:2, cmName:ICMP, ifName:FastEthernet0/0, policyDirection:2, policyMapName:LAN_Out } Field:cbQosCMPrePolicyPkt64 ValueType:int64  Value:8
	// Measurement:interfaces_data Tags:{ cmInfo:3, cmName:class-default, ifName:FastEthernet0/0, policyDirection:2, policyMapName:LAN_Out } Field:cbQosCMPrePolicyByte64 ValueType:int64  Value:114630
	// Measurement:interfaces_data Tags:{ cmInfo:3, cmName:class-default, ifName:FastEthernet0/0, policyDirection:2, policyMapName:LAN_Out } Field:cbQosCMPrePolicyPkt64 ValueType:int64  Value:1131
}

func Example_Measurement_GetMode_Indexed_MultiIndex_QOS_MATCH_NAME() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			// cbQosMatchPrePolicyPkt64
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.3.1043.1047", Type: gosnmp.Counter64, Value: uint64(8)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.3.1043.1053", Type: gosnmp.Counter64, Value: uint64(1131)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.3.1099.1103", Type: gosnmp.Counter64, Value: uint64(281)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.3.1099.1109", Type: gosnmp.Counter64, Value: uint64(7016)},
			// cbQosCMPrePolicyByte64
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.6.1043.1047", Type: gosnmp.Counter64, Value: uint64(784)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.6.1043.1053", Type: gosnmp.Counter64, Value: uint64(114630)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.6.1099.1103", Type: gosnmp.Counter64, Value: uint64(69858)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.16.1.1.6.1099.1109", Type: gosnmp.Counter64, Value: uint64(658800)},
			//	cbQoSIfIndex
			//	- cbQoSIfIndex
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.4.1043", Type: gosnmp.Integer, Value: int(1)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.4.1099", Type: gosnmp.Integer, Value: int(1)},
			{Name: ".1.3.6.1.2.1.31.1.1.1.1.1", Type: gosnmp.OctetString, Value: "FastEthernet0/0"},
			//	cbQosPolicyDirection
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.3.1043", Type: gosnmp.Integer, Value: int(2)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.1.1.1.3.1099", Type: gosnmp.Integer, Value: int(1)}, // FastEthernet0/0 | input

			// cbQosPolicyMapName
			// - cbQosParentObjectsIndex
			// -- CM -> Policy
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1043.1045", Type: gosnmp.Integer, Value: int(1043)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1043.1051", Type: gosnmp.Integer, Value: int(1043)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1099.1101", Type: gosnmp.Integer, Value: int(1099)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1099.1107", Type: gosnmp.Integer, Value: int(1099)},
			// -- Match -> CM
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1043.1047", Type: gosnmp.Integer, Value: int(1045)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1043.1053", Type: gosnmp.Integer, Value: int(1051)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1099.1103", Type: gosnmp.Integer, Value: int(1101)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4.1099.1109", Type: gosnmp.Integer, Value: int(1107)},

			// cbQosParentObjectsIndex
			// - cbQosConfigIndex
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1043", Type: gosnmp.Integer, Value: int(1035)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1099", Type: gosnmp.Integer, Value: int(1063)},

			// - cbQosPolicyMapName
			{Name: ".1.3.6.1.4.1.9.9.166.1.6.1.1.1.1035", Type: gosnmp.OctetString, Value: "LAN_Out"},
			{Name: ".1.3.6.1.4.1.9.9.166.1.6.1.1.1.1063", Type: gosnmp.OctetString, Value: "CPP"},

			//	cbQosCMName
			//	- cbQosConfigIndex
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1045", Type: gosnmp.Integer, Value: int(1029)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1051", Type: gosnmp.Integer, Value: int(1025)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1101", Type: gosnmp.Integer, Value: int(1057)}, // FastEthernet0/0 | input
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1107", Type: gosnmp.Integer, Value: int(1025)}, // FastEthernet0/0 | input
			//- cbQosCMName
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.1.1025", Type: gosnmp.OctetString, Value: "class-default"}, // FastEthernet0/0 | output | class-default || //FastEthernet0/0 | input | class-default
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.1.1029", Type: gosnmp.OctetString, Value: "ICMP"},          // FastEthernet0/0 | input | ICMP
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.1.1057", Type: gosnmp.OctetString, Value: "NonLocal"},      // FastEthernet0/0 | input | NonLocal
			//- cbQosCMInfo
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.3.1025", Type: gosnmp.Integer, Value: int(3)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.3.1029", Type: gosnmp.Integer, Value: int(2)},
			{Name: ".1.3.6.1.4.1.9.9.166.1.7.1.1.3.1057", Type: gosnmp.Integer, Value: int(3)},

			// cbQosMatchStmtName
			// - cbQosConfigIndex
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1047", Type: gosnmp.Integer, Value: int(1053)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1043.1053", Type: gosnmp.Integer, Value: int(1053)}, // FastEthernet0/0 | output
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1103", Type: gosnmp.Integer, Value: int(1054)}, // FastEthernet0/0 | input
			{Name: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2.1099.1109", Type: gosnmp.Integer, Value: int(1054)}, // FastEthernet0/0 | input

			// - cbQosMatchStmtName
			{Name: ".1.3.6.1.4.1.9.9.166.1.8.1.1.1.1053", Type: gosnmp.OctetString, Value: "CLASS_BACKUP"},   // FastEthernet0/0 | output | class-default || //FastEthernet0/0 | input | class-default
			{Name: ".1.3.6.1.4.1.9.9.166.1.8.1.1.1.1054", Type: gosnmp.OctetString, Value: "CLASS_INTERNET"}, // FastEthernet0/0 | input | ICMP
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"cisco_cbQosMatchPrePolicyPkt64": {
			ID: "cisco_cbQosMatchPrePolicyPkt64",
			FieldName: "cbQosMatchPrePolicyPkt64	",
			Description: "",
			BaseOID:     ".1.3.6.1.4.1.9.9.166.1.16.1.1.3",
			DataSrcType: "Counter64",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"cisco_cbQosMatchPrePolicyByte64": {
			ID:          "cisco_cbQosMatchPrePolicyByte64",
			FieldName:   "cbQosMatchPrePolicyByte64",
			Description: "",
			BaseOID:     ".1.3.6.1.4.1.9.9.166.1.16.1.1.6",
			DataSrcType: "Counter64",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:      "interfaces_data",
		Name:    "interfaces_data",
		GetMode: "indexed_multiple",
		MultiIndexCfg: []config.MultiIndexCfg{
			{
				Label:          "cbQosIfIndex",
				Description:    "Index to retrieve interface name from cbQosIfIndex",
				IndexOID:       ".1.3.6.1.4.1.9.9.166.1.1.1.1.4",
				TagOID:         ".1.3.6.1.2.1.31.1.1.1.1",
				IndexTag:       "ifName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     "",
			},
			{
				Label:          "cbQosPolicyDirection",
				Description:    "Index to retrieve interface Policy Direction",
				IndexOID:       ".1.3.6.1.4.1.9.9.166.1.1.1.1.3",
				IndexTag:       "policyDirection",
				IndexTagFormat: "",
				GetMode:        "indexed",
				Dependency:     `IDX{0};DOT[0:0];SKIP`,
			},
			{
				Label:       "cbQosCMName",
				Description: "Index to retrieve CMName",
				IndexOID:    ".1.3.6.1.4.1.9.9.166.1.5.1.1.4",
				MultiTagOID: []config.MultipleTagOID{
					{TagOID: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2", IndexFormat: "${IDX1|DOT[0:0]|STRING}.$VAL1"},
					{TagOID: ".1.3.6.1.4.1.9.9.166.1.7.1.1.1", IndexFormat: ""},
				},
				IndexTag:       "cmName",
				IndexTagFormat: "",
				GetMode:        "indexed_mit",
				Dependency:     `IDX{1};DOT[0:0];SKIP`,
			},
			{
				Label:       "cbQosPolicyMapName",
				Description: "Index to retrieve cbQosPolicyMapName",
				IndexOID:    ".1.3.6.1.4.1.9.9.166.1.5.1.1.4",
				MultiTagOID: []config.MultipleTagOID{
					{TagOID: ".1.3.6.1.4.1.9.9.166.1.5.1.1.4", IndexFormat: "${IDX1|DOT[0:0]|STRING}.$VAL1"},
					{TagOID: ".1.3.6.1.4.1.9.9.166.1.5.1.1.2", IndexFormat: "${IDX1|DOT[0:0]|STRING}.$VAL1"},
					{TagOID: ".1.3.6.1.4.1.9.9.166.1.6.1.1.1", IndexFormat: ""},
				},
				IndexTag:       "policyMapName",
				IndexTagFormat: "",
				GetMode:        "indexed_mit",
				Dependency:     `IDX{2};DOT[0:1];SKIP`,
			},
			{
				Label:          "cbQosMatchStmtName",
				Description:    "Index to retrieve cbQosMatchStmtName",
				IndexOID:       ".1.3.6.1.4.1.9.9.166.1.5.1.1.2",
				TagOID:         ".1.3.6.1.4.1.9.9.166.1.8.1.1.1",
				IndexTag:       "matchStmtName",
				IndexTagFormat: "",
				GetMode:        "indexed_it",
				Dependency:     `IDX{3};DOT[0:1];SKIP`,
			},
		},
		MultiIndexResult: "IDX{4}",
		Fields: []config.MeasurementFieldReport{
			{ID: "cisco_cbQosMatchPrePolicyPkt64", Report: metric.AlwaysReport},
			{ID: "cisco_cbQosMatchPrePolicyByte64", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ cmName:ICMP, ifName:FastEthernet0/0, matchStmtName:CLASS_BACKUP, policyDirection:2, policyMapName:LAN_Out } Field:cbQosMatchPrePolicyPkt64	 ValueType:int64  Value:8
	// Measurement:interfaces_data Tags:{ cmName:ICMP, ifName:FastEthernet0/0, matchStmtName:CLASS_BACKUP, policyDirection:2, policyMapName:LAN_Out } Field:cbQosMatchPrePolicyByte64 ValueType:int64  Value:784
	// Measurement:interfaces_data Tags:{ cmName:NonLocal, ifName:FastEthernet0/0, matchStmtName:CLASS_INTERNET, policyDirection:1, policyMapName:CPP } Field:cbQosMatchPrePolicyByte64 ValueType:int64  Value:69858
	// Measurement:interfaces_data Tags:{ cmName:NonLocal, ifName:FastEthernet0/0, matchStmtName:CLASS_INTERNET, policyDirection:1, policyMapName:CPP } Field:cbQosMatchPrePolicyPkt64	 ValueType:int64  Value:281
	// Measurement:interfaces_data Tags:{ cmName:class-default, ifName:FastEthernet0/0, matchStmtName:CLASS_INTERNET, policyDirection:1, policyMapName:CPP } Field:cbQosMatchPrePolicyByte64 ValueType:int64  Value:658800
	// Measurement:interfaces_data Tags:{ cmName:class-default, ifName:FastEthernet0/0, matchStmtName:CLASS_INTERNET, policyDirection:1, policyMapName:CPP } Field:cbQosMatchPrePolicyPkt64	 ValueType:int64  Value:7016
	// Measurement:interfaces_data Tags:{ cmName:class-default, ifName:FastEthernet0/0, matchStmtName:CLASS_BACKUP, policyDirection:2, policyMapName:LAN_Out } Field:cbQosMatchPrePolicyByte64 ValueType:int64  Value:114630
	// Measurement:interfaces_data Tags:{ cmName:class-default, ifName:FastEthernet0/0, matchStmtName:CLASS_BACKUP, policyDirection:2, policyMapName:LAN_Out } Field:cbQosMatchPrePolicyPkt64	 ValueType:int64  Value:1131
}

func Example_Measurement_value_STRINGEVAL() {
	// 1.- LOGGER SETUP

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.2.1", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.2.2", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.2.3", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.2.4", Type: gosnmp.OctetString, Value: "eth4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"metric_1": {
			ID:          "metric_1",
			FieldName:   "metric_1_name",
			Description: "",
			BaseOID:     ".1.1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"metric_2": {
			ID:          "metric_2",
			FieldName:   "metric_2_name",
			Description: "",
			BaseOID:     ".1.1.2",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"metric_3": {
			ID:          "metric_3",
			FieldName:   "metric_3_oper",
			Description: "",
			BaseOID:     "",
			DataSrcType: "STRINGEVAL",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "metric_2_name - metric_1_name + var2", // 52 - 51 + 30 = 31
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	cfg := &config.MeasurementCfg{
		ID:             "test_id",
		Name:           "test_name",
		GetMode:        "value",
		IndexOID:       "",
		TagOID:         "",
		IndexTag:       "",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "metric_1", Report: metric.NeverReport},
			{ID: "metric_2", Report: metric.AlwaysReport},
			{ID: "metric_3", Report: metric.AlwaysReport},
		},
	}

	vars := map[string]interface{}{
		"var1": 25,
		"var2": 30,
	}

	err = cfg.Init(&metrics, vars)
	if err != nil {
		l.Errorf("Can not create measurement %s", err)
		return
	}

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return

	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:test_name Tags:{} Field:metric_2_name ValueType:int64  Value:52
	// Measurement:test_name Tags:{} Field:metric_3_oper ValueType:int64  Value:31
}

func Example_Measurement_Indexed_STRINGEVAL() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.2.1", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.2.2", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.2.3", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.2.4", Type: gosnmp.OctetString, Value: "eth4"},
			{Name: ".1.3.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4", Type: gosnmp.Integer, Value: int(24)},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_total": {
			ID:          "value_total",
			FieldName:   "total",
			Description: "",
			BaseOID:     "",
			DataSrcType: "STRINGEVAL",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "input + output",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:             "interfaces_data",
		Name:           "interfaces_data",
		GetMode:        "indexed",
		IndexOID:       ".1.2",
		TagOID:         "",
		IndexTag:       "portName",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
			{ID: "value_total", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:total ValueType:int64  Value:74
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:total ValueType:int64  Value:76
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:output ValueType:int64  Value:24
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:total ValueType:int64  Value:78
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:total ValueType:int64  Value:72
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:output ValueType:int64  Value:21
}

func Example_Measurement_Indexed_Indirect_STRINGEVAL() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4", Type: gosnmp.Integer, Value: int(24)},
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.2.3", Type: gosnmp.Integer, Value: int(92)},
			{Name: ".1.2.4", Type: gosnmp.Integer, Value: int(93)},
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.4.92", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.4.93", Type: gosnmp.OctetString, Value: "eth4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_total": {
			ID:          "value_total",
			FieldName:   "total",
			Description: "",
			BaseOID:     "",
			DataSrcType: "STRINGEVAL",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "input + output",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:             "interfaces_data",
		Name:           "interfaces_data",
		GetMode:        "indexed_it",
		IndexOID:       ".1.2",
		TagOID:         ".1.4",
		IndexTag:       "portName",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
			{ID: "value_total", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:total ValueType:int64  Value:74
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:total ValueType:int64  Value:76
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:output ValueType:int64  Value:24
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:total ValueType:int64  Value:78
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:total ValueType:int64  Value:72
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:output ValueType:int64  Value:21
}

func Example_Measurement_Indexed_Multi_Indirect_STRINGEVAL() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			// Metrics
			{Name: ".1.1.1", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.1.2", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.1.3", Type: gosnmp.Integer, Value: int(53)},
			{Name: ".1.1.4", Type: gosnmp.Integer, Value: int(54)},
			{Name: ".1.3.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4", Type: gosnmp.Integer, Value: int(24)},

			// Tables
			{Name: ".1.2.1", Type: gosnmp.Integer, Value: int(50)},
			{Name: ".1.2.2", Type: gosnmp.Integer, Value: int(51)},
			{Name: ".1.2.3", Type: gosnmp.Integer, Value: int(52)},
			{Name: ".1.2.4", Type: gosnmp.Integer, Value: int(53)},

			// Tables
			{Name: ".1.5.50", Type: gosnmp.Integer, Value: int(90)},
			{Name: ".1.5.51", Type: gosnmp.Integer, Value: int(91)},
			{Name: ".1.5.52", Type: gosnmp.Integer, Value: int(92)},
			{Name: ".1.5.53", Type: gosnmp.Integer, Value: int(93)},

			// Tables
			{Name: ".1.4.90", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.4.91", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.4.92", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.4.93", Type: gosnmp.OctetString, Value: "eth4"},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "input",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
		"value_total": {
			ID:          "value_total",
			FieldName:   "total",
			Description: "",
			BaseOID:     "",
			DataSrcType: "STRINGEVAL",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "input + output",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:       "interfaces_data",
		Name:     "interfaces_data",
		GetMode:  "indexed_mit",
		IndexOID: ".1.2",
		MultiTagOID: []config.MultipleTagOID{
			{TagOID: ".1.5", IndexFormat: ""},
			{TagOID: ".1.4", IndexFormat: ""},
		},
		IndexTag:       "portName",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
			{ID: "value_total", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:input ValueType:int64  Value:52
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ portName:eth2 } Field:total ValueType:int64  Value:74
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:input ValueType:int64  Value:53
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ portName:eth3 } Field:total ValueType:int64  Value:76
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:input ValueType:int64  Value:54
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:output ValueType:int64  Value:24
	// Measurement:interfaces_data Tags:{ portName:eth4 } Field:total ValueType:int64  Value:78
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:total ValueType:int64  Value:72
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:input ValueType:int64  Value:51
	// Measurement:interfaces_data Tags:{ portName:eth1 } Field:output ValueType:int64  Value:21
}

func Example_Measurement_GetMode_Indexed_MULTISTRINGPARSER_SKIPFIELD() {
	// 1.- SETUP LOGGER

	l := logrus.New()
	// l.Level = logrus.DebugLevel

	mock.SetLogger(l)
	config.SetLogger(l)

	// 2.- MOCK SERVER SETUP

	s := &mock.SnmpServer{
		Listen: "127.0.0.1:1161",
		Want: []gosnmp.SnmpPDU{
			{Name: ".1.1.1", Type: gosnmp.OctetString, Value: "value1;51"},
			{Name: ".1.1.2", Type: gosnmp.OctetString, Value: "value2;52"},
			{Name: ".1.1.3", Type: gosnmp.OctetString, Value: "value3;53"},
			{Name: ".1.1.4", Type: gosnmp.OctetString, Value: "value4;"}, // input with tag value4 will be ommited
			{Name: ".1.2.1", Type: gosnmp.OctetString, Value: "eth1"},
			{Name: ".1.2.2", Type: gosnmp.OctetString, Value: "eth2"},
			{Name: ".1.2.3", Type: gosnmp.OctetString, Value: "eth3"},
			{Name: ".1.2.4", Type: gosnmp.OctetString, Value: "eth4"},
			{Name: ".1.3.1", Type: gosnmp.Integer, Value: int(21)},
			{Name: ".1.3.2", Type: gosnmp.Integer, Value: int(22)},
			{Name: ".1.3.3", Type: gosnmp.Integer, Value: int(23)},
			{Name: ".1.3.4", Type: gosnmp.Integer, Value: int(24)},
		},
	}

	err := s.Start()
	if err != nil {
		l.Errorf("error on start snmp mock server: %s", err)
		return
	}
	defer s.Stop()

	// 3.- SNMP CLIENT SETUP
	connectionParams := snmp.ConnectionParams{
		Host:        "127.0.0.1",
		Port:        1161,
		Timeout:     5,
		Retries:     0,
		SnmpVersion: "2c",
		Community:   "test1",
	}

	cli := snmp.Client{
		ID:               "test",
		ConnectionParams: connectionParams,
		Log:              l,
	}
	_, err = cli.Connect([]string{})
	if err != nil {
		panic(err)
	}
	defer cli.Release()

	// 4.- METRICMAP SETUP

	metrics := map[string]*config.SnmpMetricCfg{
		"value_input": {
			ID:          "value_input",
			FieldName:   "T|myValue|STR,F|input|FP",
			Description: "",
			BaseOID:     ".1.1",
			DataSrcType: "MULTISTRINGPARSER",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "(.*);(.*)",
			Conversion:  0,
		},
		"value_output": {
			ID:          "value_output",
			FieldName:   "output",
			Description: "",
			BaseOID:     ".1.3",
			DataSrcType: "Integer32",
			GetRate:     false,
			Scale:       0.0,
			Shift:       0.0,
			IsTag:       false,
			ExtraData:   "",
			Conversion:  1,
		},
	}

	// 5.- MEASUREMENT CONFIG SETUP

	vars := map[string]interface{}{}

	cfg := &config.MeasurementCfg{
		ID:             "interfaces_data",
		Name:           "interfaces_data",
		GetMode:        "indexed",
		IndexOID:       ".1.2",
		TagOID:         "",
		IndexTag:       "portName",
		IndexTagFormat: "",
		Fields: []config.MeasurementFieldReport{
			{ID: "value_input", Report: metric.AlwaysReport},
			{ID: "value_output", Report: metric.AlwaysReport},
		},
	}

	cfg.Init(&metrics, vars)

	// 6.- MEASUREMENT ENGINE SETUP

	m := New(cfg, []string{}, map[string]*config.MeasFilterCfg{}, true, l)
	m.SetSNMPClient(cli)

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return
	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	// Measurement:interfaces_data Tags:{ myValue:value2, portName:eth2 } Field:output ValueType:int64  Value:22
	// Measurement:interfaces_data Tags:{ myValue:value2, portName:eth2 } Field:input ValueType:float64  Value:52
	// Measurement:interfaces_data Tags:{ myValue:value3, portName:eth3 } Field:input ValueType:float64  Value:53
	// Measurement:interfaces_data Tags:{ myValue:value3, portName:eth3 } Field:output ValueType:int64  Value:23
	// Measurement:interfaces_data Tags:{ myValue:value4, portName:eth4 } Field:output ValueType:int64  Value:24
	// Measurement:interfaces_data Tags:{ myValue:value1, portName:eth1 } Field:input ValueType:float64  Value:51
	// Measurement:interfaces_data Tags:{ myValue:value1, portName:eth1 } Field:output ValueType:int64  Value:21
}
