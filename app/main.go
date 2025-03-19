package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
	"github.com/codecrafters-io/http-server-starter-go/app/router"
)

func main() {
	directory := flag.String("directory", ".", "directory to serve")
	flag.Parse()

	fmt.Println("Serving directory:", *directory)

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	router := router.NewRouter()
	router.GET("/", func(*request.Request) *response.Response {
		return response.NewResponse(response.OK, nil, nil)
	})
	router.GET("/echo/{str}", func(req *request.Request) *response.Response {
		headers := make(map[string]string)
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = fmt.Sprintf("%d", len(req.PathParams["str"]))
		return response.NewResponse(response.OK, headers, []byte(req.PathParams["str"]))
	})
	router.GET("/user-agent", func(req *request.Request) *response.Response {
		headers := make(map[string]string)
		headers["Content-Type"] = "text/plain"
		headers["Content-Length"] = fmt.Sprintf("%d", len(req.Headers["User-Agent"]))
		return response.NewResponse(response.OK, headers, []byte(req.Headers["User-Agent"]))
	})

	router.GET("/files/{filepath}", func(req *request.Request) *response.Response {
		filepath := strings.TrimPrefix(req.PathParams["filepath"], "/")
		fullPath := fmt.Sprintf("%s%c%s", *directory, os.PathSeparator, filepath)
		file, err := os.ReadFile(fullPath)
		if err != nil {
			return response.NewResponse(response.NotFound, nil, nil)
		}

		headers := make(map[string]string)
		headers["Content-Type"] = "application/octet-stream"
		headers["Content-Length"] = fmt.Sprintf("%d", len(file))

		return response.NewResponse(response.OK, headers, file)
	})

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go func() {
			defer conn.Close()
			router.Handle(conn)
		}()
	}
}
