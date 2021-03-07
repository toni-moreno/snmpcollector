package webui

import (
	"fmt"

	"github.com/toni-moreno/snmpcollector/pkg/data/utils"
	"gopkg.in/macaron.v1"
)

// Context custom context for http session handler
type Context struct {
	*macaron.Context
	SignedInUser string
	Session      SessionStore
	IsSignedIn   bool
}

func accessForbidden(ctx *Context) {
	ctx.JSON(403, fmt.Errorf("access forbidden %+v", ctx))
	//c.SetCookie("redirect_to", url.QueryEscape(c.Req.RequestURI), 0, "/")
	//c.Redirect("/login")
}

var reqSignedIn = func(ctx *Context) {
	if !ctx.IsSignedIn {
		accessForbidden(ctx)
		log.Infof("CONTEXT %+v", ctx)
		return
	}
}

func initContextWithUserSessionCookie(ctx *Context) bool {
	// initialize session
	if err := ctx.Session.Start(ctx); err != nil {
		log.Error("Failed to start session", "error", err)
		return false
	}
	userID := ctx.Session.Get(SessKeyUserID)

	if userID != nil {
		ctx.SignedInUser = ctx.Session.Get(SessKeyUserID).(string)
		ctx.IsSignedIn = true
		return true
	}

	return false

}

func initContextWithBasicAuth(ctx *Context) bool {

	header := ctx.Req.Header.Get("Authorization")
	if header == "" {
		return false
	}

	username, password, err := utils.DecodeBasicAuthHeader(header)
	if err != nil {
		ctx.JSON(401, fmt.Errorf("Invalid Basic Auth Header: %s", err))
		return true
	}

	if username == confHTTP.AdminUser && password == confHTTP.AdminPassword {
		log.Println("BASIC Auth: Admin login OK")
		ctx.SignedInUser = username
		ctx.IsSignedIn = true
	}
	return true
}

// GetContextHandler get context handler
func GetContextHandler() macaron.Handler {
	return func(c *macaron.Context) {
		ctx := &Context{
			Context:      c,
			SignedInUser: "",
			Session:      GetSession(),
			IsSignedIn:   false,
		}

		// the order in which these are tested are important
		// look for api key in Authorization header first
		// then init session and look for userId in session
		// then look for api key in session (special case for render calls via api)
		// then test if anonymous access is enabled
		switch {
		case initContextWithBasicAuth(ctx):
		case initContextWithUserSessionCookie(ctx):
			//case initContextWithAuthProxy(remoteCache, ctx, orgId):
			//case initContextWithToken(ats, ctx, orgId):
			//case initContextWithAnonymousUser(ctx):

		}

		ctx.Data["ctx"] = ctx
		c.Map(ctx)
	}
}

// RawAsJSON raw to json conversion
func (ctx *Context) RawAsJSON(status int, json []byte) {

	// json rendered fine, write out the result
	ctx.Header().Set("Content-Type", "application/json; charset=UTF-8")
	ctx.WriteHeader(status)
	ctx.Write(json)
}
