package main

import (
    "log"
    "fmt"
    "time"
    "net/http"
    "html/template"

    "github.com/Unknwon/macaron"
    "github.com/macaron-contrib/binding"
    "github.com/macaron-contrib/session"
    "github.com/macaron-contrib/cache"

    // gorm database ORM
    "github.com/jinzhu/gorm"
    _ "github.com/go-sql-driver/mysql"
)

type Person struct {
    Name string
    Age  int
    Time  string
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

func main() {
    // initiate the app
    m := macaron.Classic()

    // register middleware
    m.Use(macaron.Recovery())
    m.Use(macaron.Gziper())
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
        Provider:       "memory",
        // Provider configuration, it's corresponding to provider.
        ProviderConfig: "",
        // Cookie name to save session ID. Default is "MacaronSession".
        CookieName:     "MacaronSession",
        // Cookie path to store. Default is "/".
        CookiePath:     "/",
        // GC interval time in seconds. Default is 3600.
        Gclifetime:     3600,
        // Max life time in seconds. Default is whatever GC interval time is.
        Maxlifetime:    3600,
        // Use HTTPS only. Default is false.
        Secure:         false,
        // Cookie life time. Default is 0.
        CookieLifeTime: 0,
        // Cookie domain name. Default is empty.
        Domain:         "",
        // Session ID length. Default is 16.
        IDLength:       16,
        // Configuration section name. Default is "session".
        Section:        "session",
    }))
    m.Use(macaron.Renderer(macaron.RenderOptions{
        // Directory to load templates. Default is "templates".
        Directory: "templates",
        // Extensions to parse template files from. Defaults are [".tmpl", ".html"].
        Extensions: []string{".tmpl", ".html"},
        // Funcs is a slice of FuncMaps to apply to the template upon compilation. Default is [].
        Funcs: []template.FuncMap{map[string]interface{}{
            "AppName": func() string {
                return "Macaron"
            },
            "AppVer": func() string {
                return "1.0.0"
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
    m.Get("/", myHandler)
    m.Get("/welcome", myOtherHandler)
    m.Get("/query", myQueryStringHandler) // /query?name=Some+Name
    m.Get("/json", myJsonHandler)
    m.Post("/contact/submit", binding.Bind(ContactForm{}), mySubmitHandler)
    m.Get("/session", mySessionHandler)
    m.Get("/set/cookie/:value", mySetCookieHandler)
    m.Get("/get/cookie", myGetCookieHandler)
    m.Get("/database", myDatabaseHandler)
    m.Get("/database/list", myDatabaseListHandler)
    m.Get("/cache/write/:key/:value", myCacheWriteHandler)
    m.Get("/cache/read/:key", myCacheReadHandler)

    log.Println("Server is running on localhost:4000...")
    log.Println(http.ListenAndServe("0.0.0.0:4000", m))
}

func myHandler(ctx *macaron.Context) {
    ctx.Data["Name"] = "Person"
    ctx.HTML(200, "hello") // 200 is the response code.
}

func myOtherHandler(ctx *macaron.Context) {
    ctx.Data["Message"] = "the request path is: " + ctx.Req.RequestURI
    ctx.HTML(200, "welcome")
}

func myJsonHandler(ctx *macaron.Context) {
    t := time.Now()
    p := Person{"James", 25, fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())}
    ctx.JSON(200, &p)
}

func myQueryStringHandler(ctx *macaron.Context) {
    ctx.Data["Name"] = ctx.QueryEscape("name")
    ctx.HTML(200, "hello")
}

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
        Name: "James",
        Email: "james2doyle@gmail.com",
        Message: fmt.Sprintf("The time is %02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()),
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
}