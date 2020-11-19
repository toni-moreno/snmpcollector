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
		m.Post("/", reqSignedIn, bind(config.SnmpDeviceCfg{}), AddSNMPDevice)
		m.Post("/:mode", reqSignedIn, bind(config.SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", reqSignedIn, bind(config.SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Put("/:id/:mode", reqSignedIn, bind(config.SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", reqSignedIn, DeleteSNMPDevice)
		m.Delete("/:id/:mode", reqSignedIn, DeleteSNMPDevice)
		m.Get("/:id", reqSignedIn, GetSNMPDeviceByID)
		m.Get("/checkondel/:id", reqSignedIn, GetSNMPDevicesAffectOnDel)
	})

	return nil
}

//DeviceStatMap
type DeviceStatMap struct {
	config.SnmpDeviceCfg
	IsRuntime bool
}

// GetSNMPDevices Return snmpdevice list to frontend
func GetSNMPDevices(ctx *Context) {
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

func addDeviceOnline(mode string, id string, dev *config.SnmpDeviceCfg) error {
	//First doing Ping
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

	//Next Adding to the Device Runtime
	agent.AddDeviceInRuntime(dev.ID, dev)
	return nil
}

// AddSNMPDevice Insert new snmpdevice to de internal BBDD --pending--
func AddSNMPDevice(ctx *Context, dev config.SnmpDeviceCfg) {
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
			//TODO: review if needed return data  or affected
			ctx.JSON(200, &dev)
		}
	}

}

// UpdateSNMPDevice --pending--
func UpdateSNMPDevice(ctx *Context, dev config.SnmpDeviceCfg) {
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
			//TODO: review if needed return device data
			ctx.JSON(200, &dev)
		}
	}

}

//DeleteSNMPDevice --pending--
func DeleteSNMPDevice(ctx *Context) {
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
		log.Debugf("Tying to delete device on database: %s", id)
		affected, err := agent.MainConfig.Database.DelSnmpDeviceCfg(id)
		if err != nil {
			log.Warningf("Error on delete1 for device %s  , affected : %+v , error: %s", id, affected, err)
			ctx.JSON(404, err.Error())
		} else {
			ctx.JSON(200, "deleted")
		}
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
