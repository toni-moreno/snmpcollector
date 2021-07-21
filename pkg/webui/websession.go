package webui

import (
	"math/rand"
	"time"

	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
)

// SessKeyUserID type of session key
const (
	SessKeyUserID = "uid"
)

var sessionManager *session.Manager
var sessionOptions *session.Options
var startSessionGC func()
var getSessionCount func() int

func init() {
	startSessionGC = func() {
		sessionManager.GC()
		time.AfterFunc(time.Duration(sessionOptions.Gclifetime)*time.Second, startSessionGC)
	}
	getSessionCount = func() int {
		return sessionManager.Count()
	}
}

func prepareOptions(opt *session.Options) *session.Options {
	if len(opt.Provider) == 0 {
		opt.Provider = "memory"
	}
	if len(opt.ProviderConfig) == 0 {
		opt.ProviderConfig = "data/sessions"
	}
	if len(opt.CookieName) == 0 {
		opt.CookieName = "snmpcollector-session"
	}
	if len(opt.CookiePath) == 0 {
		opt.CookiePath = "/"
	}
	if opt.Gclifetime == 0 {
		opt.Gclifetime = 3600
	}
	if opt.Maxlifetime == 0 {
		opt.Maxlifetime = opt.Gclifetime
	}
	if opt.IDLength == 0 {
		opt.IDLength = 16
	}

	return opt
}

// Sessioner return a Macaron handler
func Sessioner(options session.Options) macaron.Handler {
	var err error
	sessionOptions = prepareOptions(&options)
	sessionManager, err = session.NewManager(options.Provider, options)
	if err != nil {
		panic(err)
	}

	// start GC threads after some random seconds
	rndSeconds := 10 + rand.Int63n(180)
	time.AfterFunc(time.Duration(rndSeconds)*time.Second, startSessionGC)

	return func(ctx *Context) {
		ctx.Next()

		if err := ctx.Session.Release(); err != nil {
			panic("session(release): " + err.Error())
		}
	}
}

// GetSession get session store
func GetSession() SessionStore {
	return &SessionWrapper{manager: sessionManager}
}

// SessionStore Session Interface
type SessionStore interface {
	// Set sets value to given key in session.
	Set(interface{}, interface{}) error
	// Get gets value by given key in session.
	Get(interface{}) interface{}
	// ID returns current session ID.
	ID() string
	// Release releases session resource and save data to provider.
	Release() error
	// Destory deletes a session.
	Destory(*Context) error
	// init
	Start(*Context) error
}

// SessionWrapper session wrapper
type SessionWrapper struct {
	session session.RawStore
	manager *session.Manager
}

// Start start session context
func (s *SessionWrapper) Start(c *Context) error {
	var err error
	s.session, err = s.manager.Start(c.Context)
	return err
}

// Set session key/value setter
func (s *SessionWrapper) Set(k interface{}, v interface{}) error {
	if s.session != nil {
		return s.session.Set(k, v)
	}
	return nil
}

// Get session Key/Value getter
func (s *SessionWrapper) Get(k interface{}) interface{} {
	if s.session != nil {
		return s.session.Get(k)
	}
	return nil
}

// ID get session identificator
func (s *SessionWrapper) ID() string {
	if s.session != nil {
		return s.session.ID()
	}
	return ""
}

// Release session release
func (s *SessionWrapper) Release() error {
	if s.session != nil {
		return s.session.Release()
	}
	return nil
}

// Destory destory session
func (s *SessionWrapper) Destory(c *Context) error {
	if s.session != nil {
		if err := s.manager.Destory(c.Context); err != nil {
			return err
		}
		s.session = nil
	}
	return nil
}
