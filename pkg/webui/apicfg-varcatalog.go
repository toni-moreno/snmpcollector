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
		m.Post("/", reqSignedIn, bind(config.VarCatalogCfg{}), AddVarCatalog)
		m.Put("/:id", reqSignedIn, bind(config.VarCatalogCfg{}), UpdateVarCatalog)
		m.Delete("/:id", reqSignedIn, DeleteVarCatalog)
		m.Get("/:id", reqSignedIn, GetVarCatalogByID)
		m.Get("/checkondel/:id", reqSignedIn, GetInfluxAffectOnDel)
	})

	return nil
}

// GetVarCatalog Return Server Array
func GetVarCatalog(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetVarCatalogCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get VarCatalogiable :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Gloval Variable s %+v", &cfgarray)
}

// AddVarCatalog Insert new global var into the database
func AddVarCatalog(ctx *Context, dev config.VarCatalogCfg) {
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

// UpdateVarCatalog --pending--
func UpdateVarCatalog(ctx *Context, dev config.VarCatalogCfg) {
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

//DeleteVarCatalog --pending--
func DeleteVarCatalog(ctx *Context) {
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

//GetVarCatalogByID --pending--
func GetVarCatalogByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetVarCatalogCfgByID(id)
	if err != nil {
		log.Warningf("Error on get gloval variable %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetVarCatalogAffectOnDel --pending--
func GetVarCatalogAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetVarCatalogCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
