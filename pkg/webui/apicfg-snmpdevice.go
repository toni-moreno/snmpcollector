package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"gopkg.in/macaron.v1"
)

// NewAPICfgSnmpDevice SnmpDevice REST API creator
func NewAPICfgSnmpDevice(m *macaron.Macaron) error {
	bind := binding.Bind

	// Data sources
	m.Group("/api/cfg/snmpdevice", func() {
		m.Get("/", reqSignedIn, GetSNMPDevices)
		m.Get("/:id", reqSignedIn, GetSNMPDeviceByID)
		m.Post("/", reqSignedIn, bind(config.SnmpDeviceCfg{}), AddSNMPDevice)
		m.Post("/:mode", reqSignedIn, bind(config.SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", reqSignedIn, bind(config.SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Put("/:id/:mode", reqSignedIn, bind(config.SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", reqSignedIn, DeleteSNMPDevice)
		m.Delete("/:id/:mode", reqSignedIn, DeleteSNMPDevice)
		m.Get("/checkondel/:id", reqSignedIn, GetSNMPDevicesAffectOnDel)
	})

	return nil
}

// DeviceStatMap
// swagger:model DeviceStatMap
type DeviceStatMap struct {
	config.SnmpDeviceCfg
	IsRuntime bool
}

// GetSNMPDevices Return snmpdevice list to frontend
func GetSNMPDevices(ctx *Context) {
	// swagger:operation GET /cfg/snmpdevice  Config_Device GetSNMPDevices
	//---
	// summary: Get All devices info from DB and Runtime
	// description: Get All Devices config info as an array of config and boolean if working in runtime.
	// tags:
	// - "Devices Config"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfArrayDeviceStatResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	devcfgarray, err := agent.MainConfig.Database.GetSnmpDeviceCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Devices :%+s", err)
		return
	}

	dsmap := []*DeviceStatMap{}
	for _, v := range devcfgarray {
		rt := agent.IsDeviceInRuntime(v.ID)
		dsmap = append(dsmap, &DeviceStatMap{*v, rt})
	}
	ctx.JSON(200, &dsmap)
	log.Debugf("Getting DEVICEs %+v", &dsmap)
}

