package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgSnmpMetric SnmpMetric  API REST creator
func NewAPICfgSnmpMetric(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/metric", func() {
		m.Get("/", reqSignedIn, GetMetrics)
		m.Post("/", reqSignedIn, bind(config.SnmpMetricCfg{}), AddMetric)
		m.Put("/:id", reqSignedIn, bind(config.SnmpMetricCfg{}), UpdateMetric)
		m.Delete("/:id", reqSignedIn, DeleteMetric)
		m.Get("/:id", reqSignedIn, GetMetricByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMetricsAffectOnDel)
		m.Post("/convmodes", reqSignedIn, bind(config.SnmpMetricCfg{}), GetConversionModes)
	})

	return nil
}

// GetMetrics Return metrics list to frontend
func GetMetrics(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetSnmpMetricCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Metrics :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Metrics %+v", &cfgarray)
}

// AddMetric Insert new metric to de internal BBDD --pending--
func AddMetric(ctx *Context, dev config.SnmpMetricCfg) {
	log.Printf("ADDING Metric %+v", dev)
	affected, err := agent.MainConfig.Database.AddSnmpMetricCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMetric --pending--
func UpdateMetric(ctx *Context, dev config.SnmpMetricCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateSnmpMetricCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMetric --pending--
func DeleteMetric(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelSnmpMetricCfg(id)
	if err != nil {
		log.Warningf("Error on delete Metric %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMetricByID --pending--
func GetMetricByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetSnmpMetricCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Metric  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetMetricsAffectOnDel --pending--
func GetMetricsAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetSnmpMetricCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for SNMP metrics %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

type ConversionItem struct {
	ID      int
	Value   string
	Default bool
}

// GetConversionModes Return conversion modes from datasource Type
func GetConversionModes(ctx *Context, dev config.SnmpMetricCfg) {
	var response []ConversionItem
	cfgarray, def, err := dev.GetValidConversions()
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get  Conversion Mode :%s", err)
		return
	}
	for k, v := range cfgarray {
		response = append(response, ConversionItem{ID: int(v), Value: v.GetString(), Default: (k == def)})
	}
	ctx.JSON(200, &response)
	log.Debugf("Got Conversion Items Array Metrics %+v", &response)
}
