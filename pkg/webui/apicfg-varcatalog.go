package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgVarCatalog VarCatalog API REST creator
func NewAPICfgVarCatalog(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/varcatalog", func() {
		m.Get("/", reqSignedIn, GetVarCatalog)
		m.Get("/:id", reqSignedIn, GetVarCatalogByID)
		m.Post("/", reqSignedIn, bind(config.VarCatalogCfg{}), AddVarCatalog)
		m.Put("/:id", reqSignedIn, bind(config.VarCatalogCfg{}), UpdateVarCatalog)
		m.Delete("/:id", reqSignedIn, DeleteVarCatalog)
		m.Get("/checkondel/:id", reqSignedIn, GetInfluxAffectOnDel)
	})

	return nil
}

// GetVarCatalog Return Server Array
func GetVarCatalog(ctx *Context) {
	// swagger:operation GET /cfg/varcatalog  Config_VarCat GetVarCatalog
	//---
	// summary: Get All variables in catalog DB
	// description: Get All variables in catalog DB
	// tags:
	// - "Variable Catalog Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayVarCatResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	cfgarray, err := agent.MainConfig.Database.GetVarCatalogCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get VarCatalogiable :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Gloval Variable s %+v", &cfgarray)
}

//GetVarCatalogByID --pending--
func GetVarCatalogByID(ctx *Context) {
	// swagger:operation GET /cfg/varcatalog/{id}  Config_VarCat GetVarCatalogByID
	//---
	// summary: Get all Variable Catalog Config from DB
	// description: Get all Variable Catalog config from DB for specified ID
	// tags:
	// - "Variable Catalog Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Variable Catalog ID to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/VarCatalogCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetVarCatalogCfgByID(id)
	if err != nil {
		log.Warningf("Error on get gloval variable %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddVarCatalog Insert new global var into the database
func AddVarCatalog(ctx *Context, dev config.VarCatalogCfg) {
	// swagger:operation POST /cfg/varcatalog  Config_VarCat AddVarCatalog
	//---
	// summary: Add new Global Variable into the DB Catalog
	// description: Add new  Global Variable into the DB Catalog
	// tags:
	// - "Variable Catalog Config"
	//
	// parameters:
	// - name: VarCatalogCfg
	//   in: body
	//   description: Variable catalog to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/VarCatalogCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/VarCatalogCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Printf("ADDING Global Variable %+v", dev)
	affected, err := agent.MainConfig.Database.AddVarCatalogCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Global Variable %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateVarCatalog -
func UpdateVarCatalog(ctx *Context, dev config.VarCatalogCfg) {
	// swagger:operation PUT /cfg/varcatalog/{id}  Config_VarCat UpdateVarCatalog
	//---
	// summary: Update existing variable into the  DB catalog
	// description: Update existing variable into the  DB catalog
	// tags:
	// - "Variable Catalog Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Variable Catalog ID to update
	//   required: true
	//   type: string
	// - name: VarCatalogCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/VarCatalogCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/VarCatalogCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateVarCatalogCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Global Variable %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

// DeleteVarCatalog
func DeleteVarCatalog(ctx *Context) {
	// swagger:operation DETELE /cfg/varcatalog/{id}  Config_VarCat DeleteVarCatalog
	//---
	// summary: Delete existing variable in DB catalog
	// description: Delete existing variable in DB catalog from specified ID
	// tags:
	// - "Variable Catalog Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Variable Catalog ID to delete
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/VarCatalogCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Trying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelVarCatalogCfg(id)
	if err != nil {
		log.Warningf("Error on delete Global Variable %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetVarCatalogAffectOnDel --pending--
func GetVarCatalogAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/varcatalog/checkondel/{id} Config_VarCat GetVarCatalogAffectOnDel
	//---
	// summary: Get List for affected Objects on delete ID
	// description: Get List for affected Objects if deleting the  Variable with selected ID
	// tags:
	// - "Variable Catalog Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The variable catalog ID to check
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
	obarray, err := agent.MainConfig.Database.GetVarCatalogCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
