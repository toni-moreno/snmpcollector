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
	return
}
