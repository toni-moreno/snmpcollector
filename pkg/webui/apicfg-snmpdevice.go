package webui

import (
	"github.com/go-macaron/binding"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"gopkg.in/macaron.v1"
)

func NewApiCfgSnmpDevice(m *macaron.Macaron) error {

	bind := binding.Bind

	// Data sources
	m.Group("/api/cfg/snmpdevice", func() {
		m.Get("/", reqSignedIn, GetSNMPDevices)
		m.Post("/", reqSignedIn, bind(config.SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", reqSignedIn, bind(config.SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", reqSignedIn, DeleteSNMPDevice)
		m.Get("/:id", reqSignedIn, GetSNMPDeviceByID)
		m.Get("/checkondel/:id", reqSignedIn, GetSNMPDevicesAffectOnDel)
	})

	return nil
}

// GetSNMPDevices Return snmpdevice list to frontend
func GetSNMPDevices(ctx *Context) {
	devcfgarray, err := agent.MainConfig.Database.GetSnmpDeviceCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Devices :%+s", err)
		return
	}
	ctx.JSON(200, &devcfgarray)
	log.Debugf("Getting DEVICEs %+v", &devcfgarray)
}

// AddSNMPDevice Insert new snmpdevice to de internal BBDD --pending--
func AddSNMPDevice(ctx *Context, dev config.SnmpDeviceCfg) {
	log.Printf("ADDING DEVICE %+v", dev)
	affected, err := agent.MainConfig.Database.AddSnmpDeviceCfg(dev)
	if err != nil {
		log.Warningf("Error on insert for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateSNMPDevice --pending--
func UpdateSNMPDevice(ctx *Context, dev config.SnmpDeviceCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateSnmpDeviceCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteSNMPDevice --pending--
func DeleteSNMPDevice(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelSnmpDeviceCfg(id)
	if err != nil {
		log.Warningf("Error on delete1 for device %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetSNMPDeviceByID --pending--
func GetSNMPDeviceByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetSnmpDeviceCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Device  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetSNMPDevicesAffectOnDel --pending--
func GetSNMPDevicesAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GeSnmpDeviceCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for SNMP metrics %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}
