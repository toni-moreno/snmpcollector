package webui

import (
	"fmt"
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/filter"
	"gopkg.in/macaron.v1"
)

func NewApiCfgMeasFilters(m *macaron.Macaron) error {

	bind := binding.Bind

	m.Group("/api/cfg/measfilters", func() {
		m.Get("/", reqSignedIn, GetMeasFilter)
		m.Post("/", reqSignedIn, bind(config.MeasFilterCfg{}), AddMeasFilter)
		m.Put("/:id", reqSignedIn, bind(config.MeasFilterCfg{}), UpdateMeasFilter)
		m.Delete("/:id", reqSignedIn, DeleteMeasFilter)
		m.Get("/:id", reqSignedIn, GetMeasFilterByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasFiltersAffectOnDel)
	})

	return nil
}

/********************/
/*MEASUREMENT FILTERS
/********************/

// GetMeasFilter Return measurements groups list to frontend
func GetMeasFilter(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetMeasFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

// AddMeasFilter Insert new measurement groups to de internal BBDD --pending--
func AddMeasFilter(ctx *Context, dev config.MeasFilterCfg) {
	log.Printf("ADDING measurement Filter %+v", dev)
	//check Filter Config
	switch dev.FType {
	case "file":
		f := filter.NewFileFilter(dev.FilterName, dev.EnableAlias, log)
		err := f.Init(confDir)
		if err != nil {
			ctx.JSON(404, err.Error())
			return
		}
	case "OIDCondition":
		//no need for check here we have needed  SNMP walk function defined at this level
	case "CustomFilter":
		f := filter.NewCustomFilter(dev.FilterName, dev.EnableAlias, log)
		err := f.Init(&agent.MainConfig.Database)
		if err != nil {
			ctx.JSON(404, err.Error())
			return
		}
	default:
		ctx.JSON(404, fmt.Errorf("Error no filter type %s supported ", dev.FType).Error())
		return
	}
	affected, err := agent.MainConfig.Database.AddMeasFilterCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasFilter --pending--
func UpdateMeasFilter(ctx *Context, dev config.MeasFilterCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMeasFilterCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeasFilter --pending--
func DeleteMeasFilter(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelMeasFilterCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Filter %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else { /****************/
		/*MEASUREMENT GROUPS
		  /****************/

		ctx.JSON(200, "deleted")
	}
}

//GetMeasFilterByID --pending--
func GetMeasFilterByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetMeasFilterCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Filter  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetMeasFiltersAffectOnDel --pending--
func GetMeasFiltersAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetMeasFilterCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement filters %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
