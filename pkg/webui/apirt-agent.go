package webui

import (
	"strings"
	"time"

	"github.com/go-macaron/binding"
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
		m.Post("/snmpconsole/ping/", reqSignedIn, bind(config.SnmpDeviceCfg{}), PingSNMPDevice)
		m.Post("/snmpconsole/query/:getmode/:obtype/:data", reqSignedIn, bind(config.SnmpDeviceCfg{}), QuerySNMPDevice)
		m.Get("/info/version/", RTGetVersion)
	})

	return nil
}

// AgentReloadConf xx
func AgentReloadConf(ctx *Context) {
	log.Info("trying to reload configuration for all devices")
	time, err := agent.ReloadConf()
	if err != nil {
		ctx.JSON(405, err.Error())
		return
	}
	ctx.JSON(200, time)
}

//PingSNMPDevice xx
func PingSNMPDevice(ctx *Context, cfg config.SnmpDeviceCfg) {
	log.Infof("trying to ping device %s : %+v", cfg.ID, cfg)

	_, sysinfo, err := snmp.GetClient(&cfg, log, "ping", false, 0)
	if err != nil {
		log.Debugf("ERROR  on query device : %s", err)
		ctx.JSON(400, err.Error())
	} else {
		log.Debugf("OK on query device ")
		ctx.JSON(200, sysinfo)
	}
}

// QuerySNMPDevice xx
func QuerySNMPDevice(ctx *Context, cfg config.SnmpDeviceCfg) {
	getmode := ctx.Params(":getmode")
	obtype := ctx.Params(":obtype")
	data := strings.TrimSpace(ctx.Params(":data"))

	log.Infof("trying to query device %s : getmode: %s objectype: %s data %s", cfg.ID, getmode, obtype, data)

	if obtype != "oid" {
		log.Warnf("Object Type [%s] Not Supperted", obtype)
		ctx.JSON(400, "Object Type [ "+obtype+"] Not Supperted")
		return
	}

	snmpcli, info, err := snmp.GetClient(&cfg, log, "query", false, 0)
	if err != nil {
		log.Debugf("ERROR  on open connection with device %s : %s", cfg.ID, err)
		ctx.JSON(400, err.Error())
		return
	}
	start := time.Now()
	result, err := snmp.Query(snmpcli, getmode, data)
	elapsed := time.Since(start)
	if err != nil {
		log.Debugf("ERROR  on query device : %s", err)
		ctx.JSON(400, err.Error())
		return
	}
	log.Debugf("OK on query device ")
	snmpdata := struct {
		DeviceCfg   *config.SnmpDeviceCfg
		TimeTaken   float64
		PingInfo    *snmp.SysInfo
		QueryResult []snmp.EasyPDU
	}{
		&cfg,
		elapsed.Seconds(),
		info,
		result,
	}
	ctx.JSON(200, snmpdata)
}

//RTGetVersion xx
func RTGetVersion(ctx *Context) {
	info := agent.GetRInfo()
	ctx.JSON(200, &info)
}
