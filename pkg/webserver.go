package main

import (
	"fmt"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
	"html/template"
	"net/http"
)

//HTTPConfig has webserver config options
type HTTPConfig struct {
	Port          int    `toml:"port"`
	AdminUser     string `toml:"adminuser"`
	AdminPassword string `toml:"adminpassword"`
}

//UserLogin for login purposes
type UserLogin struct {
	UserName string `form:"username" binding:"Required"`
	Password string `form:"password" binding:"Required"`
}

func webServer(port int) {

	bind := binding.Bind

	/*	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("My Secret"), nil
		},
		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodHS256,
	})*/
	// initiate the app
	m := macaron.Classic()

	// register middleware
	m.Use(macaron.Recovery())
	//	m.Use(gzip.Gziper())
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			// Prefix is the optional prefix used to serve the static directory content. Default is empty string.
			Prefix: "public",
			// SkipLogging will disable [Static] log messages when a static file is served. Default is false.
			SkipLogging: true,
			// IndexFile defines which file to serve as index if it exists. Default is "index.html".
			IndexFile: "index.html",
			// Expires defines which user-defined function to use for producing a HTTP Expires Header. Default is nil.
			// https://developers.google.com/speed/docs/insights/LeverageBrowserCaching
			Expires: func() string { return "max-age=0" },
		}))
	m.Use(session.Sessioner(session.Options{
		// Name of provider. Default is "memory".
		Provider: "memory",
		// Provider configuration, it's corresponding to provider.
		ProviderConfig: "",
		// Cookie name to save session ID. Default is "MacaronSession".
		CookieName: "MacaronSession",
		// Cookie path to store. Default is "/".
		CookiePath: "/",
		// GC interval time in seconds. Default is 3600.
		Gclifetime: 3600,
		// Max life time in seconds. Default is whatever GC interval time is.
		Maxlifetime: 3600,
		// Use HTTPS only. Default is false.
		Secure: false,
		// Cookie life time. Default is 0.
		CookieLifeTime: 0,
		// Cookie domain name. Default is empty.
		Domain: "",
		// Session ID length. Default is 16.
		IDLength: 16,
		// Configuration section name. Default is "session".
		Section: "session",
	}))

	m.Use(macaron.Renderer(macaron.RenderOptions{
		// Directory to load templates. Default is "templates".
		Directory: "pkg/templates",
		// Extensions to parse template files from. Defaults are [".tmpl", ".html"].
		Extensions: []string{".tmpl", ".html"},
		// Funcs is a slice of FuncMaps to apply to the template upon compilation. Default is [].
		Funcs: []template.FuncMap{map[string]interface{}{
			"AppName": func() string {
				return "snmpcollector"
			},
			"AppVer": func() string {
				return "0.1.0"
			},
		}},
		// Delims sets the action delimiters to the specified strings. Defaults are ["{{", "}}"].
		Delims: macaron.Delims{"{{", "}}"},
		// Appends the given charset to the Content-Type header. Default is "UTF-8".
		Charset: "UTF-8",
		// Outputs human readable JSON. Default is false.
		IndentJSON: true,
		// Outputs human readable XML. Default is false.
		IndentXML: true,
		// Prefixes the JSON output with the given bytes. Default is no prefix.
		// PrefixJSON: []byte("macaron"),
		// Prefixes the XML output with the given bytes. Default is no prefix.
		// PrefixXML: []byte("macaron"),
		// Allows changing of output to XHTML instead of HTML. Default is "text/html".
		HTMLContentType: "text/html",
	}))
	m.Use(cache.Cacher(cache.Options{
		// Name of adapter. Default is "memory".
		Adapter: "memory",
		// Adapter configuration, it's corresponding to adapter.
		AdapterConfig: "",
		// GC interval time in seconds. Default is 60.
		Interval: 60,
		// Configuration section name. Default is "cache".
		Section: "cache",
	}))

	// setup handlers
	//	m.Get("/", myHandler)
	//	m.Get("/welcome", myOtherHandler)
	//	m.Get("/query", myQueryStringHandler) // /query?name=Some+Name
	//	m.Get("/json", myJsonHandler)
	//	m.Post("/contact/submit", binding.Bind(ContactForm{}), mySubmitHandler)
	//	m.Get("/session", mySessionHandler)
	m.Post("/session/create", bind(UserLogin{}), myLoginHandler)

	//m.Get("/set/cookie/:value", mySetCookieHandler)
	//m.Get("/get/cookie", myGetCookieHandler)
	//m.Get("/database", myDatabaseHandler)
	//m.Get("/snpmdevices/list", myDatabaseListHandler)
	//m.Get("/cache/write/:key/:value", myCacheWriteHandler)
	//m.Get("/cache/read/:key", myCacheReadHandler)

	// Data sources
	m.Get("/runtimeinfo", GetRuntimeInfo)

	m.Group("/snmpdevice", func() {
		m.Get("/", GetSNMPDevices)
		m.Post("/", bind(SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", bind(SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", DeleteSNMPDevice)
		m.Get("/:id", GetSNMPDeviceByID)
	})

	m.Group("/metric", func() {
		m.Get("/", GetMetrics)
		m.Post("/", bind(SnmpMetricCfg{}), AddMetric)
		m.Put("/:id", bind(SnmpMetricCfg{}), UpdateMetric)
		m.Delete("/:id", DeleteMetric)
		m.Get("/:id", GetMetricByID)
	})

	m.Group("/measurement", func() {
		m.Get("/", GetMeas)
		m.Post("/", bind(InfluxMeasurementCfg{}), AddMeas)
		m.Put("/:id", bind(InfluxMeasurementCfg{}), UpdateMeas)
		m.Delete("/:id", DeleteMeas)
		m.Get("/:id", GetMeasByID)
	})

	m.Group("/measgroups", func() {
		m.Get("/", GetMeasGroup)
		m.Post("/", bind(MGroupsCfg{}), AddMeasGroup)
		m.Put("/:id", bind(MGroupsCfg{}), UpdateMeasGroup)
		m.Delete("/:id", DeleteMeasGroup)
		m.Get("/:id", GetMeasGroupByID)
	})

	m.Group("/measfilters", func() {
		m.Get("/", GetMeasFilter)
		m.Post("/", bind(MeasFilterCfg{}), AddMeasFilter)
		m.Put("/:id", bind(MeasFilterCfg{}), UpdateMeasFilter)
		m.Delete("/:id", DeleteMeasFilter)
		m.Get("/:id", GetMeasFilterByID)
	})

	m.Group("/influxservers", func() {
		m.Get("/", GetInfluxServer)
		m.Post("/", bind(InfluxCfg{}), AddInfluxServer)
		m.Put("/:id", bind(InfluxCfg{}), UpdateInfluxServer)
		m.Delete("/:id", DeleteInfluxServer)
		m.Get("/:id", GetInfluxServerByID)
	})

	log.Printf("Server is running on localhost:%d...", port)
	httpServer := fmt.Sprintf("0.0.0.0:%d", port)
	log.Println(http.ListenAndServe(httpServer, m))
}

