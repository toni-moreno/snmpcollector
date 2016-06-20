package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	//"github.com/go-macaron/gzip"
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
	//"github.com/go-macaron/auth"
	// gorm database ORM
	//_ "github.com/go-sql-driver/mysql"
	//	"github.com/jinzhu/gorm"
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
	m.Group("/snmpdevice", func() {
		m.Get("/", GetSNMPDevices)
		m.Post("/", bind(SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", bind(SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", DeleteSNMPDevice)
		m.Get("/:id", GetSNMPDeviceByID)
	})

	log.Printf("Server is running on localhost:%d...", port)
	httpServer := fmt.Sprintf("0.0.0.0:%d", port)
	log.Println(http.ListenAndServe(httpServer, m))
}

// GetSNMPDevices Return snmpdevice list to frontend
func GetSNMPDevices(ctx *macaron.Context) {
	ctx.JSON(200, &cfg.SnmpDevice)
}

// AddSNMPDevice Insert new snmpdevice to de internal BBDD --pending--
func AddSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
	log.Printf("ADDING DEVICE %+v", dev)
}

// UpdateSNMPDevice --pending--
func UpdateSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
}

//DeleteSNMPDevice --pending--
func DeleteSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
}

//GetSNMPDeviceByID --pending--
func GetSNMPDeviceByID(ctx *macaron.Context) {
}

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