// GetSNMPDeviceByID --pending--
func GetSNMPDeviceByID(ctx *Context) {
	// swagger:operation GET /cfg/snmpdevice/{id}  Config_Device GetSNMPDeviceByID
	//---
	// summary: Get devices config from DB
	// description: Get Devicesconfig info from DB specified by ID
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to get data
	//   required: true
	//   type: string
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetSnmpDeviceCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Device  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

func addDeviceOnline(mode string, id string, dev *config.SnmpDeviceCfg) error {
	// First doing Ping
	log.Infof("trying to ping device %s : %+v", dev.ID, dev)

	_, sysinfo, err := snmp.GetClient(dev, log, "ping", false, 0)
	if err != nil {
		log.Debugf("ERROR  on query device : %s", err)
		return err
	}
	log.Infof("Device Ping ok : %#v", sysinfo)
	// Next updating database
	switch mode {
	case "add":
		affected, err := agent.MainConfig.Database.AddSnmpDeviceCfg(*dev)
		if err != nil {
			log.Warningf("Error on insert for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
			return err
		}
	case "update":
		affected, err := agent.MainConfig.Database.UpdateSnmpDeviceCfg(id, *dev)
		if err != nil {
			log.Warningf("Error on insert for device %s  , affected : %+v , error: %s", id, affected, err)
			return err
		}
	default:
	}

	// Next Adding to the Device Runtime
	agent.AddDeviceInRuntime(dev.ID, dev)
	return nil
}

// AddSNMPDevice Insert new snmpdevice to de internal BBDD --pending--
func AddSNMPDevice(ctx *Context, dev config.SnmpDeviceCfg) {
	// swagger:operation POST /cfg/snmpdevice Config_Device AddSNMPDevice1
	//---
	// summary: Add a new Device into de config database and/or in runtime.
	// description: Add a new Device into de config database and/or in runtime.
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: SnmpDeviceCfg
	//   in: body
	//   description: device to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//
	// responses:
	//   '200':
	//     description: Added Device config
	//     schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	// swagger:operation POST /cfg/snmpdevice/{mode} Config_Device AddSNMPDevice2
	//---
	// summary: Add a new Device into de config database and/or in runtime.
	// description: |
	//   Add a new  existing Device into de config database with specified ID and/or reload new config in runtime.
	//   Modes:
	//    - "config": Only  in config database (equivalent to delete without mode parameter)
	//    - "runtime":  Only e in active and running devices (runtime) WARN: this config will be lost on next reload.
	//    - "full": in both on database and also in runtime devices.
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: mode
	//   in: path
	//   description: Adition mode
	//   required: true
	//   type: string
	//   enum: [runtime,full,config]
	//
	// - name: SnmpDeviceCfg
	//   in: body
	//   description: device to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//
	// responses:
	//   '200':
	//     description: Added Device config
	//     schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	mode := ctx.Params(":mode")
	log.Printf("ADDING DEVICE %+v mode(%s)", dev, mode)
	switch mode {
	case "runtime":
		err := addDeviceOnline("deploy", dev.ID, &dev)
		if err != nil {
			log.Warningf("Error on insert for device %s  , error: %s", dev.ID, err)
			ctx.JSON(404, err.Error())
		} else {
			ctx.JSON(200, &dev)
		}
	case "full":
		err := addDeviceOnline("add", dev.ID, &dev)
		if err != nil {
			log.Warningf("Error on insert for device %s  , error: %s", dev.ID, err)
			ctx.JSON(404, err.Error())
		} else {
			ctx.JSON(200, &dev)
		}
	default:
		affected, err := agent.MainConfig.Database.AddSnmpDeviceCfg(dev)
		if err != nil {
			log.Warningf("Error on insert for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
			ctx.JSON(404, err.Error())
		} else {
			// TODO: review if needed return data  or affected
			ctx.JSON(200, &dev)
		}
	}
}

// UpdateSNMPDevice --pending--
func UpdateSNMPDevice(ctx *Context, dev config.SnmpDeviceCfg) {
	// swagger:operation PUT /cfg/snmpdevice/{id} Config_Device UpdateSNMPDevice1
	//---
	// summary: Update an existing Device into de config database
	// description: Update an existing Device into de config database with specified ID
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The device ID to update
	//   required: true
	//   type: string
	//
	// - name: SnmpDeviceCfg
	//   in: body
	//   description: device to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//
	// responses:
	//   '200':
	//     description: Device Config
	//     schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	// swagger:operation PUT /cfg/snmpdevice/{id}/{mode} Config_Device UpdateSNMPDevice2
	//---
	// summary: Update an existing Device into de config database and/or reload new config in runtime.
	// description: |
	//   Update an existing Device into de config database with specified ID and/or reload new config in runtime.
	//   Modes:
	//    - "config": Only update in config database (equivalent to delete without mode parameter)
	//    - "runtime":  Only update in active and running devices (runtime) WARN: this config will be lost on next reload.
	//    - "full": Update on database and also in runtime
	//
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The device ID to update
	//   required: true
	//   type: string
	// - name: mode
	//   in: path
	//   description: SNMP Get type
	//   required: true
	//   type: string
	//   enum: [runtime,full,config]
	//
	// - name: SnmpDeviceCfg
	//   in: body
	//   description: device to add
	//   required: true
	//   schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//
	// responses:
	//   '200':
	//     description: Device Config
	//     schema:
	//       "$ref": "#/definitions/SnmpDeviceCfg"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	mode := ctx.Params(":mode")
	log.Printf("UPDATING DEVICE %s in mode(%s)", id, mode)
	switch mode {
	case "runtime":
		var err error
		err = agent.DeleteDeviceInRuntime(id)
		if err != nil {
			log.Warningf("Error on online delete device %s  , error %s", dev.ID, err)
			ctx.JSON(404, err.Error())
			return
		}
		err = addDeviceOnline("deploy", id, &dev)
		if err != nil {
			log.Warningf("Error on insert for device %s  , error: %s", dev.ID, err)
			ctx.JSON(404, err.Error())
		} else {
			ctx.JSON(200, &dev)
		}
	case "full":
		var err error
		err = agent.DeleteDeviceInRuntime(id)
		if err != nil {
			log.Warningf("Error on online delete device %s  , error %s", dev.ID, err)
			ctx.JSON(404, err.Error())
			return
		}
		err = addDeviceOnline("update", id, &dev)
		if err != nil {
			log.Warningf("Error on insert for device %s  , error: %s", dev.ID, err)
			ctx.JSON(404, err.Error())
		} else {
			ctx.JSON(200, &dev)
		}
	default:
		log.Debugf("Tying to update device  %s on  database: %+v", id, dev)
		affected, err := agent.MainConfig.Database.UpdateSnmpDeviceCfg(id, dev)
		if err != nil {
			log.Warningf("Error on update for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
			ctx.JSON(404, err.Error())
		} else {
			// TODO: review if needed return device data
			ctx.JSON(200, &dev)
		}
	}
}

// DeleteSNMPDevice --pending--
func DeleteSNMPDevice(ctx *Context) {
	// swagger:operation DELETE /cfg/snmpdevice/{id} Config_Device DeleteSNMPDevice1
	//---
	// summary: Delete existing Device into de config database
	// description: Delete an existing Device into de config database with specified ID
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The device ID to update
	//   required: true
	//   type: string
	//
	// responses:
	//   '200':
	//     description: OK Response
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	//

	// swagger:operation DELETE /cfg/snmpdevice/{id}/{mode} Config_Device DeleteSNMPDevice2
	//---
	// summary: Delete an existing Device into de config database and/or reload new config in runtime.
	// description: |
	//   Delete an existing Device into de config database with specified ID and/or reload new config in runtime.
	//   Modes:
	//    - "config": Only update in config database (equivalent to delete without mode parameter)
	//    - "runtime":  Only update in active and running devices (runtime) WARN: this config will be lost on next reload.
	//    - "full": Update on database and also in runtime
	//
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The device ID to update
	//   required: true
	//   type: string
	// - name: mode
	//   in: path
	//   description: SNMP Get type
	//   required: true
	//   type: string
	//   enum: [runtime,full]
	//
	// responses:
	//   '200':
	//     description: OK response
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	mode := ctx.Params(":mode")
	log.Printf("DELETING DEVICE %s in mode(%s)", id, mode)
	switch mode {
	case "runtime":
		err := agent.DeleteDeviceInRuntime(id)
		if err != nil {
			log.Warningf("Error on online delete device %s  , error %s", id, err)
			ctx.JSON(404, err.Error())
			return
		} else {
			ctx.JSON(200, "deleted")
		}
	case "full":
		err := agent.DeleteDeviceInRuntime(id)
		if err != nil {
			log.Warningf("Error on online delete device %s  , error %s", id, err)
			ctx.JSON(404, err.Error())
			return
		}
		fallthrough
	default:
		log.Debugf("Trying to delete device on database: %s", id)
		affected, err := agent.MainConfig.Database.DelSnmpDeviceCfg(id)
		if err != nil {
			log.Warningf("Error on delete1 for device %s  , affected : %+v , error: %s", id, affected, err)
			ctx.JSON(404, err.Error())
		} else {
			ctx.JSON(200, "deleted")
		}
	}
}

// GetSNMPDevicesAffectOnDel --pending--
func GetSNMPDevicesAffectOnDel(ctx *Context) {
	// swagger:operation GET /cfg/snmpdevice/checkondel/{id} Config_Device GetSNMPDevicesAffectOnDel
	//---
	// summary: Get List for affected Objects on delete ID
	// description: Get List for affected Objects if deleting the  Device with ID
	// tags:
	// - "Devices Config"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: The device ID to check
	//   required: true
	//   type: string
	//
	// responses:
	//   '200':
	//     description: List for Affected Items
	//     schema:
	//       "$ref": "#/responses/idOfCheckOnDelResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GeSnmpDeviceCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for SNMP metrics %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
