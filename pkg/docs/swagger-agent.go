package docs

import (
	"time"

	"github.com/toni-moreno/snmpcollector/pkg/agent/device"
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

// swagger:response idOfDeviceStatResp
type rtAgentDeviceStatResponseWrapper struct {
	// in:body
	Body map[string]*device.DevStat
}

// swagger:response idOfArrayDeviceStatResp
type rtAgentWUIDeviceStatResponseWrapper struct {
	// in:body
	Body []*webui.DeviceStatMap
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

// swagger:response idOfCheckOnDelResp
type rtCfgCheckOnDelResponseWrapper struct {
	// in:body
	Body []*config.DbObjAction
}
