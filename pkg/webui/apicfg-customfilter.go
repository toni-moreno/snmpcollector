package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgCustomFilter CustomFilter REST API creator
func NewAPICfgCustomFilter(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/customfilter", func() {
		m.Get("/", reqSignedIn, GetCustomFilter)
		m.Get("/:id", reqSignedIn, GetCustomFilterByID)
		m.Post("/", reqSignedIn, bind(config.CustomFilterCfg{}), AddCustomFilter)
		m.Put("/:id", reqSignedIn, bind(config.CustomFilterCfg{}), UpdateCustomFilter)
		m.Delete("/:id", reqSignedIn, DeleteCustomFilter)
		m.Get("/checkondel/:id", reqSignedIn, GetCustomFiltersAffectOnDel)
	})

	return nil
}

// GetCustomFilter Return measurements groups list to frontend
func GetCustomFilter(ctx *Context) {
	// swagger:operation GET /cfg/customfilter  Config_CustomFilter GetCustomFilter
	//---
	// summary: Get All Custom Filters info
	// description:  Get All Custom Filters config info as an array
	// tags:
	// - "CustomFilter Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayCustomFilterCfgResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	cfgarray, err := agent.MainConfig.Database.GetCustomFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Custom Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

//GetCustomFilterByID --pending--
func GetCustomFilterByID(ctx *Context) {
	// swagger:operation GET /cfg/customfilter/{id} Config_CustomFilter GetCustomFilterByID
	//---
	// summary: Get CustomFilter Info for this ID
	// description: Get CustomFilter Info for this ID
	// tags:
	// - "CustomFilter Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: CustomFilter to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/CustomFilterCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetCustomFilterCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Filter  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddCustomFilter Insert new measurement groups to de internal BBDD --pending--
func AddCustomFilter(ctx *Context, dev config.CustomFilterCfg) {
	// swagger:operation POST /cfg/customfilter Config_CustomFilter AddCustomFilter
	//---
	// summary: Add CustomFilter
	// description: Add  CustomFilter with body Filter data
	// tags:
	// - "CustomFilter Config"
	//
	// parameters:
	// - name: CustomFilterCfg
	//   in: body
	//   description: Custom Filter to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/CustomFilterCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/CustomFilterCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Printf("ADDING measurement Filter %+v", dev)
	affected, err := agent.MainConfig.Database.AddCustomFilterCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateCustomFilter --pending--
func UpdateCustomFilter(ctx *Context, dev config.CustomFilterCfg) {
	// swagger:operation PUT /cfg/customfilter/{id} Config_CustomFilter UpdateCustomFilter
	//---
	// summary: Update CustomFilter
	// description: Update  CustomFilter with defined config data
	// tags:
	// - "CustomFilter Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Custom Filter ID to update
	//   required: true
	//   type: string
	// - name: CustomFilterCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/CustomFilterCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/CustomFilterCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateCustomFilterCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteCustomFilter --pending--
func DeleteCustomFilter(ctx *Context) {
	// swagger:operation DELETE /cfg/customfilter/{id} Config_CustomFilter DeleteCustomFilter
	//---
	// summary: Delete CustomFilter
	// description: Delete  CustomFilter with defined id
	// tags:
	// - "CustomFilter Config"
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
	affected, err := agent.MainConfig.Database.DelCustomFilterCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Filter %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetCustomFiltersAffectOnDel --pending--
func GetCustomFiltersAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/customfilter/checkondel/{id} Config_CustomFilter GetCustomFiltersAffectOnDel
	//---
	// summary: Get all Strng CustomFilter
	// description: Delete  CustomFilter with defined id
	// tags:
	// - "CustomFilter Config"
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
	obarray, err := agent.MainConfig.Database.GetCustomFilterCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement filters %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