/****************/
/*Runtime Info
/****************/

func GetRuntimeInfo(ctx *macaron.Context) {
	log.Debugf("Got device runtime info s %+v", &devices)
	ctx.JSON(200, &devices)
}

/****************/
/*SNMP DEVICES
/****************/

// GetSNMPDevices Return snmpdevice list to frontend
func GetSNMPDevices(ctx *macaron.Context) {

	devcfgarray := make([]SnmpDeviceCfg, 0)
	err := cfg.Database.x.Find(&devcfgarray)
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Devices :%+s", err)
		return
	}
	ctx.JSON(200, &devcfgarray)
	log.Debugf("Getting DEVICEs %+v", &devcfgarray)

}

// AddSNMPDevice Insert new snmpdevice to de internal BBDD --pending--
func AddSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
	log.Printf("ADDING DEVICE %+v", dev)
	cfg.SnmpDevice[dev.ID] = &dev
	affected, err := cfg.Database.x.Insert(dev)
	if err != nil {
		log.Warningf("Error on insert for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.SnmpDevice[dev.ID])
	}

}

// UpdateSNMPDevice --pending--
func UpdateSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
	id := ctx.Params(":id")

	log.Debugf("CFG SNMPDEVICE [%s]: %+v", id, cfg.SnmpDevice)
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.x.Where("id='" + id + "'").Update(dev)
	if err != nil {
		log.Warningf("Error on update for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.SnmpDevice[dev.ID])
	}
}

