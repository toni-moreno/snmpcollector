package main

import (
	"fmt"
	"github.com/go-macaron/binding"
	//"github.com/go-macaron/cache"
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
	//	"html/template"
	"crypto/md5"
	"net/http"
)

//HTTPConfig has webserver config options
type HTTPConfig struct {
	Port          int    `toml:"port"`
	AdminUser     string `toml:"adminuser"`
	AdminPassword string `toml:"adminpassword"`
	CookieID      string `toml:"cookieid"`
}

//UserLogin for login purposes
type UserLogin struct {
	UserName string `form:"username" binding:"Required"`
	Password string `form:"password" binding:"Required"`
}

var cookie string

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
	m.Use(GetContextHandler())
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

	//Cookie should be unique for each snmpcollector instance ,
	//if cockie_id is not set it takes the instanceID value to generate a unique array with as a md5sum

	cookie = cfg.HTTP.CookieID

	if len(cfg.HTTP.CookieID) == 0 {
		currentsum := md5.Sum([]byte(cfg.General.InstanceID))
		cookie = fmt.Sprintf("%x", currentsum)
	}

	m.Use(Sessioner(session.Options{
		// Name of provider. Default is "memory".
		Provider: "memory",
		// Provider configuration, it's corresponding to provider.
		ProviderConfig: "",
		// Cookie name to save session ID. Default is "MacaronSession".
		CookieName: "snmpcollector-sess-" + cookie,
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
		/*Funcs: []template.FuncMap{map[string]interface{}{
			"AppName": func() string {
				return "snmpcollector"
			},
			"AppVer": func() string {
				return "0.5.1"
			},
		}},*/
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
	/*
		m.Use(cache.Cacher(cache.Options{
			// Name of adapter. Default is "memory".
			Adapter: "memory",
			// Adapter configuration, it's corresponding to adapter.
			AdapterConfig: "",
			// GC interval time in seconds. Default is 60.
			Interval: 60,
			// Configuration section name. Default is "cache".
			Section: "cache",
		}))*/

	m.Post("/login", bind(UserLogin{}), myLoginHandler)
	m.Post("/logout", myLogoutHandler)

	m.Group("/metric", func() {
		m.Get("/", reqSignedIn, GetMetrics)
		m.Post("/", reqSignedIn, bind(SnmpMetricCfg{}), AddMetric)
		m.Put("/:id", reqSignedIn, bind(SnmpMetricCfg{}), UpdateMetric)
		m.Delete("/:id", reqSignedIn, DeleteMetric)
		m.Get("/:id", reqSignedIn, GetMetricByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMetricsAffectOnDel)
	})

	m.Group("/measurement", func() {
		m.Get("/", reqSignedIn, GetMeas)
		m.Post("/", reqSignedIn, bind(InfluxMeasurementCfg{}), AddMeas)
		m.Put("/:id", reqSignedIn, bind(InfluxMeasurementCfg{}), UpdateMeas)
		m.Delete("/:id", reqSignedIn, DeleteMeas)
		m.Get("/:id", reqSignedIn, GetMeasByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasAffectOnDel)
	})

	m.Group("/measgroups", func() {
		m.Get("/", reqSignedIn, GetMeasGroup)
		m.Post("/", reqSignedIn, bind(MGroupsCfg{}), AddMeasGroup)
		m.Put("/:id", reqSignedIn, bind(MGroupsCfg{}), UpdateMeasGroup)
		m.Delete("/:id", reqSignedIn, DeleteMeasGroup)
		m.Get("/:id", reqSignedIn, GetMeasGroupByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasGroupsAffectOnDel)
	})

	m.Group("/measfilters", func() {
		m.Get("/", reqSignedIn, GetMeasFilter)
		m.Post("/", reqSignedIn, bind(MeasFilterCfg{}), AddMeasFilter)
		m.Put("/:id", reqSignedIn, bind(MeasFilterCfg{}), UpdateMeasFilter)
		m.Delete("/:id", reqSignedIn, DeleteMeasFilter)
		m.Get("/:id", reqSignedIn, GetMeasFilterByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasFiltersAffectOnDel)
	})

	m.Group("/influxservers", func() {
		m.Get("/", reqSignedIn, GetInfluxServer)
		m.Post("/", reqSignedIn, bind(InfluxCfg{}), AddInfluxServer)
		m.Put("/:id", reqSignedIn, bind(InfluxCfg{}), UpdateInfluxServer)
		m.Delete("/:id", reqSignedIn, DeleteInfluxServer)
		m.Get("/:id", reqSignedIn, GetInfluxServerByID)
		m.Get("/checkondel/:id", reqSignedIn, GetInfluxAffectOnDel)
	})

	// Data sources
	m.Group("/snmpdevice", func() {
		m.Get("/", reqSignedIn, GetSNMPDevices)
		m.Post("/", reqSignedIn, bind(SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", reqSignedIn, bind(SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", reqSignedIn, DeleteSNMPDevice)
		m.Get("/:id", reqSignedIn, GetSNMPDeviceByID)
		m.Get("/checkondel/:id", reqSignedIn, GetSNMPDevicesAffectOnDel)
	})

	m.Group("/runtime", func() {
		m.Post("/snmpping/", reqSignedIn, bind(SnmpDeviceCfg{}), PingSNMPDevice)
		m.Get("/version/", reqSignedIn, RTGetVersion)
		m.Get("/info/", reqSignedIn, RTGetInfo)
		m.Get("/info/:id", reqSignedIn, RTGetInfo)
		m.Put("/activatedev/:id", reqSignedIn, RTActivateDev)
		m.Put("/deactivatedev/:id", reqSignedIn, RTDeactivateDev)
		m.Put("/actsnmpdbg/:id", reqSignedIn, RTActSnmpDebugDev)
		m.Put("/deactsnmpdbg/:id", reqSignedIn, RTDeactSnmpDebugDev)
		m.Put("/setloglevel/:id/:level", reqSignedIn, RTSetLogLevelDev)
	})

	log.Printf("Server is running on localhost:%d...", port)
	httpServer := fmt.Sprintf("0.0.0.0:%d", port)
	log.Println(http.ListenAndServe(httpServer, m))
}

/****************/
/*Runtime Info
/****************/

//PingSNMPDevice xx
func PingSNMPDevice(ctx *macaron.Context, cfg SnmpDeviceCfg) {
	log.Infof("trying to ping device %s : %+v", cfg.ID, cfg)

	_, sysinfo, err := SnmpClient(&cfg, log)
	if err != nil {
		log.Debugf("ERROR  on query device : %s", err)
		ctx.JSON(400, err.Error())
	} else {
		log.Debugf("OK on query device ")
		ctx.JSON(200, sysinfo)
	}
}

//RTSetLogLevelDev xx
func RTSetLogLevelDev(ctx *Context) {
	id := ctx.Params(":id")
	level := ctx.Params(":level")
	if dev, ok := devices[id]; !ok {
		ctx.JSON(404, fmt.Errorf("there is not any device with id %s running", id))
		return
	} else {
		log.Infof("set runtime log level from device id %s : %s", id, level)
		dev.RTSetLogLevel(level)
		ctx.JSON(200, dev)
	}
}

//RTActivateDev xx
func RTActivateDev(ctx *Context) {
	id := ctx.Params(":id")
	if dev, ok := devices[id]; !ok {
		ctx.JSON(404, fmt.Errorf("there is not any device with id %s running", id))
		return
	} else {
		log.Infof("activating runtime on device %s", id)
		dev.RTActivate(true)
		ctx.JSON(200, dev)
	}
}

//RTDeactivateDev xx
func RTDeactivateDev(ctx *Context) {
	id := ctx.Params(":id")
	if dev, ok := devices[id]; !ok {
		ctx.JSON(404, fmt.Errorf("there is not any device with id %s running", id))
		return
	} else {
		log.Infof("deactivating runtime on device  %s", id)
		dev.RTActivate(false)
		ctx.JSON(200, dev)
	}
}

//RTActSnmpDebugDev xx
func RTActSnmpDebugDev(ctx *Context) {
	id := ctx.Params(":id")
	if dev, ok := devices[id]; !ok {
		ctx.JSON(404, fmt.Errorf("there is not any device with id %s running", id))
		return
	} else {
		log.Infof("activating snmpdebug  %s", id)
		dev.RTActSnmpDebug(true)
		ctx.JSON(200, dev)
	}
}

//RTDeactSnmpDebugDev xx
func RTDeactSnmpDebugDev(ctx *Context) {
	id := ctx.Params(":id")
	if dev, ok := devices[id]; !ok {
		ctx.JSON(404, fmt.Errorf("there is not any device with id %s running", id))
		return
	} else {
		log.Infof("deactivating snmpdebug  %s", id)
		dev.RTActSnmpDebug(false)
		ctx.JSON(200, dev)
	}
}

type devStat struct {
	Requests           int64
	Gets               int64
	Errors             int64
	ReloadLoopsPending int
	DeviceActive       bool
	DeviceConnected    bool
	NumMeasurements    int
	NumMetrics         int
}

//RTGetInfo xx
func RTGetInfo(ctx *Context) {
	id := ctx.Params(":id")
	if len(id) > 0 {
		if dev, ok := devices[id]; !ok {
			ctx.JSON(404, fmt.Errorf("there is not any device with id %s running", id))
			return
		} else {
			log.Infof("get runtime data from id %s", id)
			ctx.JSON(200, dev)
		}
		//get only one device info
	} else {
		devstats := make(map[string]*devStat)
		for k, v := range devices {
			sum := 0
			for _, m := range v.Measurements {
				sum += len(m.OidSnmpMap)
			}
			devstats[k] = &devStat{
				Requests:           v.Requests,
				Gets:               v.Gets,
				Errors:             v.Errors,
				ReloadLoopsPending: v.ReloadLoopsPending,
				DeviceActive:       v.DeviceActive,
				DeviceConnected:    v.DeviceConnected,
				NumMeasurements:    len(v.Measurements),
				NumMetrics:         sum,
			}
		}
		ctx.JSON(200, &devstats)
	}
	return
}

type RInfo struct {
	InstanceID string
	Version    string
	Commit     string
	Branch     string
	BuildStamp string
}

//RTGetVersion xx
func RTGetVersion(ctx *Context) {
	info := &RInfo{
		InstanceID: cfg.General.InstanceID,
		Version:    version,
		Commit:     commit,
		Branch:     branch,
		BuildStamp: buildstamp,
	}
	ctx.JSON(200, &info)
}

/****************/
/*SNMP DEVICES
/****************/

// GetSNMPDevices Return snmpdevice list to frontend
func GetSNMPDevices(ctx *Context) {
	devcfgarray, err := cfg.Database.GetSnmpDeviceCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Devices :%+s", err)
		return
	}
	ctx.JSON(200, &devcfgarray)
	log.Debugf("Getting DEVICEs %+v", &devcfgarray)
}

// AddSNMPDevice Insert new snmpdevice to de internal BBDD --pending--
func AddSNMPDevice(ctx *Context, dev SnmpDeviceCfg) {
	log.Printf("ADDING DEVICE %+v", dev)
	affected, err := cfg.Database.AddSnmpDeviceCfg(dev)
	if err != nil {
		log.Warningf("Error on insert for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateSNMPDevice --pending--
func UpdateSNMPDevice(ctx *Context, dev SnmpDeviceCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateSnmpDeviceCfg(id, dev)
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
	affected, err := cfg.Database.DelSnmpDeviceCfg(id)
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
	dev, err := cfg.Database.GetSnmpDeviceCfgByID(id)
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
	obarray, err := cfg.Database.GeSnmpDeviceCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for SNMP metrics %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/****************/
/*SNMP METRICS
/****************/

// GetMetrics Return metrics list to frontend
func GetMetrics(ctx *Context) {
	cfgarray, err := cfg.Database.GetSnmpMetricCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Metrics :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Metrics %+v", &cfgarray)
}

// AddMetric Insert new metric to de internal BBDD --pending--
func AddMetric(ctx *Context, dev SnmpMetricCfg) {
	log.Printf("ADDING Metric %+v", dev)
	affected, err := cfg.Database.AddSnmpMetricCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMetric --pending--
func UpdateMetric(ctx *Context, dev SnmpMetricCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateSnmpMetricCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMetric --pending--
func DeleteMetric(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelSnmpMetricCfg(id)
	if err != nil {
		log.Warningf("Error on delete Metric %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMetricByID --pending--
func GetMetricByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetSnmpMetricCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Metric  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetMetricsAffectOnDel --pending--
func GetMetricsAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := cfg.Database.GetSnmpMetricCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for SNMP metrics %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/****************/
/*INFLUX MEASUREMENTS
/****************/

// GetMeas Return measurements list to frontend
func GetMeas(ctx *Context) {
	cfgarray, err := cfg.Database.GetInfluxMeasurementCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx Measurements :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurements %+v", &cfgarray)
}

// AddMeas Insert new measurement to de internal BBDD --pending--
func AddMeas(ctx *Context, dev InfluxMeasurementCfg) {
	log.Printf("ADDING Measurement %+v", dev)
	affected, err := cfg.Database.AddInfluxMeasurementCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeas --pending--
func UpdateMeas(ctx *Context, dev InfluxMeasurementCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateInfluxMeasurementCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeas --pending--
func DeleteMeas(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelInfluxMeasurementCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMeasByID --pending--
func GetMeasByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetInfluxMeasurementCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetMeasAffectOnDel --pending--
func GetMeasAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := cfg.Database.GetInfluxMeasurementCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurements %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/****************/
/*MEASUREMENT GROUPS
/****************/

// GetMeasGroup Return measurements groups list to frontend
func GetMeasGroup(ctx *Context) {
	cfgarray, err := cfg.Database.GetMGroupsCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Group :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Meas Group %+v", &cfgarray)
}

// AddMeasGroup Insert new measurement groups to de internal BBDD --pending--
func AddMeasGroup(ctx *Context, dev MGroupsCfg) {
	log.Printf("ADDING Measurement Group %+v", dev)
	affected, err := cfg.Database.AddMGroupsCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement Group %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasGroup --pending--
func UpdateMeasGroup(ctx *Context, dev MGroupsCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateMGroupsCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurement Group %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeasGroup --pending--
func DeleteMeasGroup(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelMGroupsCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Group %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMeasGroupByID --pending--
func GetMeasGroupByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetMGroupsCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Group for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetMeasGroupsAffectOnDel --pending--
func GetMeasGroupsAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := cfg.Database.GetMGroupsCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement Groups %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/********************/
/*MEASUREMENT FILTERS
/********************/

// GetMeasFilter Return measurements groups list to frontend
func GetMeasFilter(ctx *Context) {
	cfgarray, err := cfg.Database.GetMeasFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

// AddMeasFilter Insert new measurement groups to de internal BBDD --pending--
func AddMeasFilter(ctx *Context, dev MeasFilterCfg) {
	log.Printf("ADDING measurement Filter %+v", dev)
	affected, err := cfg.Database.AddMeasFilterCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasFilter --pending--
func UpdateMeasFilter(ctx *Context, dev MeasFilterCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateMeasFilterCfg(id, dev)
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
	affected, err := cfg.Database.DelMeasFilterCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Filter %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMeasFilterByID --pending--
func GetMeasFilterByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetMeasFilterCfgByID(id)
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
	obarray, err := cfg.Database.GetMeasFilterCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement filters %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/****************/
/* INFLUX SERVERS
/****************/

// GetInfluxServer Return Server Array
func GetInfluxServer(ctx *Context) {
	cfgarray, err := cfg.Database.GetInfluxCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx db :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// AddInfluxServer Insert new measurement groups to de internal BBDD --pending--
func AddInfluxServer(ctx *Context, dev InfluxCfg) {
	log.Printf("ADDING Influx Backend %+v", dev)
	affected, err := cfg.Database.AddInfluxCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Backend %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateInfluxServer --pending--
func UpdateInfluxServer(ctx *Context, dev InfluxCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateInfluxCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Influx db %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteInfluxServer --pending--
func DeleteInfluxServer(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelInfluxCfg(id)
	if err != nil {
		log.Warningf("Error on delete influx db %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetInfluxServerByID --pending--
func GetInfluxServerByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetInfluxCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Influx db data for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetInfluxAffectOnDel --pending--
func GetInfluxAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := cfg.Database.GetInfluxCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/****************/
/*LOGIN
/****************/

func myLoginHandler(ctx *Context, user UserLogin) {
	fmt.Printf("USER LOGIN: USER: +%#v (Config: %#v)", user, cfg.HTTP)
	if user.UserName == cfg.HTTP.AdminUser && user.Password == cfg.HTTP.AdminPassword {
		ctx.SignedInUser = user.UserName
		ctx.IsSignedIn = true
		ctx.Session.Set(SESS_KEY_USERID, user.UserName)
		log.Println("OK")
		ctx.JSON(200, cookie)
	} else {
		log.Println("ERROR")
		ctx.JSON(404, "ERROR user or password not match")
	}
}

func myLogoutHandler(ctx *Context) {
	log.Printf("USER LOGOUT: USER: +%#v ", ctx.SignedInUser)
	ctx.Session.Destory(ctx)
	//ctx.Redirect("/login")
}
