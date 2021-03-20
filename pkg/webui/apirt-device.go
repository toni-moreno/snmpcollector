package webui

import (
	//"github.com/go-macaron/binding"
	"strconv"

	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"gopkg.in/macaron.v1"
)

// NewAPIRtDevice Runtime Device REST API creator
func NewAPIRtDevice(m *macaron.Macaron) error {

	//bind := binding.Bind

	m.Group("/api/rt/device", func() {
		m.Get("/info/", reqSignedIn, RTGetInfo)
		m.Get("/info/:id", reqSignedIn, RTGetInfo)
		m.Put("/status/activate/:id", reqSignedIn, RTActivateDev)
		m.Put("/status/deactivate/:id", reqSignedIn, RTDeactivateDev)
		m.Put("/debug/activate/:id", reqSignedIn, RTActSnmpDebugDev)
		m.Put("/debug/deactivate/:id", reqSignedIn, RTDeactSnmpDebugDev)
		m.Get("/snmpreset/:id/:mode", reqSignedIn, RTSnmpReset)
		m.Get("/forcegather/:id", reqSignedIn, RTForceGather)
		m.Put("/log/setloglevel/:id/:level", reqSignedIn, RTSetLogLevelDev)
		m.Get("/log/getdevicelog/:id", reqSignedIn, RTGetLogFileDev)
		m.Get("/filter/forcefltupdate/:id", reqSignedIn, RTForceFltUpdate)
		m.Get("/snmpmaxrep/:id/:maxrep", reqSignedIn, RTSnmpSetMaxRep)
	})

	return nil
}

/****************/
/*Runtime Info
/****************/

// RTSnmpSetMaxRep runtime set max repeticions
func RTSnmpSetMaxRep(ctx *Context) {
	// swagger:operation GET /rt/device/snmpmaxrep/{id}/{maxrep} Runtime_Devices RTSnmpSetMaxRep
	//---
	// summary: Set SNMP Maxrepetitions for a device
	// description: Set SNMP Maxrepetitions for a device
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID
	//   required: true
	//   type: string
	// - name: maxrep
	//   in: path
	//   description: num of max repetitions
	//   required: true
	//   type: integer
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
	maxrep := ctx.Params(":maxrep")
	d, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("set maxrepetitions for snmp device %s", id)
	i, _ := strconv.Atoi(maxrep)
	d.RTActSnmpMaxRep(uint8(i))
	ctx.JSON(200, "OK")
}

// RTForceFltUpdate xx
func RTForceFltUpdate(ctx *Context) {
	// swagger:operation GET /rt/device/filter/forcefltupdate/{id} Runtime_Devices RTForceFltUpdate
	//---
	// summary: Force Updating filters
	// description: Force Updating filters on device in next gathering period
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to force
	//   required: true
	//   type: string
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
	d, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("trying to force filter for device %s", id)
	d.ForceFltUpdate()
	ctx.JSON(200, "OK")
}

// RTGetLogFileDev get file dev
func RTGetLogFileDev(ctx *Context) {
	// swagger:operation GET /rt/device/log/getdevicelog/{id} Runtime_Devices RTGetLogFileDev
	//---
	// summary: Download Device Log
	// description: Download Log File for device specified by ID
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to force
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       type: file
	//     headers:
	//        Content-Disposition:
	//           type: string
	//           description: the value is `attachment; filename="{id}.log"`

	id := ctx.Params(":id")
	d, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	ctx.ServeFile(d.GetLogFilePath())
}

//RTSetLogLevelDev xx
func RTSetLogLevelDev(ctx *Context) {
	// swagger:operation PUT /rt/device/log/setloglevel/{id}/{level} Runtime_Devices RTSetLogLevelDev
	//---
	// summary: Set SNMP Maxrepetitions for a device
	// description: Set SNMP Maxrepetitions for a device
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID
	//   required: true
	//   type: string
	// - name: level
	//   in: path
	//   description: Level
	//   required: true
	//   type: string
	//   enum: [error,warn,info,debug,trace]
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
	level := ctx.Params(":level")
	dev, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("set runtime log level from device id %s : %s", id, level)
	dev.RTSetLogLevel(level)
	ctx.JSON(200, "OK")

}

