package router

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

type RequestHandler func(*request.Request) *response.Response

type RouteNode struct {
	path     string
	param    string
	isWild   bool
	handlers map[request.Method]*RequestHandler
	children []*RouteNode
}

type Router struct {
	root *RouteNode
}

func NewRouter() *Router {
	return &Router{
		root: &RouteNode{
			path:     "/",
			handlers: make(map[request.Method]*RequestHandler),
			children: []*RouteNode{},
		},
	}
}

func (r *Router) Get(path string, handler RequestHandler) {
	r.registerHandler(request.Get, path, &handler)
}

func (r *Router) Post(path string, handler RequestHandler) {
	r.registerHandler(request.Post, path, &handler)
}

func (r *Router) Handle(conn net.Conn) {
	req := request.ParseRequest(conn)

	if req == nil {
		res := response.NewResponse(response.InternalServerError, nil, nil)
		res.Write(conn)
		return
	}

	params := make(map[string]string)
	node, found := r.findNode(req.Path, params)

	if node == nil || !found {
		res := response.NewResponse(response.NotFound, nil, nil)
		res.Write(conn)
		return
	}

	handler, found := node.handlers[req.Method]
	if handler == nil || !found {
		res := response.NewResponse(response.MethodNotAllowed, nil, nil)
		res.Write(conn)
		return
	}

	req.PathParams = params
	res := (*handler)(req)

	handleEncoding(req, res)

	res.Headers["Content-Length"] = fmt.Sprintf("%d", len(*res.Body))

	res.Write(conn)
}

func handleEncoding(req *request.Request, res *response.Response) {
	reqHeader, ok := req.Headers["Accept-Encoding"]

	if !ok || reqHeader == "" {
		return
	}

	encodings := strings.SplitSeq(reqHeader, ",")
	for encoding := range encodings {
		if strings.TrimSpace(encoding) == "gzip" {
			res.Headers["Content-Encoding"] = "gzip"

			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			defer gz.Close()
			gz.Write(*res.Body)
			compressed := buf.Bytes()
			res.Body = &compressed
			break
		}
	}
}

func (r *Router) findNode(path string, params map[string]string) (*RouteNode, bool) {
	if path == "" || path == "/" {
		return r.root, true
	}

	if path[0] != '/' {
		path = "/" + path
	}

	current := r.root
	segments := strings.Split(path, "/")[1:]

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		matched := false
		for _, child := range current.children {
			if !child.isWild && strings.TrimPrefix(child.path, "/") == segment {
				current = child
				matched = true
				break
			}
		}

		if !matched {
			for _, child := range current.children {
				if child.isWild {
					params[child.param] = segment
					current = child
					matched = true
					break
				}
			}
		}

		if !matched {
			return nil, false
		}
	}

	return current, true
}

func (r *Router) registerHandler(method request.Method, path string, handler *RequestHandler) {
	r.insertRoute(method, path, handler)
}

func (r *Router) insertRoute(method request.Method, path string, handler *RequestHandler) {
	if path == "" || path == "/" {
		if _, exists := r.root.handlers[method]; exists {
			panic(fmt.Sprintf("Route already exists: %s %s", method, path))
		}
		r.root.handlers[method] = handler
		return
	}

	if path[0] != '/' {
		path = "/" + path
	}

	segments := strings.Split(path, "/")[1:]
	current := r.root

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		matched := false
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			paramName := segment[1 : len(segment)-1]
			for _, child := range current.children {
				if child.isWild {
					if child.param != paramName {
						panic(fmt.Sprintf("Conflicting parameter names: %s and %s", child.param, paramName))
					}
					current = child
					matched = true
					break
				}
			}

			if !matched {
				child := &RouteNode{
					path:     "/" + segment,
					isWild:   true,
					param:    paramName,
					handlers: make(map[request.Method]*RequestHandler),
					children: []*RouteNode{},
				}
				current.children = append(current.children, child)
				current = child
			}
		} else {
			segmentPath := "/" + segment
			for _, child := range current.children {
				if !child.isWild && child.path == segmentPath {
					current = child
					matched = true
					break
				}
			}

			if !matched {
				child := &RouteNode{
					path:     segmentPath,
					handlers: make(map[request.Method]*RequestHandler),
					children: []*RouteNode{},
				}
				current.children = append(current.children, child)
				current = child
			}
		}
	}

	if _, exists := current.handlers[method]; exists {
		panic(fmt.Sprintf("Route already exists: %s %s", method, path))
	}
	current.handlers[method] = handler
}

func (r *Router) PrintTree() {
	fmt.Println("\nRouter Tree:")
	r.printNode(r.root, 0)
}

func (r *Router) printNode(node *RouteNode, level int) {
	indent := strings.Repeat("  ", level)

	fmt.Printf("%s[Node] Path: %s", indent, node.path)
	if node.isWild {
		fmt.Printf(" (Param: %s)", node.param)
	}

	if len(node.handlers) > 0 {
		fmt.Printf(" Methods: [")
		methods := make([]string, 0)
		for method := range node.handlers {
			methods = append(methods, string(method))
		}
		fmt.Printf("%s]", strings.Join(methods, ", "))
	}
	fmt.Println()

	for _, child := range node.children {
		r.printNode(child, level+1)
	}
}
