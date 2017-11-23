package webui

import (
	"time"

	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/agent/output"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgInfluxServer InfluxServer API REST creator
func NewAPICfgInfluxServer(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/influxservers", func() {
		m.Get("/", reqSignedIn, GetInfluxServer)
		m.Post("/", reqSignedIn, bind(config.InfluxCfg{}), AddInfluxServer)
		m.Put("/:id", reqSignedIn, bind(config.InfluxCfg{}), UpdateInfluxServer)
		m.Delete("/:id", reqSignedIn, DeleteInfluxServer)
		m.Get("/:id", reqSignedIn, GetInfluxServerByID)
		m.Get("/checkondel/:id", reqSignedIn, GetInfluxAffectOnDel)
		m.Post("/ping/", reqSignedIn, bind(config.InfluxCfg{}), PingInfluxServer)
	})

	return nil
}

// GetInfluxServer Return Server Array
func GetInfluxServer(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetInfluxCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx db :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// AddInfluxServer Insert new measurement groups to de internal BBDD --pending--
func AddInfluxServer(ctx *Context, dev config.InfluxCfg) {
	log.Printf("ADDING Influx Backend %+v", dev)
	affected, err := agent.MainConfig.Database.AddInfluxCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Backend %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateInfluxServer --pending--
func UpdateInfluxServer(ctx *Context, dev config.InfluxCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateInfluxCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Influx db %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteInfluxServer --pending--
func DeleteInfluxServer(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelInfluxCfg(id)
	if err != nil {
		log.Warningf("Error on delete influx db %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetInfluxServerByID --pending--
func GetInfluxServerByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetInfluxCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Influx db data for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetInfluxAffectOnDel --pending--
func GetInfluxAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetInfluxCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

//PingInfluxServer Return ping result
func PingInfluxServer(ctx *Context, cfg config.InfluxCfg) {
	log.Infof("trying to ping influx server %s : %+v", cfg.ID, cfg)
	_, elapsed, message, err := output.Ping(&cfg)
	type result struct {
		Result  string
		Elapsed time.Duration
		Message string
	}
	if err != nil {
		log.Debugf("ERROR on ping InfluxDB Server : %s", err)
		res := result{Result: "NOOK", Elapsed: elapsed, Message: err.Error()}
		ctx.JSON(400, res)
	} else {
		log.Debugf("OK on ping InfluxDB Server %+v, %+v", elapsed, message)
		res := result{Result: "OK", Elapsed: elapsed, Message: message}
		ctx.JSON(200, res)
	}
}
