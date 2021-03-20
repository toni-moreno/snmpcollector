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
		m.Get("/:id", reqSignedIn, GetInfluxServerByID)
		m.Post("/", reqSignedIn, bind(config.InfluxCfg{}), AddInfluxServer)
		m.Put("/:id", reqSignedIn, bind(config.InfluxCfg{}), UpdateInfluxServer)
		m.Delete("/:id", reqSignedIn, DeleteInfluxServer)
		m.Get("/checkondel/:id", reqSignedIn, GetInfluxAffectOnDel)
		m.Post("/ping/", reqSignedIn, bind(config.InfluxCfg{}), PingInfluxServer)
	})

	return nil
}

// GetInfluxServer Return Server Array
func GetInfluxServer(ctx *Context) {
	// swagger:operation GET /cfg/influxservers  Config_InfluxServers GetInfluxServer
	//---
	// summary: Get All Influx Servers Config Items from DB
	// description: Get All Influx Servers config Items as an array from DB
	// tags:
	// - "Influx Servers Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayInfluxCfgResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	cfgarray, err := agent.MainConfig.Database.GetInfluxCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx db :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

//GetInfluxServerByID --pending--
func GetInfluxServerByID(ctx *Context) {
	// swagger:operation GET /cfg/influxservers/{id}  Config_InfluxServers GetInfluxServerByID
	//---
	// summary: Get InfluxServer Config from DB
	// description: Get InfluxServers config info by ID from DB
	// tags:
	// - "Influx Servers Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: InfluxServer to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/InfluxCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetInfluxCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Influx db data for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddInfluxServer Insert new measurement groups to de internal BBDD --pending--
func AddInfluxServer(ctx *Context, dev config.InfluxCfg) {
	// swagger:operation POST /cfg/influxservers Config_InfluxServers AddInfluxServer
	//---
	// summary: Add new Influx Server Config
	// description: Add InfluxServer from Data
	// tags:
	// - "Influx Servers Config"
	//
	// parameters:
	// - name: InfluxCfg
	//   in: body
	//   description: InfluxConfig to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/InfluxCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/InfluxCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
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
	// swagger:operation PUT /cfg/influxservers/{id} Config_InfluxServers UpdateInfluxServer
	//---
	// summary: Update Influx Server Config
	// description: Update InfluxServer from Data with specified ID
	// tags:
	// - "Influx Servers Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Influx Config ID to update
	//   required: true
	//   type: string
	// - name: InfluxCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/InfluxCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/InfluxCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
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
	// swagger:operation DELETE /cfg/influxservers/{id} Config_InfluxServers DeleteInfluxServer
	//---
	// summary: Delete Influx Server Config on DB
	// description: Delete Influx Server on DB with specified ID
	// tags:
	// - "Influx Servers Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Measurement Group ID to delete
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
	affected, err := agent.MainConfig.Database.DelInfluxCfg(id)
	if err != nil {
		log.Warningf("Error on delete influx db %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetInfluxAffectOnDel --pending--
func GetInfluxAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/influxservers/checkondel/{id} Config_InfluxServers GetInfluxAffectOnDel
	//---
	// summary: Check affected sources.
	// description: Get all existing Objects affected when deleted the InfluxServer.
	// tags:
	// - "Influx Servers Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The measurement Group ID to check
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
	// swagger:operation POST /cfg/influxservers/ping Config_InfluxServers PingInfluxServer
	//---
	// summary: Connection Test (Ping) to the Influx Server
	// description: Performs a Test Connection to the InfluxServer With specified Config in the Body
	// tags:
	// - "Influx Servers Config"
	//
	// parameters:
	// - name: InfluxCfg
	//   in: body
	//   description: InfluxConfig to ping
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/InfluxCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/InfluxCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

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