// RTSnmpReset runtime send reset
func RTSnmpReset(ctx *Context) {
	// swagger:operation GET /rt/device/snmpreset/{id}/{mode} Runtime_Devices RTSnmpReset
	//---
	// summary: Reset all SNMP connections currently stablished with remote device
	// description: Reset all SNMP connections currently stablished with remote device
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to reset
	//   required: true
	//   type: string
	// - name: mode
	//   in: path
	//   description: you can reset in hard or soft mode.
	//   required: true
	//   type: string
	//   enum: [hard,soft]
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
	mode := ctx.Params(":mode")
	log.Infof("activating runtime on device %s", id)
	dev, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("activating runtime on device %s", id)
	dev.SnmpReset(mode)
	ctx.JSON(200, "OK")
}

// RTForceGather force gather
func RTForceGather(ctx *Context) {
	// swagger:operation GET /rt/device/forcegather/{id} Runtime_Devices RTForceGather
	//---
	// summary: Launch a complete cicle of gathering data, even though device was not active
	// description: Launch a complete cicle of gathering data, even though device was not active
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to reset
	//   required: true
	//   type: string
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
	log.Infof("activating runtime on device %s", id)
	dev, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("activating runtime on device %s", id)
	dev.ForceGather()
	ctx.JSON(200, "OK")
}

//RTActivateDev xx
func RTActivateDev(ctx *Context) {
	// swagger:operation PUT /rt/device/status/activate/{id} Runtime_Devices RTActivateDev
	//---
	// summary: Activate Device
	// description: Activate Device
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to reset
	//   required: true
	//   type: string
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
	dev, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("activating runtime on device %s", id)
	dev.RTActivate(true)
	ctx.JSON(200, "OK")
}

//RTDeactivateDev xx
func RTDeactivateDev(ctx *Context) {
	// swagger:operation PUT /rt/device/status/deactivate/{id} Runtime_Devices RTDeactivateDev
	//---
	// summary: Deactivate Device
	// description: Deactivate Device
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to reset
	//   required: true
	//   type: string
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
	dev, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("deactivating runtime on device  %s", id)
	dev.RTActivate(false)
	ctx.JSON(200, "OK")

}

//RTActSnmpDebugDev xx
func RTActSnmpDebugDev(ctx *Context) {
	// swagger:operation PUT /rt/device/debug/activate/{id} Runtime_Devices RTActSnmpDebugDev
	//---
	// summary: Activate SNMP Debugging
	// description: Activate SNMP Debugging (generate snmp packet tracing in one extra log file)
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to reset
	//   required: true
	//   type: string
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
	dev, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("activating snmpdebug  %s", id)
	dev.RTActSnmpDebug(true)
	ctx.JSON(200, "OK")
}

//RTDeactSnmpDebugDev xx
func RTDeactSnmpDebugDev(ctx *Context) {
	// swagger:operation PUT /rt/device/debug/deactivate/{id} Runtime_Devices RTDeactSnmpDebugDev
	//---
	// summary: Deactivate SNMP Debugging
	// description: Deactivate SNMP Debugging (stop snmp packet tracing in the extra log file)
	// tags:
	// - "Runtime Device"
	//
	// parameters:
	// - name: id
	//   in: path
	//   description: Device ID to reset
	//   required: true
	//   type: string
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
	dev, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Infof("deactivating snmpdebug  %s", id)
	dev.RTActSnmpDebug(false)
	ctx.JSON(200, "OK")
}

//RTGetInfo xx
func RTGetInfo(ctx *Context) {
	// swagger:operation GET /rt/device/info Runtime_Devices RTGetInfo
	//---
	// summary: Get Device Status info
	// description: Get Device Status info
	// tags:
	// - "Runtime Device"
	//
	// responses:
	//   '200':
	//     description: "OK"
	//     schema:
	//       "$ref": "#/responses/idOfDeviceStatResp"
	//   '404':
	//     description: unexpected error
	//     schema:
	//       "$ref": "#/responses/idOfStringResp"

	id := ctx.Params(":id")
	if len(id) > 0 {
		json, err := agent.GetDeviceJSONInfo(id)
		if err != nil {
			ctx.JSON(404, err.Error())
			return
		}

		log.Infof("get runtime data from id %s", id)
		ctx.RawAsJSON(200, json)

		//get only one device info
	} else {
		devstats := agent.GetDevStats()
		ctx.JSON(200, &devstats)
	}
}