//DeleteSNMPDevice --pending--
func DeleteSNMPDevice(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.x.Where("id='" + id + "'").Delete(&SnmpDeviceCfg{})
	if err != nil {
		log.Warningf("Error on delete1 for device %s  , affected : %+v , error: %s", id, affected, err)
	} else {
		ctx.JSON(200, "deleted")
	}

}

//GetSNMPDeviceByID --pending--
func GetSNMPDeviceByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("ID: %+v", cfg.SnmpDevice[id])
	ctx.JSON(200, cfg.SnmpDevice[id])
}

/****************/
/*SNMP METRICS
/****************/

// GetMetrics Return metrics list to frontend
func GetMetrics(ctx *macaron.Context) {
	metcfgarray := make([]SnmpMetricCfg, 0)
	err := cfg.Database.x.Find(&metcfgarray)
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Metrics :%+s", err)
		return
	}
	ctx.JSON(200, &metcfgarray)
	log.Debugf("Getting Metric Array %+v", &metcfgarray)
}

// AddMetric Insert new metric to de internal BBDD --pending--
func AddMetric(ctx *macaron.Context, met SnmpMetricCfg) {
	log.Printf("ADDING METRIC %+v", met)
	affected, err := cfg.Database.x.Insert(met)
	if err != nil {
		log.Warningf("Error on insert for metric %s  , affected : %+v , error: %s", met.ID, affected, err)
	} else {
		ctx.JSON(200, &met)
	}

}

// UpdateMetric --pending--
func UpdateMetric(ctx *macaron.Context, met SnmpMetricCfg) {
	id := ctx.Params(":id")
	log.Debugf("CFG METRIC [%s]: %+v", id, cfg.Metrics)
	log.Debugf("Tying to update: %+v", met)
	affected, err := cfg.Database.x.Where("id='" + id + "'").Update(met)
	if err != nil {
		log.Warningf("Error on update for metric %s  , affected : %+v , error: %s", met.ID, affected, err)
	} else {
		ctx.JSON(200, &met)
	}
}

//DeleteMetric --pending--
func DeleteMetric(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	//TODO: these two deletes should be done on the same commit (xorm session?)
	//Deleting from SQL DDBB - snmp_metric_cfg
	affected, err := cfg.Database.x.Where("id='" + id + "'").Delete(&SnmpMetricCfg{})
	if err != nil {
		log.Warningf("Error on delete for metric %s  , affected : %+v , error: %s", id, affected, err)
	} else {
		ctx.JSON(200, "deleted")
	}

	//Relationals: Measurement|Metric -> measurement_field_cfg
	// If it is deleted, delete all relations with the ID
	// Creating the interface:
	affected, err = cfg.Database.x.Where("id_metric_cfg='" + id + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		log.Warningf("Error on delete for metric %s  , affected : %+v , error: %s", id, affected, err)
	}

}

//GetMetricByID --pending--
func GetMetricByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Printf("ID: %+v", cfg.Metrics[id])
	ctx.JSON(200, cfg.Metrics[id])
}

/****************/
/*INFLUX MEASUREMENTS
/****************/

