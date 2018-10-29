package webui

import (
	"github.com/go-macaron/binding"
	"snmpcollector/pkg/agent"
	"snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgCustomFilter CustomFilter REST API creator
func NewAPICfgCustomFilter(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/customfilter", func() {
		m.Get("/", reqSignedIn, GetCustomFilter)
		m.Post("/", reqSignedIn, bind(config.CustomFilterCfg{}), AddCustomFilter)
		m.Put("/:id", reqSignedIn, bind(config.CustomFilterCfg{}), UpdateCustomFilter)
		m.Delete("/:id", reqSignedIn, DeleteCustomFilter)
		m.Get("/:id", reqSignedIn, GetCustomFilterByID)
		m.Get("/checkondel/:id", reqSignedIn, GetCustomFiltersAffectOnDel)
	})

	return nil
}

// GetCustomFilter Return measurements groups list to frontend
func GetCustomFilter(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetCustomFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Custom Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

// AddCustomFilter Insert new measurement groups to de internal BBDD --pending--
func AddCustomFilter(ctx *Context, dev config.CustomFilterCfg) {
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

//GetCustomFilterByID --pending--
func GetCustomFilterByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetCustomFilterCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Filter  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetCustomFiltersAffectOnDel --pending--
func GetCustomFiltersAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetCustomFilterCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement filters %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
