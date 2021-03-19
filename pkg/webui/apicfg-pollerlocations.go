package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgPollerLocation PollerLocation API REST creator
func NewAPICfgPollerLocation(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/pollerlocations", func() {
		m.Get("/", reqSignedIn, GetPollerLocation)
		m.Post("/", reqSignedIn, bind(config.PollerLocationCfg{}), AddPollerLocation)
		m.Put("/:id", reqSignedIn, bind(config.PollerLocationCfg{}), UpdatePollerLocation)
		m.Delete("/:id", reqSignedIn, DeletePollerLocation)
		m.Get("/:id", reqSignedIn, GetPollerLocationByID)
		m.Get("/checkondel/:id", reqSignedIn, GetInfluxAffectOnDel)
	})

	return nil
}

// GetPollerLocation Return Server Array
func GetPollerLocation(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetPollerLocationCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get PollerLocationiable :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Gloval Variable s %+v", &cfgarray)
}

// AddPollerLocation Insert new global var into the database
func AddPollerLocation(ctx *Context, dev config.PollerLocationCfg) {
	log.Printf("ADDING Global Variable %+v", dev)
	affected, err := agent.MainConfig.Database.AddPollerLocationCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Global Variable %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdatePollerLocation --pending--
func UpdatePollerLocation(ctx *Context, dev config.PollerLocationCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdatePollerLocationCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Global Variable %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeletePollerLocation --pending--
func DeletePollerLocation(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Trying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelPollerLocationCfg(id)
	if err != nil {
		log.Warningf("Error on delete Global Variable %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetPollerLocationByID --pending--
func GetPollerLocationByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetPollerLocationCfgByID(id)
	if err != nil {
		log.Warningf("Error on get gloval variable %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetPollerLocationAffectOnDel --pending--
func GetPollerLocationAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetPollerLocationCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
