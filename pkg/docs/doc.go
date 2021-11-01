// Package docs SnmpCollector
//
// Documentation of SnmpCollector API, runtime, and config operations.
//
//     Schemes: http, https
//
//     BasePath: /api
//     Version: 0.10.1
//     License: MIT http://opensource.org/licenses/MIT
//     Contact: Toni Moreno<toni.moreno@gmail.com> http://snmpcollector.org
//     Security:
//     - basic
//
//    SecurityDefinitions:
//    basicAuth:
//      type: basic
//    security:
//     	- basicAuth:
//
// swagger:meta
package docs

import (
	"time"

	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"github.com/toni-moreno/snmpcollector/pkg/data/stats"
	"github.com/toni-moreno/snmpcollector/pkg/webui"
)

// swagger:response idOfSnmpSysInfoResp
type rtAgentSnmpSysInfoResponseWrapper struct {
	// SNMP System Info
	// in:body
	Body snmp.SysInfo
}

// swagger:response idOfStringResp
type rtAgentStringErrorResponseWrapper struct {
	// Error response
	// in:body
	Body string
}

// swagger:response idOfDurationResp
type rtAgentDurationResponseWrapper struct {
	// in:body
	Body time.Duration
}

// swagger:response idOfDeviceStatResp
type rtAgentDeviceStatResponseWrapper struct {
	// in:body
	Body map[string]*stats.GatherStats
}

// swagger:response idOfArrayDeviceStatResp
type rtAgentWUIDeviceStatResponseWrapper struct {
	// in:body
	Body []*webui.DeviceStatMap
}

// swagger:response idOfArrayInfluxCfgResp
type rtCfgArrayInfluxCfgResponseWrapper struct {
	// in:body
	Body []*config.InfluxCfg
}

// swagger:response idOfArrayCustomFilterCfgResp
type rtCfgArrayCustomFilterCfgResponseWrapper struct {
	// in:body
	Body []*config.CustomFilterCfg
}

// swagger:response idOfArrayMetricResp
type rtCfgArrayMetricResponseWrapper struct {
	// in:body
	Body []*config.SnmpMetricCfg
}

// swagger:response idOfArrayMeasGroupResp
type rtCfgArrayMeasGroupResponseWrapper struct {
	// in:body
	Body []*config.MGroupsCfg
}

// swagger:response idOfArrayMeasResp
type rtCfgArrayMeasResponseWrapper struct {
	// in:body
	Body []*config.MeasurementCfg
}

// swagger:response idOfArrayMeasFilterResp
type rtCfgArrayMeasFilterResponseWrapper struct {
	// in:body
	Body []*config.MeasFilterCfg
}

// swagger:response idOfArrayOidCondResp
type rtCfgArrayOidCondResponseWrapper struct {
	// in:body
	Body []*config.OidConditionCfg
}

// swagger:response idOfArrayVarCatResp
type rtCfgArrayVarCatResponseWrapper struct {
	// in:body
	Body []*config.VarCatalogCfg
}

// swagger:response idOfCheckOnDelResp
type rtCfgCheckOnDelResponseWrapper struct {
	// in:body
	Body []*config.DbObjAction
}
