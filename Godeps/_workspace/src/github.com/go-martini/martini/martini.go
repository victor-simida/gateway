// Package martini is a powerful package for quickly writing modular web applications/services in Golang.
//
// For a full guide visit http://github.com/go-martini/martini
//
//  package main
//
//  import "github.com/go-martini/martini"
//
//  func main() {
//    m := martini.Classic()
//
//    m.Get("/", func() string {
//      return "Hello world!"
//    })
//
//    m.Run()
//  }
package martini

import (
	"log"
	"net/http"
	"os"
	"reflect"

	"gateway/Godeps/_workspace/src/github.com/codegangsta/inject"
	"net"
	"sync"
	"time"
)

// Martini represents the top level web application. inject.Injector methods can be invoked to map services on a global level.
type Martini struct {
	inject.Injector
	handlers []Handler
	action   Handler
	logger   *log.Logger
	server   MartiniServer
}

type connectState struct {
	conn  net.Conn
	state http.ConnState
}

type MartiniServer struct {
	l            net.Listener
	stateMap     map[net.Conn]http.ConnState
	stateChannel chan connectState
	connsLock    sync.Locker
}


// New creates a bare bones Martini instance. Use this method if you want to have full control over the middleware that is used.
func New() *Martini {
	m := &Martini{Injector: inject.New(), action: func() {}, logger: log.New(os.Stdout, "[martini] ", 0)}
	m.server.stateMap = make(map[net.Conn]http.ConnState)
	m.server.stateChannel = make(chan connectState, 9192)

	m.Map(m.logger)
	m.Map(defaultReturnHandler())
	return m
}

// Handlers sets the entire middleware stack with the given Handlers. This will clear any current middleware handlers.
// Will panic if any of the handlers is not a callable function
func (m *Martini) Handlers(handlers ...Handler) {
	m.handlers = make([]Handler, 0)
	for _, handler := range handlers {
		m.Use(handler)
	}
}

// Action sets the handler that will be called after all the middleware has been invoked. This is set to martini.Router in a martini.Classic().
func (m *Martini) Action(handler Handler) {
	validateHandler(handler)
	m.action = handler
}

// Use adds a middleware Handler to the stack. Will panic if the handler is not a callable func. Middleware Handlers are invoked in the order that they are added.
func (m *Martini) Use(handler Handler) {
	validateHandler(handler)

	m.handlers = append(m.handlers, handler)
}

// ServeHTTP is the HTTP Entry point for a Martini instance. Useful if you want to control your own HTTP server.
func (m *Martini) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m.createContext(res, req).run()
}

// Run the http server on a given host and port.
func (m *Martini) RunOnAddr(addr string) {
	// TODO: Should probably be implemented using a new instance of http.Server in place of
	// calling http.ListenAndServer directly, so that it could be stored in the martini struct for later use.
	// This would also allow to improve testing when a custom host and port are passed.

	logger := m.Injector.Get(reflect.TypeOf(m.logger)).Interface().(*log.Logger)
	logger.Printf("listening on %s (%s)\n", addr, Env)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalln(err.Error())
		return
	}
	var server http.Server
	server.Handler = m
	server.ConnState = m.connectStateChange
	m.server.l = l

	go func() {
		channel := m.server.stateChannel
		for input := range channel {
			switch input.state {
			case http.StateNew, http.StateActive, http.StateIdle:
				m.server.connsLock.Lock()
				m.server.stateMap[input.conn] = input.state
				m.server.connsLock.Unlock()
			case http.StateHijacked, http.StateClosed:
				m.server.connsLock.Lock()
				delete(m.server.stateMap, input.conn)
				m.server.connsLock.Unlock()
			}
		}
	}()
	server.Serve(l)
}

// Run the http server. Listening on os.GetEnv("PORT") or 3000 by default.
func (m *Martini) Run() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	host := os.Getenv("HOST")

	m.RunOnAddr(host + ":" + port)
}

func (m *Martini) createContext(res http.ResponseWriter, req *http.Request) *context {
	c := &context{inject.New(), m.handlers, m.action, NewResponseWriter(res), 0}
	c.SetParent(m)
	c.MapTo(c, (*Context)(nil))
	c.MapTo(c.rw, (*http.ResponseWriter)(nil))
	c.Map(req)
	return c
}

func (m *Martini)closeIdleConn() {
	m.server.connsLock.Lock()
	for k, v := range m.server.stateMap {
		if v == http.StateIdle {
			delete(m.server.stateMap, k)
		}
	}
	m.server.connsLock.Unlock()
}

func (m *Martini) Close() {
	if m.server.l != nil {
		m.server.l.Close()
		m.server.l = nil
	}
	for {
		length := len(m.server.stateMap)
		if length == 0 {
			break
		}
		m.closeIdleConn()
		time.Sleep(500 * time.Millisecond)
	}

	if m.server.stateChannel != nil {
		close(m.server.stateChannel)
		m.server.stateChannel = nil
	}

}

func (m *Martini)connectStateChange(conn net.Conn, state http.ConnState) {
	var input connectState = connectState{conn:conn, state:state}
	select {
	case m.server.stateChannel <- input:
	default:
		logger := m.Injector.Get(reflect.TypeOf(m.logger)).Interface().(*log.Logger)
		logger.Printf("stateChannel full\n")
	}
}

// ClassicMartini represents a Martini with some reasonable defaults. Embeds the router functions for convenience.
type ClassicMartini struct {
	*Martini
	Router
}

// Classic creates a classic Martini with some basic default middleware - martini.Logger, martini.Recovery and martini.Static.
// Classic also maps martini.Routes as a service.
func Classic() *ClassicMartini {
	r := NewRouter()
	m := New()
	m.Use(Logger())
	m.Use(Recovery())
	m.Use(Static("public"))
	m.MapTo(r, (*Routes)(nil))
	m.Action(r.Handle)
	return &ClassicMartini{m, r}
}

// Handler can be any callable function. Martini attempts to inject services into the handler's argument list.
// Martini will panic if an argument could not be fullfilled via dependency injection.
type Handler interface{}

func validateHandler(handler Handler) {
	if reflect.TypeOf(handler).Kind() != reflect.Func {
		panic("martini handler must be a callable func")
	}
}

// Context represents a request context. Services can be mapped on the request level from this interface.
type Context interface {
	inject.Injector
	// Next is an optional function that Middleware Handlers can call to yield the until after
	// the other Handlershave been executed. This works really well for any operations that must
	// happen after an http request
	Next()
	// Written returns whether or not the response for this context has been written.
	Written() bool
}

type context struct {
	inject.Injector
	handlers []Handler
	action   Handler
	rw       ResponseWriter
	index    int
}

func (c *context) handler() Handler {
	if c.index < len(c.handlers) {
		return c.handlers[c.index]
	}
	if c.index == len(c.handlers) {
		return c.action
	}
	panic("invalid index for context handler")
}

func (c *context) Next() {
	c.index += 1
	c.run()
}

func (c *context) Written() bool {
	return c.rw.Written()
}

func (c *context) run() {
	for c.index <= len(c.handlers) {
		_, err := c.Invoke(c.handler())
		if err != nil {
			panic(err)
		}
		c.index += 1

		if c.Written() {
			return
		}
	}
}