// GetMeas Return measurements list to frontend
func GetMeas(ctx *macaron.Context) {
	ctx.JSON(200, &cfg.Measurements)
	log.Printf("Getting Measurements %+v", &cfg.Measurements)
}

// AddMeas Insert new measurement to de internal BBDD --pending--
func AddMeas(ctx *macaron.Context, meas InfluxMeasurementCfg) {
	log.Printf("ADDING Measurement %+v", meas)
	//Actualizando la configuracion de measurements
	cfg.Measurements[meas.ID] = &meas

	//Actualizando la relacional MeasurementFieldCfg????
	//TODO: review this what this loop do
	tempCfg := &MeasurementFieldCfg{}
	for _, Field := range meas.Fields {
		tempCfg.IDMeasurementCfg = (meas.ID)
		tempCfg.IDMetricCfg = (Field)
	}
	affected, err := cfg.Database.x.Insert(tempCfg)
	if err != nil {
		log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", meas.ID, affected, err)
	}

	log.Printf("TESTING: +%v", tempCfg)

	affected, err = cfg.Database.x.Insert(meas)
	if err != nil {
		log.Warningf("Error on insert for measurement %s  , affected : %+v , error: %s", meas.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.Measurements[meas.ID])
	}

}

// UpdateMeas --pending--
func UpdateMeas(ctx *macaron.Context, meas InfluxMeasurementCfg) {
	id := ctx.Params(":id")

	log.Debugf("CFG MEASUREMENTS [%s]: %+v", id, cfg.Measurements)
	log.Debugf("Tying to update: %+v", meas)
	affected, err := cfg.Database.x.Where("id='" + id + "'").Update(meas)
	if err != nil {
		log.Warningf("Error on update for measurement %s  , affected : %+v , error: %s", meas.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.Measurements[meas.ID])
	}
	tempCfg := &MeasurementFieldCfg{}
	for _, Field := range meas.Fields {
		tempCfg.IDMeasurementCfg = (meas.ID)
		tempCfg.IDMetricCfg = (Field)
	}
	//Update in case that is has changed using the OLD ID
	affected, err = cfg.Database.x.Where("id_measurement_cfg='" + id + "' AND id_metric_cfg='" + cfg.Measurements[id].Fields[0] + "'").Update(tempCfg)
	if err != nil {
		log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", meas.ID, affected, err)
	}

	delete(cfg.Measurements, id)
	cfg.Measurements[meas.ID] = &meas

}

//DeleteMeas --pending--
func DeleteMeas(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	//Delete from SQL:
	affected, err := cfg.Database.x.Where("id='" + id + "'").Delete(&InfluxMeasurementCfg{})
	if err != nil {
		log.Warningf("Error on delete for measurement %s  , affected : %+v , error: %s", id, affected, err)
	} else {
		ctx.JSON(200, "deleted")
	}
	affected, err = cfg.Database.x.Where("id_measurement_cfg='" + cfg.Measurements[id].ID + "' AND id_metric_cfg='" + cfg.Measurements[id].Fields[0] + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		log.Warningf("Error on delete for measurement %s  , affected : %+v , error: %s", id, affected, err)
	}
	//Delete from config map
	delete(cfg.Measurements, id)
}

//GetMeasByID --pending--
func GetMeasByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Printf("ID: %+v", cfg.Measurements[id])
	ctx.JSON(200, cfg.Measurements[id])
}

/****************/
/*MEASUREMENT GROUPS
/****************/

// GetMeasGroup Return measurements groups list to frontend
func GetMeasGroup(ctx *macaron.Context) {
	ctx.JSON(200, &cfg.GetGroups)
	log.Printf("Getting Measurement Groups %+v", &cfg.GetGroups)
}

