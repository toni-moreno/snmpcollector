package webui

import (
	"fmt"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/session"
	"github.com/go-macaron/toolbox"

	"github.com/Sirupsen/logrus"
	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"github.com/toni-moreno/snmpcollector/pkg/config"
	"github.com/toni-moreno/snmpcollector/pkg/data/snmp"
	"os"

	"gopkg.in/macaron.v1"
	//	"html/template"
	"crypto/md5"
	"net/http"
	"time"
)

var (
	logDir     string
	log        *logrus.Logger
	confHTTP   *config.HTTPConfig
	instanceID string
)

// SetLogDir et dir for logs
func SetLogDir(dir string) {
	logDir = dir
}

// SetLogger set output log
func SetLogger(l *logrus.Logger) {
	log = l
}

//UserLogin for login purposes
type UserLogin struct {
	UserName string `form:"username" binding:"Required"`
	Password string `form:"password" binding:"Required"`
}

var cookie string

// WebServer the main process
func WebServer(publicPath string, httpPort int, cfg *config.HTTPConfig, id string) {
	confHTTP = cfg
	instanceID = id
	var port int
	if cfg.Port > 0 {
		port = cfg.Port
	} else {
		port = httpPort
	}

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

	f, _ := os.OpenFile(logDir+"/http_access.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	m := macaron.NewWithLogger(f)
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(toolbox.Toolboxer(m))
	// register middleware
	m.Use(GetContextHandler())
	//	m.Use(gzip.Gziper())
	log.Infof("setting HTML Static Path to %s", publicPath)
	m.Use(macaron.Static(publicPath,
		macaron.StaticOptions{
			// Prefix is the optional prefix used to serve the static directory content. Default is empty string.
			Prefix: "",
			// SkipLogging will disable [Static] log messages when a static file is served. Default is false.
			SkipLogging: false,
			// IndexFile defines which file to serve as index if it exists. Default is "index.html".
			IndexFile: "index.html",
			// Expires defines which user-defined function to use for producing a HTTP Expires Header. Default is nil.
			// https://developers.google.com/speed/docs/insights/LeverageBrowserCaching
			Expires: func() string { return "max-age=0" },
		}))

	//Cookie should be unique for each snmpcollector instance ,
	//if cockie_id is not set it takes the instanceID value to generate a unique array with as a md5sum

	cookie = confHTTP.CookieID

	if len(confHTTP.CookieID) == 0 {
		currentsum := md5.Sum([]byte(instanceID))
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

	m.Group("/api/cfg/oidcondition", func() {
		m.Get("/", reqSignedIn, GetOidConditions)
		m.Post("/", reqSignedIn, bind(config.OidConditionCfg{}), AddOidCondition)
		m.Put("/:id", reqSignedIn, bind(config.OidConditionCfg{}), UpdateOidCondition)
		m.Delete("/:id", reqSignedIn, DeleteOidCondition)
		m.Get("/:id", reqSignedIn, GetOidConditionByID)
		m.Get("/checkondel/:id", reqSignedIn, GetOidConditionAffectOnDel)
	})

	m.Group("/api/cfg/metric", func() {
		m.Get("/", reqSignedIn, GetMetrics)
		m.Post("/", reqSignedIn, bind(config.SnmpMetricCfg{}), AddMetric)
		m.Put("/:id", reqSignedIn, bind(config.SnmpMetricCfg{}), UpdateMetric)
		m.Delete("/:id", reqSignedIn, DeleteMetric)
		m.Get("/:id", reqSignedIn, GetMetricByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMetricsAffectOnDel)
	})

	m.Group("/api/cfg/measurement", func() {
		m.Get("/", reqSignedIn, GetMeas)
		m.Get("/type/:type", reqSignedIn, GetMeasByType)
		m.Post("/", reqSignedIn, bind(config.MeasurementCfg{}), AddMeas)
		m.Put("/:id", reqSignedIn, bind(config.MeasurementCfg{}), UpdateMeas)
		m.Delete("/:id", reqSignedIn, DeleteMeas)
		m.Get("/:id", reqSignedIn, GetMeasByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasAffectOnDel)
	})

	m.Group("/api/cfg/measgroup", func() {
		m.Get("/", reqSignedIn, GetMeasGroup)
		m.Post("/", reqSignedIn, bind(config.MGroupsCfg{}), AddMeasGroup)
		m.Put("/:id", reqSignedIn, bind(config.MGroupsCfg{}), UpdateMeasGroup)
		m.Delete("/:id", reqSignedIn, DeleteMeasGroup)
		m.Get("/:id", reqSignedIn, GetMeasGroupByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasGroupsAffectOnDel)
	})

	m.Group("/api/cfg/measfilters", func() {
		m.Get("/", reqSignedIn, GetMeasFilter)
		m.Post("/", reqSignedIn, bind(config.MeasFilterCfg{}), AddMeasFilter)
		m.Put("/:id", reqSignedIn, bind(config.MeasFilterCfg{}), UpdateMeasFilter)
		m.Delete("/:id", reqSignedIn, DeleteMeasFilter)
		m.Get("/:id", reqSignedIn, GetMeasFilterByID)
		m.Get("/checkondel/:id", reqSignedIn, GetMeasFiltersAffectOnDel)
	})

	m.Group("/api/cfg/influxservers", func() {
		m.Get("/", reqSignedIn, GetInfluxServer)
		m.Post("/", reqSignedIn, bind(config.InfluxCfg{}), AddInfluxServer)
		m.Put("/:id", reqSignedIn, bind(config.InfluxCfg{}), UpdateInfluxServer)
		m.Delete("/:id", reqSignedIn, DeleteInfluxServer)
		m.Get("/:id", reqSignedIn, GetInfluxServerByID)
		m.Get("/checkondel/:id", reqSignedIn, GetInfluxAffectOnDel)
	})

	// Data sources
	m.Group("/api/cfg/snmpdevice", func() {
		m.Get("/", reqSignedIn, GetSNMPDevices)
		m.Post("/", reqSignedIn, bind(config.SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", reqSignedIn, bind(config.SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", reqSignedIn, DeleteSNMPDevice)
		m.Get("/:id", reqSignedIn, GetSNMPDeviceByID)
		m.Get("/checkondel/:id", reqSignedIn, GetSNMPDevicesAffectOnDel)
	})

	m.Group("/api/cfg/customfilter", func() {
		m.Get("/", reqSignedIn, GetCustomFilter)
		m.Post("/", reqSignedIn, bind(config.CustomFilterCfg{}), AddCustomFilter)
		m.Put("/:id", reqSignedIn, bind(config.CustomFilterCfg{}), UpdateCustomFilter)
		m.Delete("/:id", reqSignedIn, DeleteCustomFilter)
		m.Get("/:id", reqSignedIn, GetCustomFilterByID)
		m.Get("/checkondel/:id", reqSignedIn, GetCustomFiltersAffectOnDel)
	})

	m.Group("/api/rt/agent", func() {
		m.Get("/reload/", reqSignedIn, AgentReloadConf)
		m.Post("/snmpconsole/ping/", reqSignedIn, bind(config.SnmpDeviceCfg{}), PingSNMPDevice)
		m.Post("/snmpconsole/query/:getmode/:obtype/:data", reqSignedIn, bind(config.SnmpDeviceCfg{}), QuerySNMPDevice)
		m.Get("/info/version/", reqSignedIn, RTGetVersion)
	})

	m.Group("/api/rt/device", func() {
		m.Get("/info/", reqSignedIn, RTGetInfo)
		m.Get("/info/:id", reqSignedIn, RTGetInfo)
		m.Put("/status/activate/:id", reqSignedIn, RTActivateDev)
		m.Put("/status/deactivate/:id", reqSignedIn, RTDeactivateDev)
		m.Put("/debug/activate/:id", reqSignedIn, RTActSnmpDebugDev)
		m.Put("/debug/deactivate/:id", reqSignedIn, RTDeactSnmpDebugDev)
		m.Put("/log/setloglevel/:id/:level", reqSignedIn, RTSetLogLevelDev)
		m.Get("/log/getdevicelog/:id", reqSignedIn, RTGetLogFileDev)
		m.Get("/filter/forcefltupdate/:id", reqSignedIn, RTForceFltUpdate)
	})

	log.Printf("Server is running on localhost:%d...", port)
	httpServer := fmt.Sprintf("0.0.0.0:%d", port)
	log.Println(http.ListenAndServe(httpServer, m))
}

/****************/
/*Runtime Info
/****************/

// AgentReloadConf xx
func AgentReloadConf(ctx *Context) {
	log.Info("trying to reload configuration for all devices")
	time := agent.ReloadConf()
	ctx.JSON(200, time)
}

// RTForceFltUpdate xx
func RTForceFltUpdate(ctx *Context) {
	id := ctx.Params(":id")
	d, err := agent.GetDevice(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	log.Info("trying to force filter for device %s", id)
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

//PingSNMPDevice xx
func PingSNMPDevice(ctx *Context, cfg config.SnmpDeviceCfg) {
	log.Infof("trying to ping device %s : %+v", cfg.ID, cfg)

	_, sysinfo, err := snmp.GetClient(&cfg, log)
	if err != nil {
		log.Debugf("ERROR  on query device : %s", err)
		ctx.JSON(400, err.Error())
	} else {
		log.Debugf("OK on query device ")
		ctx.JSON(200, sysinfo)
	}
}

// QuerySNMPDevice xx
func QuerySNMPDevice(ctx *Context, cfg config.SnmpDeviceCfg) {
	getmode := ctx.Params(":getmode")
	obtype := ctx.Params(":obtype")
	data := ctx.Params(":data")

	log.Infof("trying to query device %s : getmode: %s objectype: %s data %s", cfg.ID, getmode, obtype, data)

	if obtype != "oid" {
		log.Warnf("Object Type [%s] Not Supperted", obtype)
		ctx.JSON(400, "Object Type [ "+obtype+"] Not Supperted")
		return
	}

	snmpcli, info, err := snmp.GetClient(&cfg, log)
	if err != nil {
		log.Debugf("ERROR  on open connection with device %s : %s", cfg.ID, err)
		ctx.JSON(400, err.Error())
		return
	}
	start := time.Now()
	result, err := snmp.Query(snmpcli, getmode, data)
	elapsed := time.Since(start)
	if err != nil {
		log.Debugf("ERROR  on query device : %s", err)
		ctx.JSON(400, err.Error())
		return
	}
	log.Debugf("OK on query device ")
	snmpdata := struct {
		DeviceCfg   *config.SnmpDeviceCfg
		TimeTaken   float64
		PingInfo    *snmp.SysInfo
		QueryResult []snmp.EasyPDU
	}{
		&cfg,
		elapsed.Seconds(),
		info,
		result,
	}
	ctx.JSON(200, snmpdata)
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
	ctx.JSON(200, dev)

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
	ctx.JSON(200, dev)
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
	ctx.JSON(200, dev)

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
	ctx.JSON(200, dev)
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
	ctx.JSON(200, dev)
}

//RTGetInfo xx
func RTGetInfo(ctx *Context) {
	id := ctx.Params(":id")
	if len(id) > 0 {
		dev, err := agent.GetDevice(id)
		if err != nil {
			ctx.JSON(404, err.Error())
			return
		}

		log.Infof("get runtime data from id %s", id)
		ctx.JSON(200, dev)

		//get only one device info
	} else {
		devstats := agent.GetDevStats()
		ctx.JSON(200, &devstats)
	}
	return
}

//RTGetVersion xx
func RTGetVersion(ctx *Context) {
	info := agent.GetRInfo()
	ctx.JSON(200, &info)
}

/****************/
/*SNMP DEVICES
/****************/

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

/****************/
/*OID CONDITIONS
/****************/

// GetOidConditions Return metrics list to frontend
func GetOidConditions(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetOidConditionCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get OID contition :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting OID contitions %+v", &cfgarray)
}

// AddOidCondition Insert new condition to de internal BBDD --pending--
func AddOidCondition(ctx *Context, dev config.OidConditionCfg) {
	log.Printf("ADDING OidCondition %+v", dev)
	affected, err := agent.MainConfig.Database.AddOidConditionCfg(dev)
	if err != nil {
		log.Warningf("Error on insert OID condition %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMetric --pending--
func UpdateOidCondition(ctx *Context, dev config.OidConditionCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateOidConditionCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update OID Condition %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteOidCondition --pending--
func DeleteOidCondition(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelOidConditionCfg(id)
	if err != nil {
		log.Warningf("Error on delete OidCondition %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetOidConditionByID --pending--
func GetOidConditionByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetOidConditionCfgByID(id)
	if err != nil {
		log.Warningf("Error on get OidCondition  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetOidConditionAffectOnDel --pending--
func GetOidConditionAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetOidConditionCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for OID conditions  %s  , error: %s", id, err)
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
	cfgarray, err := agent.MainConfig.Database.GetSnmpMetricCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Metrics :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Metrics %+v", &cfgarray)
}

// AddMetric Insert new metric to de internal BBDD --pending--
func AddMetric(ctx *Context, dev config.SnmpMetricCfg) {
	log.Printf("ADDING Metric %+v", dev)
	affected, err := agent.MainConfig.Database.AddSnmpMetricCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMetric --pending--
func UpdateMetric(ctx *Context, dev config.SnmpMetricCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateSnmpMetricCfg(id, dev)
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
	affected, err := agent.MainConfig.Database.DelSnmpMetricCfg(id)
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
	dev, err := agent.MainConfig.Database.GetSnmpMetricCfgByID(id)
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
	obarray, err := agent.MainConfig.Database.GetSnmpMetricCfgAffectOnDel(id)
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
	cfgarray, err := agent.MainConfig.Database.GetMeasurementCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx Measurements :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurements %+v", &cfgarray)
}

// GetMeasByType Return measurements list to frontend
func GetMeasByType(ctx *Context) {
	t := ctx.Params(":type")
	cfgarray, err := agent.MainConfig.Database.GetMeasurementCfgArray("getmode like '%" + t + "%'")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx Measurements :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurements %+v", &cfgarray)
}

// AddMeas Insert new measurement to de internal BBDD --pending--
func AddMeas(ctx *Context, dev config.MeasurementCfg) {
	log.Printf("ADDING Measurement %+v", dev)
	affected, err := agent.MainConfig.Database.AddMeasurementCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeas --pending--
func UpdateMeas(ctx *Context, dev config.MeasurementCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMeasurementCfg(id, dev)
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
	affected, err := agent.MainConfig.Database.DelMeasurementCfg(id)
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
	dev, err := agent.MainConfig.Database.GetMeasurementCfgByID(id)
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
	obarray, err := agent.MainConfig.Database.GetMeasurementCfgAffectOnDel(id)
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
	cfgarray, err := agent.MainConfig.Database.GetMGroupsCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Group :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Meas Group %+v", &cfgarray)
}

// AddMeasGroup Insert new measurement groups to de internal BBDD --pending--
func AddMeasGroup(ctx *Context, dev config.MGroupsCfg) {
	log.Printf("ADDING Measurement Group %+v", dev)
	affected, err := agent.MainConfig.Database.AddMGroupsCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement Group %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasGroup --pending--
func UpdateMeasGroup(ctx *Context, dev config.MGroupsCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMGroupsCfg(id, dev)
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
	affected, err := agent.MainConfig.Database.DelMGroupsCfg(id)
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
	dev, err := agent.MainConfig.Database.GetMGroupsCfgByID(id)
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
	obarray, err := agent.MainConfig.Database.GetMGroupsCfgAffectOnDel(id)
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
	cfgarray, err := agent.MainConfig.Database.GetMeasFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Measurement Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

// AddMeasFilter Insert new measurement groups to de internal BBDD --pending--
func AddMeasFilter(ctx *Context, dev config.MeasFilterCfg) {
	log.Printf("ADDING measurement Filter %+v", dev)
	affected, err := agent.MainConfig.Database.AddMeasFilterCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasFilter --pending--
func UpdateMeasFilter(ctx *Context, dev config.MeasFilterCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateMeasFilterCfg(id, dev)
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
	affected, err := agent.MainConfig.Database.DelMeasFilterCfg(id)
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
	dev, err := agent.MainConfig.Database.GetMeasFilterCfgByID(id)
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
	obarray, err := agent.MainConfig.Database.GetMeasFilterCfgAffectOnDel(id)
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
	cfgarray, err := agent.MainConfig.Database.GetInfluxCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Influx db :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// AddInfluxServer Insert new measurement groups to de internal BBDD --pending--
func AddInfluxServer(ctx *Context, dev config.InfluxCfg) {
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

//GetInfluxServerByID --pending--
func GetInfluxServerByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetInfluxCfgByID(id)
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
	obarray, err := agent.MainConfig.Database.GetInfluxCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/********************/
/*CUSTOM FILTER API
/********************/

// GetCustomFilter Return measurements groups list to frontend
func GetCustomFilter(ctx *Context) {
	cfgarray, err := agent.MainConfig.Database.GetCustomFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err.Error())
		log.Errorf("Error on get Custom Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

// AddCustomFilter Insert new measurement groups to de internal BBDD --pending--
func AddCustomFilter(ctx *Context, dev config.CustomFilterCfg) {
	log.Printf("ADDING measurement Filter %+v", dev)
	affected, err := agent.MainConfig.Database.AddCustomFilterCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateCustomFilter --pending--
func UpdateCustomFilter(ctx *Context, dev config.CustomFilterCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := agent.MainConfig.Database.UpdateCustomFilterCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteCustomFilter --pending--
func DeleteCustomFilter(ctx *Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := agent.MainConfig.Database.DelCustomFilterCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Filter %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetCustomFilterByID --pending--
func GetCustomFilterByID(ctx *Context) {
	id := ctx.Params(":id")
	dev, err := agent.MainConfig.Database.GetCustomFilterCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Filter  for device %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetCustomFiltersAffectOnDel --pending--
func GetCustomFiltersAffectOnDel(ctx *Context) {
	id := ctx.Params(":id")
	obarray, err := agent.MainConfig.Database.GetCustomFilterCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for Measurement filters %s  , error: %s", id, err)
		ctx.JSON(404, err.Error())
	} else {
		ctx.JSON(200, &obarray)
	}
}

/****************/
/*LOGIN
/****************/

func myLoginHandler(ctx *Context, user UserLogin) {
	//fmt.Printf("USER LOGIN: USER: +%#v (Config: %#v)", user, confHTTP)
	if user.UserName == confHTTP.AdminUser && user.Password == confHTTP.AdminPassword {
		ctx.SignedInUser = user.UserName
		ctx.IsSignedIn = true
		ctx.Session.Set(SESS_KEY_USERID, user.UserName)
		log.Println("Admin login OK")
		ctx.JSON(200, cookie)
	} else {
		log.Println("Admin login ERROR")
		ctx.JSON(400, "ERROR user or password not match")
	}
}

func myLogoutHandler(ctx *Context) {
	log.Printf("USER LOGOUT: USER: +%#v ", ctx.SignedInUser)
	ctx.Session.Destory(ctx)
	//ctx.Redirect("/login")
}
