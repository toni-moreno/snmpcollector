package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgOutput Output API REST creator
func NewAPICfgOutput(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/cfg/output", func() {
		m.Get("/", reqSignedIn, GetOutput)
		m.Get("/:id", reqSignedIn, GetOutputByID)
		m.Post("/", reqSignedIn, bind(config.OutputCfg{}), AddOutput)
		m.Put("/:id", reqSignedIn, bind(config.OutputCfg{}), UpdateOutput)
		m.Delete("/:id", reqSignedIn, DeleteOutput)
		m.Get("/checkondel/:id", reqSignedIn, GetOutputAffectOnDel)
	})

	return nil
}

// GetOutput Return Server Array
func GetOutput(ctx *Context) {
	// swagger:operation GET /cfg/output  Config_Outputs GetOutput
	//---
	// summary: Get All Output Config Items from DB
	// description: Get All Output Config Items as an array from DB
	// tags:
	// - "Output Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayOutputCfgResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	cfgarray, err := agent.MainConfig.Database.GetOutputCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Output :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// GetOutputByID --pending--
func GetOutputByID(ctx *Context) {
	// swagger:operation GET /cfg/output/{id}  Config_Outputs GetOutputByID
	//---
	// summary: Get Output Config from DB
	// description: Get Outputs config info by ID from DB
	// tags:
	// - "Output Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Output to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/OutputCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetOutputCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Output with id %s, error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddOutput Insert new measurement groups to de internal BBDD --pending--
func AddOutput(ctx *Context, dev config.OutputCfg) {
	// swagger:operation POST /cfg/output Config_Outputs AddOutput
	//---
	// summary: Add new Output Config
	// description: Add Output from Data
	// tags:
	// - "Output Config"
	//
	// parameters:
	// - name: OutputCfg
	//   in: body
	//   description: OutputConfig to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/OutputCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/OutputCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	log.Printf("ADDING Output Backend %+v", dev)
	affected, err := agent.MainConfig.Database.AddOutputCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Backend %s, affected: %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateOutput --pending--
func UpdateOutput(ctx *Context, dev config.OutputCfg) {
	// swagger:operation PUT /cfg/output/{id} Config_Outputs UpdateOutput
	//---
	// summary: Update Output Config
	// description: Update Output from Data with specified ID
	// tags:
	// - "Output Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Output Config ID to update
	//   required: true
	//   type: string
	// - name: OutputCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/OutputCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/OutputCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateOutputCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Output %s, affected: %+v , error: %s", dev.ID, affected, err)
	} else {
		// TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

// DeleteOutput --pending--
func DeleteOutput(ctx *Context) {
	// swagger:operation DELETE /cfg/output/{id} Config_Outputs DeleteOutput
	//---
	// summary: Delete Output Config on DB
	// description: Delete Output on DB with specified ID
	// tags:
	// - "Output Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The Output ID to delete
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
	affected, err := agent.MainConfig.Database.DelOutputCfg(id)
	if err != nil {
		log.Warningf("Error on delete Output %s, affected : %+v, error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

// GetOutputAffectOnDel --pending--
func GetOutputAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/output/checkondel/{id} Config_Outputs GetOutputAffectOnDel
	//---
	// summary: Check affected sources.
	// description: Get all existing Objects affected when deleted the Output.
	// tags:
	// - "Output Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The Output ID to check
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
	obarray, err := agent.MainConfig.Database.GetOutputCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Outputs %s, error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
