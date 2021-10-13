package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgSnmpMetric SnmpMetric  API REST creator
func NewAPICfgSnmpMetric(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/cfg/metric", func() {
		m.Get("/", reqSignedIn, GetMetrics)
		m.Get("/:id", reqSignedIn, GetMetricByID)
		m.Post("/", reqSignedIn, bind(config.SnmpMetricCfg{}), AddMetric)
		m.Put("/:id", reqSignedIn, bind(config.SnmpMetricCfg{}), UpdateMetric)
		m.Delete("/:id", reqSignedIn, DeleteMetric)
		m.Get("/checkondel/:id", reqSignedIn, GetMetricsAffectOnDel)
		m.Post("/convmodes", reqSignedIn, bind(config.SnmpMetricCfg{}), GetConversionModes)
	})

	return nil
}

// GetMetrics Return metrics list to frontend
func GetMetrics(ctx *Context) {
	// swagger:operation GET /cfg/metric  Config_Metric GetMetrics
	//---
	// summary: Get all Metrics Config from DB
	// description: Get All Metrics config from DB
	// tags:
	// - "Metrics Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayMetricResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	cfgarray, err := agent.MainConfig.Database.GetSnmpMetricCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Metrics :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Metrics %+v", &cfgarray)
}

// GetMetricByID --pending--
func GetMetricByID(ctx *Context) {
	// swagger:operation GET /cfg/metric/{id}  Config_Metric GetMetricByID
	//---
	// summary: Get Metric Config from DB
	// description: Get Metric config from DB for specified ID
	// tags:
	// - "Metrics Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Metric ID to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/SnmpMetricCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetSnmpMetricCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Metric  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddMetric Insert new metric to de internal BBDD --pending--
func AddMetric(ctx *Context, dev config.SnmpMetricCfg) {
	// swagger:operation POST /cfg/metric  Config_Metric AddMetric
	//---
	// summary: Add Metric Config into DB
	// description: Add Metric config into DB with posted data
	// tags:
	// - "Metrics Config"
	//
	// parameters:
	// - name: SnmpMetricCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpMetricCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/SnmpMetricCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Printf("ADDING Metric %+v", dev)
	affected, err := agent.MainConfig.Database.AddSnmpMetricCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMetric --pending--
func UpdateMetric(ctx *Context, dev config.SnmpMetricCfg) {
	// swagger:operation PUT /cfg/metric/{id}  Config_Metric UpdateMetric
	//---
	// summary: Update existing metric into DB
	// description: Update existing metric into DB from specified ID
	// tags:
	// - "Metrics Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Metric ID to update
	//   required: true
	//   type: string
	// - name: SnmpMetricCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpMetricCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/SnmpMetricCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateSnmpMetricCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

// DeleteMetric --pending--
func DeleteMetric(ctx *Context) {
	// swagger:operation DETELE /cfg/metric/{id}  Config_Metric DeleteMetric
	//---
	// summary: Delete existing metric in DB
	// description: Delete existing metric in DB from specified ID
	// tags:
	// - "Metrics Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Metric ID to delete
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/SnmpMetricCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelSnmpMetricCfg(id)
	if err != nil {
		log.Warningf("Error on delete Metric %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

// GetMetricsAffectOnDel --pending--
func GetMetricsAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/metric/checkondel/{id} Config_Metric GetMetricsAffectOnDel
	//---
	// summary: Get List for affected Objects on delete ID
	// description: Get List for affected Objects if deleting the  Metric with selected ID
	// tags:
	// - "Metrics Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The metric ID to check
	//   required: true
	//   type: string
	//
	// responses:
	//   '200':
	//     description: Object Array
	//     schema:
	//       "$ref": "#/responses/idOfCheckOnDelResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetSnmpMetricCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for SNMP metrics %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

// Conversion Item for selection
type ConversionItem struct {
	ID    int
	Value string
}

// ConversionItems array with all items and default/suggested value for this metric
// swagger:model ConversionItems
type ConversionItems struct {
	Default int
	Items   []ConversionItem
}

// GetConversionModes Return conversion modes from datasource Type
func GetConversionModes(ctx *Context, dev config.SnmpMetricCfg) {
	// swagger:operation GET /cfg/metric/convmodes Config_Metric GetConversionModes
	//---
	// summary: Get Info about conversion modes
	// description: Get suggested conversion modes from datasource Type
	// tags:
	// - "Metrics Config"
	//
	// parameters:
	// - name: SnmpMetricCfg
	//   in: body
	//   description: Metric witch would like to query for conversion modes
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpMetricCfg"
	//
	// responses:
	//   '200':
	//     description: Conversion Modes
	//     schema:
	//       "$ref": "#/definitions/ConversionItems"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	var citem []ConversionItem
	cfgarray, def, err := dev.GetValidConversions()
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get  Conversion Mode :%s", err)
		return
	}
	for _, v := range cfgarray {
		citem = append(citem, ConversionItem{ID: int(v), Value: v.GetString()})
	}
	response := &ConversionItems{Default: int(def), Items: citem}
	ctx.JSON(200, response)
	log.Debugf("Got Conversion Items Array Metrics %+v", response)
}
