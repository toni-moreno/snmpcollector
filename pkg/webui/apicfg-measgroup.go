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
		m.Post("/", reqSignedIn, bind(config.MGroupsCfg{}), AddMeasGroup)
		m.Put("/:id", reqSignedIn, bind(config.MGroupsCfg{}), UpdateMeasGroup)
		m.Delete("/:id", reqSignedIn, DeleteMeasGroup)
		m.Get("/:id", reqSignedIn, GetMeasGroupByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasGroupsAffectOnDel)
	})

	return nil
}

// GetMeasGroup Return measurements groups list to frontend
func GetMeasGroup(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetMGroupsCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Group :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Meas Group %+v", &cfgarray)
}

// AddMeasGroup Insert new measurement groups to de internal BBDD --pending--
func AddMeasGroup(ctx *Context, dev config.MGroupsCfg) {
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

//GetMeasGroupByID --pending--
func GetMeasGroupByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetMGroupsCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Group for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetMeasGroupsAffectOnDel --pending--
func GetMeasGroupsAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetMGroupsCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement Groups %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
