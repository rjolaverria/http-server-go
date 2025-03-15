package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
	"github.com/codecrafters-io/http-server-starter-go/app/router"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	router := router.NewRouter()
	router.AddRoute(request.GET, "/", func(*request.Request) *response.Response {
		return response.NewResponse(response.OK, nil, nil)
	})

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		handleConnection(conn, router)
	}
}

func handleConnection(conn net.Conn, router *router.Router) {
	defer conn.Close()

	req := request.ParseRequest(conn)

	if req == nil {
		res := response.NewResponse(response.InternalServerError, nil, []byte("Internal Error"))
		res.Write(conn)
		return
	}

	route, found := router.GetRoute(req.Method, req.Path)
	if found {
		res := route.RequestHandler(req)
		res.Write(conn)
		return
	}

	res := response.NewResponse(response.NotFound, nil, []byte("Not Found\n"))
	res.Write(conn)
}
