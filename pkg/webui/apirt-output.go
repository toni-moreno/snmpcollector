package webui

import (
	//"github.com/go-macaron/binding"

	"github.com/toni-moreno/snmpcollector/pkg/agent"
	"gopkg.in/macaron.v1"
)

// NewAPIRtDevice Runtime Device REST API creator
func NewAPIRtOutput(m *macaron.Macaron) error {

	//bind := binding.Bind

	m.Group("/api/rt/output", func() {
		m.Get("/info/", reqSignedIn, RTGetOutputInfo)
		m.Get("/info/:id", reqSignedIn, RTGetOutputInfo)
		m.Get("/buffer/:id/:action", reqSignedIn, RTOutputBufferAction)
	})

	return nil
}

/****************/
/*Runtime Info
/****************/

// RTOutputBufferReset runtime send reset
func RTOutputBufferAction(ctx *Context) {
	id := ctx.Params(":id")
	action := ctx.Params(":action")
	log.Infof("apply action: %s on output %s runtime", action, id)
	out, err := agent.GetOutput(id)
	if err != nil {
		ctx.JSON(404, err.Error())
		return
	}
	out.Action(action)
	ctx.JSON(200, "OK")
}

//RTGetOutputInfo xx
func RTGetOutputInfo(ctx *Context) {
	id := ctx.Params(":id")
	if len(id) > 0 {
		json, err := agent.GetOutputJSONInfo(id)
		if err != nil {
			ctx.JSON(404, err.Error())
			return
		}
		log.Infof("get runtime data from id %s", id)
		ctx.RawAsJSON(200, json)

		//get only one device info
	} else {
		outputs := agent.GetOutputStats()
		ctx.JSON(200, &outputs)
	}
	return
}
