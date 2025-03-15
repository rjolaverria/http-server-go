package router

import (
	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

type RequestHandler func(*request.Request) *response.Response

type Route struct {
	Method         string
	Path           string
	RequestHandler func(*request.Request) *response.Response
}

type Router struct {
	Routes map[request.Method]map[string]*Route
}

func NewRouter() *Router {
	return &Router{
		Routes: make(map[request.Method]map[string]*Route),
	}
}

func (r *Router) AddRoute(method request.Method, path string, handler RequestHandler) {
	if _, exists := r.Routes[method]; !exists {
		r.Routes[method] = make(map[string]*Route)
	}
	r.Routes[method][path] = &Route{
		Method:         string(method),
		Path:           path,
		RequestHandler: handler,
	}
}

func (r *Router) GetRoute(method request.Method, path string) (*Route, bool) {
	if routes, exists := r.Routes[method]; exists {
		if route, exists := routes[path]; exists {
			return route, true
		}
	}
	return nil, false
}
