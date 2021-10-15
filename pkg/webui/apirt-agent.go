package webui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-macaron/binding"
	"github.com/sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"gopkg.in/macaron.v1"
)

// NewAPIRtAgent Runtime Agent REST API creator
func NewAPIRtAgent(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/rt/agent", func() {
		m.Get("/reload/", reqSignedIn, AgentReloadConf)
		m.Get("/shutdown/", reqSignedIn, AgentShutdown)
		m.Post("/snmpconsole/ping/", reqSignedIn, bind(config.SnmpDeviceCfg{}), PingSNMPDevice)
		m.Post("/snmpconsole/query/:getmode/:obtype/:data", reqSignedIn, bind(config.SnmpDeviceCfg{}), QuerySNMPDevice)
		m.Get("/info/version/", RTGetVersion)
	})

	return nil
}

// AgentReloadConf xx
func AgentReloadConf(ctx *Context) {
	// swagger:operation GET /rt/agent/reload Runtime_Agent AgentReloadConf
	//---
	// summary: Reload Configuration and restart devices
	// description: Reload Configuration and restart devices
	// tags:
	// - "Runtime Agent"
	// responses:
	//   '200':
	//     description: Reload Duration in miliseconds
	//     schema:
	//       "$ref": "#/responses/idOfDurationResp"
	//   '405':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Info("trying to reload configuration for all devices")
	time, err := agent.ReloadConf()
	if err != nil {
		ctx.JSON(405, err.Error())
		return
	}
	ctx.JSON(200, time)
}

// AgentShutdown xx
func AgentShutdown(ctx *Context) {
	// swagger:operation GET /rt/agent/shutdown Runtime_Agent AgentShutdown
	//---
	// summary: Finalices inmediately the process
	// description: shutdown the process , (usefull only with some external restart tools )
	// tags:
	// - "Runtime Agent"
	//
	// responses:
	//   '200':
	//     description: Reload Duration in miliseconds
	//     schema:
	//       "$ref": "#/responses/idOfDurationResp"

	log.Info("receiving shutdown")
	ctx.JSON(200, "Init shutdown....")
	os.Exit(0)
}

// PingSNMPDevice xx
func PingSNMPDevice(ctx *Context, cfg config.SnmpDeviceCfg) {
	// swagger:operation POST /rt/agent/snmpconsole/ping Runtime_SNMP_Console PingSNMPDevice
	//---
	// summary:  Connectivity test to the device
	// description: |
	//    Check connectivity by test snmp connection and  will return Basic system Info from SNMP device
	// tags:
	// - "SNMP Console Tool"
	//
	// parameters:
	// - name: SnmpDeviceCfg
	//   in: body
	//   description: device to query
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//
	// responses:
	//   '200':
	//     description: snmp responses
	//     schema:
	//       "$ref": "#/definitions/SnmpQueryResponse"
	//   '400':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	l := log.WithFields(logrus.Fields{
		"id": cfg.ID,
	})
	l.Infof("trying to ping device, config: %+v", cfg)

	connectionParams := snmp.ConnectionParams{
		Host:           cfg.Host,
		Port:           cfg.Port,
		Timeout:        cfg.Timeout,
		Retries:        cfg.Retries,
		SnmpVersion:    cfg.SnmpVersion,
		Community:      cfg.Community,
		MaxRepetitions: cfg.MaxRepetitions,
		MaxOids:        cfg.MaxOids,
		Debug:          cfg.SnmpDebug,
		V3Params: snmp.V3Params{
			SecLevel:        cfg.V3SecLevel,
			AuthUser:        cfg.V3AuthUser,
			AuthPass:        cfg.V3AuthPass,
			PrivPass:        cfg.V3PrivPass,
			PrivProt:        cfg.V3PrivProt,
			AuthProt:        cfg.V3AuthProt,
			ContextName:     cfg.V3ContextName,
			ContextEngineID: cfg.V3ContextEngineID,
		},
	}
	err := connectionParams.Validation()
	if err != nil {
		l.Debugf("ERROR on query device : %s", err)
		ctx.JSON(400, fmt.Errorf("SNMP parameter validation: %v", err))
		return
	}

	snmpClient := snmp.Client{
		ID:               cfg.Host,
		DisableBulk:      cfg.DisableBulk,
		ConnectionParams: connectionParams,
		Log:              l,
	}
	sysinfo, err := snmpClient.Connect(cfg.SystemOIDs)
	if err != nil {
		l.Debugf("ERROR on query device : %s", err)
		ctx.JSON(400, fmt.Errorf("unable to connect: %v", err))
		return
	}

	l.Debugf("OK on query device")
	ctx.JSON(200, sysinfo)
}

