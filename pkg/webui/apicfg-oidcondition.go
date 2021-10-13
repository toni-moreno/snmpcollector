package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgOidCondition OID Condition API REST creator
func NewAPICfgOidCondition(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/cfg/oidcondition", func() {
		m.Get("/", reqSignedIn, GetOidConditions)
		m.Get("/:id", reqSignedIn, GetOidConditionByID)
		m.Post("/", reqSignedIn, bind(config.OidConditionCfg{}), AddOidCondition)
		m.Put("/:id", reqSignedIn, bind(config.OidConditionCfg{}), UpdateOidCondition)
		m.Delete("/:id", reqSignedIn, DeleteOidCondition)
		m.Get("/checkondel/:id", reqSignedIn, GetOidConditionAffectOnDel)
	})
	return nil
}

// GetOidConditions Return metrics list to frontend
func GetOidConditions(ctx *Context) {
	// swagger:operation GET /cfg/oidcondition OID_Cond GetOidConditions
	//---
	// summary: Get all OID condition config from DB
	// description: Get All OID Conditions config from DB
	// tags:
	// - "OID Conditions Config"
	//
	// responses:
	//   '200':
	//     description: OID Condition Config Array
	//     schema:
	//       "$ref": "#/responses/idOfArrayOidCondResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	cfgarray, err := agent.MainConfig.Database.GetOidConditionCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get OID contition :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting OID contitions %+v", &cfgarray)
}

// GetOidConditionByID --pending--
func GetOidConditionByID(ctx *Context) {
	// swagger:operation GET /cfg/oidcondition/{id} OID_Cond GetOidConditionByID
	//---
	// summary: Get OID condition config from DB
	// description: Get OID Condition config from DB with specified ID
	// tags:
	// - "OID Conditions Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description:  OID condition ID to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: OID Condition Config
	//     schema:
	//       "$ref": "#/definitions/OidConditionCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetOidConditionCfgByID(id)
	if err != nil {
		log.Warningf("Error on get OidCondition  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddOidCondition Insert new condition to de internal BBDD --pending--
func AddOidCondition(ctx *Context, dev config.OidConditionCfg) {
	// swagger:operation POST /cfg/oidcondition OID_Cond AddOidCondition
	//---
	// summary: Add new OID Condition Config to DB
	// description: Add new  OID Condition Config from DB with specified ID
	// tags:
	// - "OID Conditions Config"
	//
	// parameters:
	// - name: OidConditionCfg
	//   in: body
	//   description: OidConditionCfg to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/OidConditionCfg"
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/OidConditionCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Printf("ADDING OidCondition %+v", dev)
	affected, err := agent.MainConfig.Database.AddOidConditionCfg(dev)
	if err != nil {
		log.Warningf("Error on insert OID condition %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateOidCondition Update OID contition
func UpdateOidCondition(ctx *Context, dev config.OidConditionCfg) {
	// swagger:operation PUT /cfg/oidcondition/{id} OID_Cond UpdateOidCondition
	//---
	// summary: Update Existing OID Condition Config to DB
	// description: Update Existing OID Condition Config from DB with specified ID
	// tags:
	// - "OID Conditions Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description:  OID condition ID to update
	//   required: true
	//   type: string
	// - name: OidConditionCfg
	//   in: body
	//   description: Updated OidCondition Config data
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/OidConditionCfg"
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/OidConditionCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateOidConditionCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update OID Condition %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

// DeleteOidCondition --pending--
func DeleteOidCondition(ctx *Context) {
	// swagger:operation DELETE /cfg/oidcondition/{id} OID_Cond DeleteOidCondition
	//---
	// summary: Delete Existing OID Condition Config to DB
	// description: Delete Existing OID Condition Config from DB with specified ID
	// tags:
	// - "OID Conditions Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description:  OID condition ID to update
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
	affected, err := agent.MainConfig.Database.DelOidConditionCfg(id)
	if err != nil {
		log.Warningf("Error on delete OidCondition %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

// GetOidConditionAffectOnDel --pending--
func GetOidConditionAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/oidcondition/checkondel/{id} OID_Cond GetOidConditionAffectOnDel
	//---
	// summary: Get List for affected Objects on delete ID
	// description: Get all existing Objects affected when deleted this OID Condition
	// tags:
	// - "OID Conditions Config"
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
	obarray, err := agent.MainConfig.Database.GetOidConditionCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for OID conditions  %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
