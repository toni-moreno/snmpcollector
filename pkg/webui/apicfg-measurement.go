package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgMeasurement Measurement API REST creator
func NewAPICfgMeasurement(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/cfg/measurement", func() {
		m.Get("/", reqSignedIn, GetMeas)
		m.Get("/:id", reqSignedIn, GetMeasByID)
		m.Get("/type/:type", reqSignedIn, GetMeasByType)
		m.Post("/", reqSignedIn, bind(config.MeasurementCfg{}), AddMeas)
		m.Put("/:id", reqSignedIn, bind(config.MeasurementCfg{}), UpdateMeas)
		m.Delete("/:id", reqSignedIn, DeleteMeas)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasAffectOnDel)
	})

	return nil
}

/****************/
/* MEASUREMENTS
/****************/

// GetMeas Return measurements list to frontend
func GetMeas(ctx *Context) {
	// swagger:operation GET /cfg/measurement Config_Measurement GetMeas
	//---
	// summary: Get All Configured Measurements from DB
	// description: Get All Configured Measurements from DB
	// tags:
	// - "Measurements Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayMeasResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	cfgarray, err := agent.MainConfig.Database.GetMeasurementCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx Measurements :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurements %+v", &cfgarray)
}

// GetMeasByID --pending--
func GetMeasByID(ctx *Context) {
	// swagger:operation GET /cfg/measurement/{id} Config_Measurement GetMeasByID
	//---
	// summary: Get Measurement Config from DB
	// description: Get Configured Measurements from DB with specified ID
	// tags:
	// - "Measurements Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement ID to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MeasurementCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetMeasurementCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// GetMeasByType Return measurements list to frontend
func GetMeasByType(ctx *Context) {
	// swagger:operation GET /cfg/measurement/type/{type} Config_Measurement GetMeasByType
	//---
	// summary: Get Measurement Config by type from DB
	// description: Get Configured Measurements from DB with specified type
	// tags:
	// - "Measurements Config"
	//
	// parameters:
	// - name: type
	//   in: path
	//   description: Measurement type to query
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayMeasResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	t := ctx.Params(":type")
	cfgarray, err := agent.MainConfig.Database.GetMeasurementCfgArray("getmode like '%" + t + "%'")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx Measurements :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurements %+v", &cfgarray)
}

// AddMeas Insert new measurement to de internal BBDD --pending--
func AddMeas(ctx *Context, dev config.MeasurementCfg) {
	// swagger:operation POST /cfg/measurement Config_Measurement AddMeas
	//---
	// summary: Add new Measurement Config into DB
	// description: Add new Measurements config into DB
	// tags:
	// - "Measurements Config"
	//
	// parameters:
	// - name: MeasurementCfg
	//   in: body
	//   description: Measurement to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/MeasurementCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MeasurementCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Printf("ADDING Measurement %+v", dev)
	affected, err := agent.MainConfig.Database.AddMeasurementCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeas --pending--
func UpdateMeas(ctx *Context, dev config.MeasurementCfg) {
	// swagger:operation PUT /cfg/measurement/{id} Config_Measurement UpdateMeas
	//---
	// summary: Update existing Measurement Config into DB
	// description: Update existing Measurements config into DB with specified ID
	// tags:
	// - "Measurements Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement ID to update
	//   required: true
	//   type: string
	// - name: MeasurementCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/MeasurementCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MeasurementCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMeasurementCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

// DeleteMeas --pending--
func DeleteMeas(ctx *Context) {
	// swagger:operation DETELE /cfg/measurement/{id}  Config_Measurement DeleteMeas
	//---
	// summary: Delete existing Measurement Config into DB
	// description: Delete existing Measurements config in DB with specified ID
	// tags:
	// - "Measurements Config"
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
	affected, err := agent.MainConfig.Database.DelMeasurementCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

// GetMeasAffectOnDel --pending--
func GetMeasAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/measurement/checkondel/{id} Config_Measurement GetMeasAffectOnDel
	//---
	// summary: Get List for affected Objects on delete ID
	// description: Get all existing Objects affected when deleted the measurement.
	// tags:
	// - "Measurements Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The measurement ID to check
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
	obarray, err := agent.MainConfig.Database.GetMeasurementCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurements %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