// SnmpQueryResponse response for queries in the UI
// swagger:model SnmpQueryResponse
type SnmpQueryResponse struct {
	DeviceCfg   *config.SnmpDeviceCfg
	TimeTaken   float64
	PingInfo    *snmp.SysInfo
	QueryResult []snmp.EasyPDU
}

// QuerySNMPDevice xx
func QuerySNMPDevice(ctx *Context, cfg config.SnmpDeviceCfg) {
	// swagger:operation POST /rt/agent/snmpconsole/query/{getmode}/{obtype}/{data} Runtime_SNMP_Console QuerySNMPDevice
	//---
	// summary:  Run a SNMP Query for a device
	// description: |
	//    Check connectivity by test snmp connection with Device configuration and  will return Basic system Info for the remote SNMP device
	// tags:
	// - "SNMP Console Tool"
	//
	// parameters:
	// - name: getmode
	//   in: path
	//   description: SNMP Get type
	//   required: true
	//   type: string
	//   enum: [get,walk]
	// - name: obtype
	//   in: path
	//   description: type of object in (snmpmetric,snmpmeasurement,...)
	//   required: true
	//   type: string
	//   enum: [snmpmetric,snmpmeasurement]
	// - name: data
	//   in: path
	//   description: id for the objecttype to qyery (snmpmetric,snmpmeasurement,...)
	//   required: true
	//   type: string
	// - name: SnmpDeviceCfg
	//   in: body
	//   description: device to query
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//
	// responses:
	//   '200':
	//     description: snmp responses
	//     schema:
	//       "$ref": "#/definitions/SnmpQueryResponse"
	//   '400':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	getmode := ctx.Params(":getmode")
	obtype := ctx.Params(":obtype")
	data := strings.TrimSpace(ctx.Params(":data"))

	l := log.WithFields(logrus.Fields{
		"id": cfg.ID,
	})
	l.Infof("trying to query device : getmode: %s objectype: %s data %s", getmode, obtype, data)

	if obtype != "oid" {
		l.Warnf("Object Type [%s] Not Supperted", obtype)
		ctx.JSON(400, "Object Type [ "+obtype+"] Not Supperted")
		return
	}

	connectionParams := snmp.ConnectionParams{
		Host:           cfg.Host,
		Port:           cfg.Port,
		Timeout:        cfg.Timeout,
		Retries:        cfg.Retries,
		SnmpVersion:    cfg.SnmpVersion,
		Community:      cfg.Community,
		MaxRepetitions: cfg.MaxRepetitions,
		MaxOids:        cfg.MaxOids,
		Debug:          cfg.SnmpDebug,
		V3Params: snmp.V3Params{
			SecLevel:        cfg.V3SecLevel,
			AuthUser:        cfg.V3AuthUser,
			AuthPass:        cfg.V3AuthPass,
			PrivPass:        cfg.V3PrivPass,
			PrivProt:        cfg.V3PrivProt,
			AuthProt:        cfg.V3AuthProt,
			ContextName:     cfg.V3ContextName,
			ContextEngineID: cfg.V3ContextEngineID,
		},
	}
	err := connectionParams.Validation()
	if err != nil {
		l.Debugf("ERROR on query device : %s", err)
		ctx.JSON(400, fmt.Errorf("SNMP parameter validation: %v", err))
		return
	}

	snmpClient := snmp.Client{
		ID:               cfg.Host,
		DisableBulk:      cfg.DisableBulk,
		ConnectionParams: connectionParams,
		Log:              l,
	}
	sysinfo, err := snmpClient.Connect(cfg.SystemOIDs)
	if err != nil {
		l.Debugf("ERROR on query device : %s", err)
		ctx.JSON(400, fmt.Errorf("unable to connect: %v", err))
		return
	}

	start := time.Now()
	result, err := snmpClient.Query(getmode, data)
	elapsed := time.Since(start)
	if err != nil {
		l.Debugf("ERROR  on query device : %s", err)
		ctx.JSON(400, fmt.Errorf("unable to query: %v", err))
		return
	}

	l.Debugf("OK on query device")
	snmpdata := SnmpQueryResponse{
		&cfg,
		elapsed.Seconds(),
		sysinfo,
		result,
	}
	ctx.JSON(200, snmpdata)
}

// RTGetVersion xx
func RTGetVersion(ctx *Context) {
	// swagger:operation GET /rt/agent/info/version Runtime_Agent RTGetVersion
	//---
	// summary: Get Agent Version
	// description: Get Agent Version, release , commit , compilation day
	// tags:
	// - "Runtime Agent"
	//
	// security: []
	//
	// responses:
	//   '200':
	//     description: Agent Version Info
	//     schema:
	//      "$ref": "#/definitions/RInfo"

	info := agent.GetRInfo()
	ctx.JSON(200, &info)
}