// AddMeasGroup Insert new measurement groups to de internal BBDD --pending--
func AddMeasGroup(ctx *macaron.Context, mgroup MGroupsCfg) {
	log.Printf("ADDING Measurement %+v", mgroup)
	//Actualizando la configuracion de measurements groups
	cfg.GetGroups[mgroup.ID] = &mgroup

	//Updating DDBB - m_group_cfg - {measgroup ID}
	affected, err := cfg.Database.x.Insert(mgroup)
	if err != nil {
		log.Warningf("Error on insert for measurement groups %s  , affected : %+v , error: %s", mgroup.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.GetGroups[mgroup.ID])
	}

	//Updating bound DDBB m_groups_measurement
	//Creating new interface
	tempCfg := &MGroupsMeasurements{}
	for _, Measurement := range mgroup.Measurements {
		tempCfg.IDMGroupCfg = (mgroup.ID)
		tempCfg.IDMeasurementCfg = (Measurement)
	}
	affected, err = cfg.Database.x.Insert(tempCfg)
	if err != nil {
		log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", mgroup.ID, affected, err)
	}
}

// UpdateMeasGroup --pending--
func UpdateMeasGroup(ctx *macaron.Context, mgroup MGroupsCfg) {
	id := ctx.Params(":id")

	log.Debugf("CFG MEASUREMENTS [%s]: %+v", id, cfg.GetGroups)
	log.Debugf("Tying to update: %+v", mgroup)
	affected, err := cfg.Database.x.Where("id='" + id + "'").Update(mgroup)
	if err != nil {
		log.Warningf("Error on update for measurement groups %s  , affected : %+v , error: %s", mgroup.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.GetGroups[mgroup.ID])
	}
	tempCfg := &MGroupsMeasurements{}
	for _, Measurement := range mgroup.Measurements {
		log.Debugf("MEASUREMENT: %+v", Measurement)
		tempCfg.IDMGroupCfg = (mgroup.ID)
		tempCfg.IDMeasurementCfg = (Measurement)
	}
	//Update in case that is has changed using the OLD ID
	affected, err = cfg.Database.x.Where("id_mgroup_cfg='" + id + "' AND id_measurement_cfg='" + cfg.GetGroups[id].Measurements[0] + "'").Update(tempCfg)
	if err != nil {
		log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", mgroup.ID, affected, err)
	}

	delete(cfg.GetGroups, id)
	cfg.GetGroups[mgroup.ID] = &mgroup

}

//DeleteMeasGroup --pending--
func DeleteMeasGroup(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	//Delete from SQL:
	affected, err := cfg.Database.x.Where("id='" + id + "'").Delete(&MGroupsCfg{})
	if err != nil {
		log.Warningf("Error on delete for measurement group %s  , affected : %+v , error: %s", id, affected, err)
	} else {
		ctx.JSON(200, "deleted")
	}
	affected, err = cfg.Database.x.Where("id_mgroup_cfg='" + cfg.GetGroups[id].ID + "' AND id_measurement_cfg='" + cfg.GetGroups[id].Measurements[0] + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		log.Warningf("Error on delete for measurement %s  , affected : %+v , error: %s", id, affected, err)
	}

	//Delete from config map
	delete(cfg.GetGroups, id)
}

//GetMeasGroupByID --pending--
func GetMeasGroupByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Printf("ID: %+v", cfg.GetGroups[id])
	ctx.JSON(200, cfg.GetGroups[id])
}

/****************/
/*MEASUREMENT FILTERS
/****************/

// GetMeasFilter Return measurements groups list to frontend
func GetMeasFilter(ctx *macaron.Context) {
	ctx.JSON(200, &cfg.MFilters)
	log.Printf("Getting Measurement Filters %+v", &cfg.MFilters)
}

