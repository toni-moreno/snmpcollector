package docs

import (
	"time"

	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
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

// swagger:response idOfInfoResp
type rtAgentInfoResponseWrapper struct {
	// in:body
	Body agent.RInfo
}

// swagger:response idOfSnmpQueryResp
type rtAgentSnmpQueryResponseWrapper struct {
	// in:body
	Body webui.SnmpQueryResponse
}

// swagger:parameters idOfDeviceCfg
type rtAgentSnmpDeviceCfgParamsWrapper struct {
	// SnmpDevice Config parameters
	// in:body
	Body config.SnmpDeviceCfg
}
