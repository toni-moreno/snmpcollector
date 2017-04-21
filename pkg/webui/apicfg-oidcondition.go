package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

func NewApiCfgOidCondition(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/oidcondition", func() {
		m.Get("/", reqSignedIn, GetOidConditions)
		m.Post("/", reqSignedIn, bind(config.OidConditionCfg{}), AddOidCondition)
		m.Put("/:id", reqSignedIn, bind(config.OidConditionCfg{}), UpdateOidCondition)
		m.Delete("/:id", reqSignedIn, DeleteOidCondition)
		m.Get("/:id", reqSignedIn, GetOidConditionByID)
		m.Get("/checkondel/:id", reqSignedIn, GetOidConditionAffectOnDel)
	})
	return nil
}

// GetOidConditions Return metrics list to frontend
func GetOidConditions(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetOidConditionCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get OID contition :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting OID contitions %+v", &cfgarray)
}

// AddOidCondition Insert new condition to de internal BBDD --pending--
func AddOidCondition(ctx *Context, dev config.OidConditionCfg) {
	log.Printf("ADDING OidCondition %+v", dev)
	affected, err := agent.MainConfig.Database.AddOidConditionCfg(dev)
	if err != nil {
		log.Warningf("Error on insert OID condition %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMetric --pending--
func UpdateOidCondition(ctx *Context, dev config.OidConditionCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateOidConditionCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update OID Condition %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteOidCondition --pending--
func DeleteOidCondition(ctx *Context) {
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

//GetOidConditionByID --pending--
func GetOidConditionByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetOidConditionCfgByID(id)
	if err != nil {
		log.Warningf("Error on get OidCondition  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetOidConditionAffectOnDel --pending--
func GetOidConditionAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetOidConditionCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for OID conditions  %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
