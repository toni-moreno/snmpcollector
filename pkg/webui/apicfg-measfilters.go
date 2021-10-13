package webui

import (
	"fmt"

	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/filter"
	"gopkg.in/macaron.v1"
)

// NewAPICfgMeasFilters MeasFilter API REST creator
func NewAPICfgMeasFilters(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/cfg/measfilters", func() {
		m.Get("/", reqSignedIn, GetMeasFilter)
		m.Get("/:id", reqSignedIn, GetMeasFilterByID)
		m.Post("/", reqSignedIn, bind(config.MeasFilterCfg{}), AddMeasFilter)
		m.Put("/:id", reqSignedIn, bind(config.MeasFilterCfg{}), UpdateMeasFilter)
		m.Delete("/:id", reqSignedIn, DeleteMeasFilter)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasFiltersAffectOnDel)
	})

	return nil
}

/********************/
/*MEASUREMENT FILTERS
/********************/

// GetMeasFilter Return measurements groups list to frontend
func GetMeasFilter(ctx *Context) {
	// swagger:operation GET /cfg/measfilters Meas_Filter GetMeasFilter
	//---
	// summary: Get Measurement Filters from DB
	// description: Get All measurement filter config from DB
	// tags:
	// - "Measurement Filters Config"
	//
	// responses:
	//   '200':
	//     description: Measurement Filter Array
	//     schema:
	//       "$ref": "#/responses/idOfArrayMeasFilterResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	cfgarray, err := agent.MainConfig.Database.GetMeasFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

// GetMeasFilterByID --pending--
func GetMeasFilterByID(ctx *Context) {
	// swagger:operation GET /cfg/measfilters/{id} Meas_Filter GetMeasFilterByID
	//---
	// summary: Get Measurement Filter from DB
	// description: Get measurement filter config from  DB with specified ID
	// tags:
	// - "Measurement Filters Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The measurement Filter ID to retrieve
	//   required: true
	//   type: string
	//
	// responses:
	//   '200':
	//     description: Measurement Filter Array
	//     schema:
	//       "$ref": "#/definitions/MeasFilterCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetMeasFilterCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Filter  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddMeasFilter Insert new measurement groups to de internal BBDD --pending--
func AddMeasFilter(ctx *Context, dev config.MeasFilterCfg) {
	// swagger:operation POST /cfg/measfilters Meas_Filter AddMeasFilter
	//---
	// summary: Add Measurement Filter config to DB
	// description: Add  Measurement Filter config to the DB
	// tags:
	// - "Measurement Filters Config"
	//
	// parameters:
	// - name: MeasFilterCfg
	//   in: body
	//   description: Measurement Filter to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/MeasFilterCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MeasFilterCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	log.Printf("ADDING measurement Filter %+v", dev)
	// check Filter Config
	switch dev.FType {
	case "file":
		f := filter.NewFileFilter(dev.FilterName, dev.EnableAlias, log)
		err := f.Init(confDir)
		if err != nil {
			ctx.JSON(404, err.Error())
			return
		}
	case "OIDCondition":
		// no need for check here we have needed  SNMP walk function defined at this level
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
		// TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasFilter --pending--
func UpdateMeasFilter(ctx *Context, dev config.MeasFilterCfg) {
	// swagger:operation PUT /cfg/measfilters/{id} Meas_Filter UpdateMeasFilter
	//---
	// summary: Update Measurement Filter on the config DB
	// description: Update Measurement Filter  config with defined config data on the config DB
	// tags:
	// - "Measurement Filters Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement Filter ID to update
	//   required: true
	//   type: string
	// - name: MeasFilterCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/MeasFilterCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/MeasFilterCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMeasFilterCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

// DeleteMeasFilter --pending--
func DeleteMeasFilter(ctx *Context) {
	// swagger:operation DELETE /cfg/measfilters/{id} Meas_Filter DeleteMeasFilter
	//---
	// summary: Delete Measurement Filter
	// description: Delete  Measurement Filter with defined id
	// tags:
	// - "Measurement Filters Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement Filter ID to delete
	//   required: true
	//   type: string
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelMeasFilterCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Filter %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

// GetMeasFiltersAffectOnDel --pending--
func GetMeasFiltersAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/measfilters/checkondel/{id} Meas_Filter GetMeasFiltersAffectOnDel
	//---
	// summary: Get List for affected Objects on delete ID
	// description: Get List for affected Objects if deleting the measurement filter with ID
	// tags:
	// - "Measurement Filters Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The measurement Filter ID to check
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
	obarray, err := agent.MainConfig.Database.GetMeasFilterCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement filters %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
