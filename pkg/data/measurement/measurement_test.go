package measurement

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/metric"
	"github.com/toni-moreno/snmpcollector/pkg/mock"
)

func ProcessMeasurementFull(m *Measurement, varmap map[string]interface{}) error {

	err := m.Init()
	if err != nil {
		return fmt.Errorf("Can not initialize measurement %s", err)
	}

	m.InitBuildRuntime()

	nGets, nProcs, nErrs, err := m.GetData()
	if err != nil {
		return fmt.Errorf("Error on get data from measurement %s", err)
	}

	m.ComputeOidConditionalMetrics()
	m.ComputeEvaluatedMetrics(varmap)

	m.Infof("GETS: %d,NPROCS: %d ,NERRS %d", nGets, nProcs, nErrs)
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

	metSent, metError, measSent, measError, ptarray := m.GetInfluxPoint(map[string]string{})

	m.Infof("METRIC SENT[%d],METRIC ERROR[%d],MEAS SENT[%d], MEAS ERROR[%d]", metSent, metError, measSent, measError)

	for _, v := range ptarray {
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
	//l.Level = logrus.DebugLevel

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

	cli := &gosnmp.GoSNMP{
		Target:    "127.0.0.1",
		Port:      1161,
		Version:   gosnmp.Version2c,
		Community: "test1",
		Timeout:   5 * time.Second,
		Retries:   0,
		Logger:    l,
	}
	err = cli.Connect()
	if err != nil {
		l.Fatalf("Connect() err: %v", err)
	}
	defer cli.Conn.Close()

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

	m, err := New(cfg, l, cli, false)
	if err != nil {
		l.Errorf("Can not create measurement %s", err)
		return

	}

	// 7.- PROCESS AND VERIFY

	err = ProcessMeasurementFull(m, vars)
	if err != nil {
		l.Errorf("Can not process measurement %s", err)
		return

	}

	GetOutputInfluxMetrics(m)

	// Unordered Output:
	//Measurement:test_name Tags:{} Field:metric_2_name ValueType:int64  Value:52
	//Measurement:test_name Tags:{} Field:metric_1_name ValueType:int64  Value:51

}

func Example_Measurement_GetMode_Indexed() {

	// 1.- SETUP LOGGER

	l := logrus.New()
	//l.Level = logrus.DebugLevel

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

	cli := &gosnmp.GoSNMP{
		Target:    "127.0.0.1",
		Port:      1161,
		Version:   gosnmp.Version2c,
		Community: "test1",
		Timeout:   5 * time.Second,
		Retries:   0,
		Logger:    l,
	}
	err = cli.Connect()
	if err != nil {
		l.Fatalf("Connect() err: %v", err)
	}
	defer cli.Conn.Close()

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

	m, err := New(cfg, l, cli, false)
	if err != nil {
		l.Errorf("Can not create measurement %s", err)
		return
	}

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

func Example_Measurement_value_STRINGEVAL() {

	// 1.- LOGGER SETUP

	l := logrus.New()
	//l.Level = logrus.DebugLevel

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

	cli := &gosnmp.GoSNMP{
		Target:    "127.0.0.1",
		Port:      1161,
		Version:   gosnmp.Version2c,
		Community: "test1",
		Timeout:   5 * time.Second,
		Retries:   0,
		Logger:    l,
	}
	err = cli.Connect()
	if err != nil {
		l.Fatalf("Connect() err: %v", err)
	}
	defer cli.Conn.Close()

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

	m, err := New(cfg, l, cli, false)
	if err != nil {
		l.Errorf("Can not create measurement %s", err)
		return
	}

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
	//l.Level = logrus.DebugLevel

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

	cli := &gosnmp.GoSNMP{
		Target:    "127.0.0.1",
		Port:      1161,
		Version:   gosnmp.Version2c,
		Community: "test1",
		Timeout:   5 * time.Second,
		Retries:   0,
		Logger:    l,
	}
	err = cli.Connect()
	if err != nil {
		l.Fatalf("Connect() err: %v", err)
	}
	defer cli.Conn.Close()

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

	m, err := New(cfg, l, cli, false)
	if err != nil {
		l.Errorf("Can not create measurement %s", err)
		return
	}

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