// AddMeasFilter Insert new measurement groups to de internal BBDD --pending--
func AddMeasFilter(ctx *macaron.Context, mfilter MeasFilterCfg) {
	log.Printf("ADDING Measurement Filter %+v", mfilter)
	//Actualizando la configuracion de measurements groups
	cfg.MFilters[mfilter.ID] = &mfilter

	//Updating DDBB - m_group_cfg - {measgroup ID}
	affected, err := cfg.Database.x.Insert(mfilter)
	if err != nil {
		log.Warningf("Error on insert for measurement filters %s  , affected : %+v , error: %s", mfilter.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.MFilters[mfilter.ID])
	}

	//Updating bound DDBB m_groups_measurement
	//Creating new interface
	/*	tempCfg := &MGroupsMeasurements{}
		for _, Measurement := range mfilter.Measurements {
			tempCfg.IDMGroupCfg = (mfilter.ID)
			tempCfg.IDMeasurementCfg = (Measurement)
		}
		affected, err = cfg.Database.x.Insert(tempCfg)
		if err != nil {
			log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", mfilter.ID, affected, err)
		}
	*/
}

// UpdateMeasFilter --pending--
func UpdateMeasFilter(ctx *macaron.Context, mfilter MeasFilterCfg) {
	id := ctx.Params(":id")

	log.Debugf("CFG MEASUREMENTS [%s]: %+v", id, cfg.MFilters)
	log.Debugf("Tying to update: %+v", mfilter)
	affected, err := cfg.Database.x.Where("id='" + id + "'").Update(mfilter)
	if err != nil {
		log.Warningf("Error on update for measurement filters %s  , affected : %+v , error: %s", mfilter.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.MFilters[mfilter.ID])
	}
	/*tempCfg := &MGroupsMeasurements{}
	for _, Measurement := range mfilter.Measurements {
		log.Debugf("MEASUREMENT: %+v", Measurement)
		tempCfg.IDMGroupCfg = (mgroup.ID)
		tempCfg.IDMeasurementCfg = (Measurement)
	}
	//Update in case that is has changed using the OLD ID
		affected, err = cfg.Database.x.Where("id_mgroup_cfg='" + id + "' AND id_measurement_cfg='" + cfg.MFilters[id].Measurements[0] + "'").Update(tempCfg)
		if err != nil {
			log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", mgroup.ID, affected, err)
		}
	*/
	delete(cfg.MFilters, id)
	cfg.MFilters[mfilter.ID] = &mfilter

}

//DeleteMeasFilter --pending--
func DeleteMeasFilter(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	//Delete from SQL:
	affected, err := cfg.Database.x.Where("id='" + id + "'").Delete(&MeasFilterCfg{})
	if err != nil {
		log.Warningf("Error on delete for measurement group %s  , affected : %+v , error: %s", id, affected, err)
	} else {
		ctx.JSON(200, "deleted")
	}
	/*affected, err = cfg.Database.x.Where("id_mgroup_cfg='" + cfg.GetGroups[id].ID + "' AND id_measurement_cfg='" + cfg.GetGroups[id].Measurements[0] + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		log.Warningf("Error on delete for measurement %s  , affected : %+v , error: %s", id, affected, err)
	}
	*/
	//Delete from config map
	delete(cfg.MFilters, id)
}

//GetMeasFilterByID --pending--
func GetMeasFilterByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Printf("ID: %+v", cfg.MFilters[id])
	ctx.JSON(200, cfg.MFilters[id])
}

/****************/
/* INFLUX SERVERS
/****************/

// GetInfluxServer Return measurements groups list to frontend
func GetInfluxServer(ctx *macaron.Context) {
	ctx.JSON(200, &cfg.Influxdb)
	log.Printf("Getting Measurement Filters %+v", &cfg.Influxdb)
}

// AddInfluxServer Insert new measurement groups to de internal BBDD --pending--
func AddInfluxServer(ctx *macaron.Context, iserver InfluxCfg) {
	log.Printf("ADDING Measurement Filter %+v", iserver)
	//Actualizando la configuracion de measurements groups
	cfg.Influxdb[iserver.ID] = &iserver

	//Updating DDBB - m_group_cfg - {measgroup ID}
	affected, err := cfg.Database.x.Insert(iserver)
	if err != nil {
		log.Warningf("Error on insert for influx server %s  , affected : %+v , error: %s", iserver.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.Influxdb[iserver.ID])
	}

	//Updating bound DDBB m_groups_measurement
	//Creating new interface
	/*	tempCfg := &MGroupsMeasurements{}
		for _, Measurement := range mfilter.Measurements {
			tempCfg.IDMGroupCfg = (mfilter.ID)
			tempCfg.IDMeasurementCfg = (Measurement)
		}
		affected, err = cfg.Database.x.Insert(tempCfg)
		if err != nil {
			log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", mfilter.ID, affected, err)
		}
	*/
}

