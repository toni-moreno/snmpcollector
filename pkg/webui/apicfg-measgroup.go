package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgMeasGroup MeasGroup API REST creator
func NewAPICfgMeasGroup(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/measgroup", func() {
		m.Get("/", reqSignedIn, GetMeasGroup)
		m.Get("/:id", reqSignedIn, GetMeasGroupByID)
		m.Post("/", reqSignedIn, bind(config.MGroupsCfg{}), AddMeasGroup)
		m.Put("/:id", reqSignedIn, bind(config.MGroupsCfg{}), UpdateMeasGroup)
		m.Delete("/:id", reqSignedIn, DeleteMeasGroup)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasGroupsAffectOnDel)
	})

	return nil
}

// GetMeasGroup Return measurements groups list to frontend
func GetMeasGroup(ctx *Context) {
	// swagger:operation GET /cfg/measgroup Config_MeasurementGroup GetMeasGroup
	//---
	// summary: Get measurement groups from DB
	// description: Get All measurement groups config info as an array from DB
	// tags:
	// - "Measurement Groups Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayMeasGroupResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	cfgarray, err := agent.MainConfig.Database.GetMGroupsCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Group :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Meas Group %+v", &cfgarray)
}

//GetMeasGroupByID --pending--
func GetMeasGroupByID(ctx *Context) {
	// swagger:operation GET /cfg/measgroup/{id} Config_MeasurementGroup GetMeasGroupByID
	//---
	// summary: Get measurement group from DB
	// description: Get measurement group config from DB with specified ID
	// tags:
	// - "Measurement Groups Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement Group ID to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MGroupsCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetMGroupsCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Group for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddMeasGroup Insert new measurement groups to de internal BBDD --pending--
func AddMeasGroup(ctx *Context, dev config.MGroupsCfg) {
	// swagger:operation POST /cfg/measgroup Config_MeasurementGroup AddMeasGroup
	//---
	// summary: Add new measurement group to DB
	// description: Add new measurement group to the Configuration DB
	// tags:
	// - "Measurement Groups Config"
	//
	// parameters:
	// - name: MeasGroupCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/MGroupsCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MGroupsCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Printf("ADDING Measurement Group %+v", dev)
	affected, err := agent.MainConfig.Database.AddMGroupsCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement Group %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasGroup --pending--
func UpdateMeasGroup(ctx *Context, dev config.MGroupsCfg) {
	// swagger:operation PUT /cfg/measgroup/{id} Config_MeasurementGroup UpdateMeasGroup
	//---
	// summary: Update existing measurement group into the DB
	// description: Update existing measurement group specified by ID to the Configuration DB
	// tags:
	// - "Measurement Groups Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement Group ID to update
	//   required: true
	//   type: string
	// - name: MeasGroupCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/MGroupsCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MGroupsCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMGroupsCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurement Group %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeasGroup --pending--
func DeleteMeasGroup(ctx *Context) {
	// swagger:operation DELETE /cfg/measgroup/{id} Config_MeasurementGroup DeleteMeasGroup
	//---
	// summary: Delete existing measurement group into the DB
	// description: Delete existing measurement group specified by ID to the Configuration DB
	// tags:
	// - "Measurement Groups Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement Group ID to delete
	//   required: true
	//   type: string
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelMGroupsCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Group %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMeasGroupsAffectOnDel --pending--
func GetMeasGroupsAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/measurement/checkondel/{id} Config_MeasurementGroup GetMeasGroupsAffectOnDel
	//---
	// summary: Get List for affected Objects on delete ID
	// description: Get List for affected Objects if deleting the measurement filter with ID
	// tags:
	// - "Measurement Groups Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The measurement Group ID to check
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
	obarray, err := agent.MainConfig.Database.GetMGroupsCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement Groups %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
