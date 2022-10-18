package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

// NewAPICfgKafkaServer KafkaServer API REST creator
func NewAPICfgKafkaServer(m *macaron.Macaron) error {
	bind := binding.Bind

	m.Group("/api/cfg/kafkaservers", func() {
		m.Get("/", reqSignedIn, GetKafkaServer)
		m.Get("/:id", reqSignedIn, GetKafkaServerByID)
		m.Post("/", reqSignedIn, bind(config.KafkaCfg{}), AddKafkaServer)
		m.Put("/:id", reqSignedIn, bind(config.KafkaCfg{}), UpdateKafkaServer)
		m.Delete("/:id", reqSignedIn, DeleteKafkaServer)
		m.Get("/checkondel/:id", reqSignedIn, GetKafkaAffectOnDel)
	})

	return nil
}

// GetKafkaServer Return Server Array
func GetKafkaServer(ctx *Context) {
	// swagger:operation GET /cfg/kafkaservers  Config_KafkaServers GetKafkaServer
	//---
	// summary: Get All Kafka Servers Config Items from DB
	// description: Get All Kafka Servers config Items as an array from DB
	// tags:
	// - "Kafka Servers Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayKafkaCfgResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	cfgarray, err := agent.MainConfig.Database.GetKafkaCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Kafka db :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// GetKafkaServerByID --pending--
func GetKafkaServerByID(ctx *Context) {
	// swagger:operation GET /cfg/kafkaservers/{id}  Config_KafkaServers GetKafkaServerByID
	//---
	// summary: Get KafkaServer Config from DB
	// description: Get KafkaServers config info by ID from DB
	// tags:
	// - "Kafka Servers Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: KafkaServer to get
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/KafkaCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetKafkaCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Kafka db data for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

// AddKafkaServer Insert new measurement groups to de internal BBDD --pending--
func AddKafkaServer(ctx *Context, dev config.KafkaCfg) {
	// swagger:operation POST /cfg/kafkaservers Config_KafkaServers AddKafkaServer
	//---
	// summary: Add new Kafka Server Config
	// description: Add KafkaServer from Data
	// tags:
	// - "Kafka Servers Config"
	//
	// parameters:
	// - name: KafkaCfg
	//   in: body
	//   description: KafkaConfig to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/KafkaCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/KafkaCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	log.Printf("ADDING Kafka Backend %+v", dev)
	affected, err := agent.MainConfig.Database.AddKafkaCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Backend %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		// TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateKafkaServer --pending--
func UpdateKafkaServer(ctx *Context, dev config.KafkaCfg) {
	// swagger:operation PUT /cfg/kafkaservers/{id} Config_KafkaServers UpdateKafkaServer
	//---
	// summary: Update Kafka Server Config
	// description: Update KafkaServer from Data with specified ID
	// tags:
	// - "Kafka Servers Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Kafka Config ID to update
	//   required: true
	//   type: string
	// - name: KafkaCfg
	//   in: body
	//   description: Metric to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/KafkaCfg"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/KafkaCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateKafkaCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Kafka db %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		// TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

// DeleteKafkaServer --pending--
func DeleteKafkaServer(ctx *Context) {
	// swagger:operation DELETE /cfg/kafkaservers/{id} Config_KafkaServers DeleteKafkaServer
	//---
	// summary: Delete Kafka Server Config on DB
	// description: Delete Kafka Server on DB with specified ID
	// tags:
	// - "Kafka Servers Config"
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
	affected, err := agent.MainConfig.Database.DelKafkaCfg(id)
	if err != nil {
		log.Warningf("Error on delete influx db %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

// GetKafkaAffectOnDel --pending--
func GetKafkaAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/kafkaservers/checkondel/{id} Config_KafkaServers GetKafkaAffectOnDel
	//---
	// summary: Check affected sources.
	// description: Get all existing Objects affected when deleted the KafkaServer.
	// tags:
	// - "Kafka Servers Config"
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
	obarray, err := agent.MainConfig.Database.GetKafkaCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