// UpdateInfluxServer --pending--
func UpdateInfluxServer(ctx *macaron.Context, iserver InfluxCfg) {
	id := ctx.Params(":id")

	log.Debugf("CFG InfluxServer [%s]: %+v", id, cfg.MFilters)
	log.Debugf("Tying to update: %+v", iserver)
	affected, err := cfg.Database.x.Where("id='" + id + "'").Update(iserver)
	if err != nil {
		log.Warningf("Error on update for influx server %s  , affected : %+v , error: %s", iserver.ID, affected, err)
	} else {
		ctx.JSON(200, cfg.MFilters[iserver.ID])
	}
	/*tempCfg := &MGroupsMeasurements{}
	for _, Measurement := range mfilter.Measurements {
		log.Debugf("MEASUREMENT: %+v", Measurement)
		tempCfg.IDMGroupCfg = (mgroup.ID)
		tempCfg.IDMeasurementCfg = (Measurement)
	}
	//Update in case that is has changed using the OLD ID
		affected, err = cfg.Database.x.Where("id_mgroup_cfg='" + id + "' AND id_measurement_cfg='" + cfg.MFilters[id].Measurements[0] + "'").Update(tempCfg)
		if err != nil {
			log.Warningf("Error on insert on relational table for measurement %s  , affected : %+v , error: %s", mgroup.ID, affected, err)
		}
	*/
	delete(cfg.Influxdb, id)
	cfg.Influxdb[iserver.ID] = &iserver

}

//DeleteInfluxServer --pending--
func DeleteInfluxServer(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	//Delete from SQL:
	affected, err := cfg.Database.x.Where("id='" + id + "'").Delete(&InfluxCfg{})
	if err != nil {
		log.Warningf("Error on delete for influx server %s  , affected : %+v , error: %s", id, affected, err)
	} else {
		ctx.JSON(200, "deleted")
	}
	/*affected, err = cfg.Database.x.Where("id_mgroup_cfg='" + cfg.GetGroups[id].ID + "' AND id_measurement_cfg='" + cfg.GetGroups[id].Measurements[0] + "'").Delete(&MeasurementFieldCfg{})
	if err != nil {
		log.Warningf("Error on delete for measurement %s  , affected : %+v , error: %s", id, affected, err)
	}
	*/
	//Delete from config map
	delete(cfg.Influxdb, id)
}

//GetInfluxServerByID --pending--
func GetInfluxServerByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Printf("ID: %+v", cfg.Influxdb[id])
	ctx.JSON(200, cfg.Influxdb[id])
}

/****************/
/*TEST MACARON
/****************/

func myHandler(ctx *macaron.Context) {
	ctx.Data["Name"] = "Person"
	ctx.HTML(200, "hello") // 200 is the response code.
}

func myOtherHandler(ctx *macaron.Context) {
	ctx.Data["Message"] = "the request path is: " + ctx.Req.RequestURI
	ctx.HTML(200, "welcome")
}

func myLoginHandler(ctx *macaron.Context, user UserLogin) {
	fmt.Printf("USER LOGIN: USER: +%#v (Config: %#v)", user, cfg.HTTP)
	if user.UserName == cfg.HTTP.AdminUser && user.Password == cfg.HTTP.AdminPassword {
		fmt.Println("OK")
		ctx.JSON(200, "OK")
	} else {
		fmt.Println("ERROR")
		ctx.JSON(404, "ERROR")
	}
}
