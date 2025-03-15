package request

import (
	"bufio"
	"fmt"
	"io"
)

type Method string

const (
	GET Method = "GET"
)

type Request struct {
	Method  Method
	Path    string
	Headers map[string]string
	Body    []byte
	Version string
}

func NewRequest(method Method, path string, headers map[string]string, body []byte) *Request {
	return &Request{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}

func ParseRequest(reader io.Reader) *Request {
	rdr := bufio.NewReader(reader)

	requestLine, err := rdr.ReadString('\n')
	if len(requestLine) < 2 {
		return nil
	}
	if requestLine[len(requestLine)-2] != '\r' || requestLine[len(requestLine)-1] != '\n' {
		return nil
	}
	requestLine = requestLine[:len(requestLine)-2]

	if err != nil {
		return nil
	}
	var method Method
	var path string
	var version string
	_, err = fmt.Sscanf(requestLine, "%s %s %s", &method, &path, &version)
	if err != nil {
		return nil
	}

	// TODO: Parse headers and body
	headers := make(map[string]string)
	body := []byte{}

	return &Request{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
		Version: version,
	}
}
