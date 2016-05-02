package main

import (
	"fmt"
	"html/template"
	//"log"
	"net/http"
	//"time"

	//"github.com/auth0/go-jwt-middleware"
	//"github.com/dgrijalva/jwt-go"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	//"github.com/go-macaron/gzip"
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
	//"github.com/go-macaron/auth"

	// gorm database ORM
	_ "github.com/go-sql-driver/mysql"
	//	"github.com/jinzhu/gorm"
)

/*
type Person struct {
	Name string
	Age  int
	Time string
}

type ContactEntry struct {
	ID             uint   `gorm:"primary_key"`
	Name           string `sql:"type:varchar(255)"`
	Email          string `sql:"type:varchar(255);unique_index"`
	Message        string `sql:"type:text"`
	MailingAddress string `sql:"type:varchar(255)"`
}

type ContactForm struct {
	Name           string `form:"name" binding:"Required"`
	Email          string `form:"email"`
	Message        string `form:"message" binding:"Required"`
	MailingAddress string `form:"mailing_address"`
}
*/

type HTTPConfig struct {
	Port          int    `toml:"port"`
	AdminUser     string `toml:"adminuser"`
	AdminPassword string `toml:"adminpassword"`
}

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
		m.Get("/:id", GetSNMPDeviceById)
	})

	log.Printf("Server is running on localhost:%d...", port)
	http_server := fmt.Sprintf("0.0.0.0:%d", port)
	log.Println(http.ListenAndServe(http_server, m))
}

//snmpmDevices

func GetSNMPDevices(ctx *macaron.Context) {
	ctx.JSON(200, &cfg.SnmpDevice)
}

func AddSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
}

func UpdateSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
}

func DeleteSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
}

func GetSNMPDeviceById(ctx *macaron.Context) {
}

func myHandler(ctx *macaron.Context) {
	ctx.Data["Name"] = "Person"
	ctx.HTML(200, "hello") // 200 is the response code.
}

func myOtherHandler(ctx *macaron.Context) {
	ctx.Data["Message"] = "the request path is: " + ctx.Req.RequestURI
	ctx.HTML(200, "welcome")
}

/*
func myJsonHandler(ctx *macaron.Context) {
	t := time.Now()
	p := Person{"James", 25, fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())}
	ctx.JSON(200, &p)
}

func myQueryStringHandler(ctx *macaron.Context) {
	ctx.Data["Name"] = ctx.QueryEscape("name")
	ctx.HTML(200, "hello")
}
*/
func myLoginHandler(ctx *macaron.Context, user UserLogin) {
	fmt.Printf("USER LOGIN: +%v", user)
	if user.UserName == "toni" && user.Password == "toni" {
		fmt.Println("OK")
		ctx.JSON(200, "OK")
	} else {
		fmt.Println("ERROR")
		ctx.JSON(404, "ERROR")
	}
}

/*
func mySubmitHandler(ctx *macaron.Context, contact ContactForm) {
	submission := ContactEntry{0, contact.Name, contact.Email, contact.Message, contact.MailingAddress}
	// ctx.JSON(200, &submission)
	ctx.Data["Submission"] = &submission
	ctx.HTML(200, "success")
}

func mySessionHandler(sess session.Store) string {
	sess.Set("session", "session middleware")
	return sess.Get("session").(string)
}

func mySetCookieHandler(ctx *macaron.Context) string {
	// set the cookie for 5 minutes
	ctx.SetCookie("user", ctx.Params(":value"), 300)
	return "cookie set for 5 minutes"
}

func myGetCookieHandler(ctx *macaron.Context) string {
	name := ctx.GetCookie("user")
	if name == "" {
		name = "no cookie set"
	}
	return name
}

func myDatabaseHandler(ctx *macaron.Context) string {
	db, err := gorm.Open("mysql", "gorm:gorm@/gorm?charset=utf8&parseTime=True&loc=Local")

	if err != nil {
		log.Println("Database Error")
		return fmt.Sprintf("%s", err)
	}

	db.DB()

	t := time.Now()

	entry := ContactEntry{
		Name:           "James",
		Email:          "james2doyle@gmail.com",
		Message:        fmt.Sprintf("The time is %02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()),
		MailingAddress: "998 Oxford Street East, London ON, N5Y3K7",
	}

	db.Create(&entry)

	return fmt.Sprintf("New Entry Id: %d\n", entry.ID)
}

func myDatabaseListHandler(ctx *macaron.Context) {
	db, err := gorm.Open("mysql", "gorm:gorm@/gorm?charset=utf8&parseTime=True&loc=Local")

	if err != nil {
		log.Println("Database Error")
	}

	db.DB()

	contact_entries := []ContactEntry{}
	db.Find(&contact_entries)

	ctx.JSON(200, contact_entries)
}

func myCacheWriteHandler(ctx *macaron.Context, c cache.Cache) string {
	c.Put(ctx.Params(":key"), ctx.Params(":value"), 300)
	return "cached for 5 minutes"
}

func myCacheReadHandler(ctx *macaron.Context, c cache.Cache) interface{} {
	val := c.Get(ctx.Params(":key"))
	if val != nil {
		return val
	} else {
		return fmt.Sprintf("no cache with \"%s\" set", ctx.Params(":key"))
	}
}*/
