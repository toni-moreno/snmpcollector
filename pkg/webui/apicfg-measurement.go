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
		m.Get("/type/:type", reqSignedIn, GetMeasByType)
		m.Post("/", reqSignedIn, bind(config.MeasurementCfg{}), AddMeas)
		m.Put("/:id", reqSignedIn, bind(config.MeasurementCfg{}), UpdateMeas)
		m.Delete("/:id", reqSignedIn, DeleteMeas)
		m.Get("/:id", reqSignedIn, GetMeasByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasAffectOnDel)
	})

	return nil
}

/****************/
/*INFLUX MEASUREMENTS
/****************/

// GetMeas Return measurements list to frontend
func GetMeas(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetMeasurementCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx Measurements :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurements %+v", &cfgarray)
}

// GetMeasByType Return measurements list to frontend
func GetMeasByType(ctx *Context) {
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
	log.Printf("ADDING Measurement %+v", dev)
	affected, err := agent.MainConfig.Database.AddMeasurementCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeas --pending--
func UpdateMeas(ctx *Context, dev config.MeasurementCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMeasurementCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeas --pending--
func DeleteMeas(ctx *Context) {
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

//GetMeasByID --pending--
func GetMeasByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetMeasurementCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetMeasAffectOnDel --pending--
func GetMeasAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetMeasurementCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurements %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
