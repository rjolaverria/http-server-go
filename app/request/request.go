package request

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Method string

const (
	Get  Method = "GET"
	Post Method = "POST"
)

type Request struct {
	Method     Method
	Path       string
	PathParams map[string]string
	Headers    map[string]string
	Body       []byte
	Version    string
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

	headers := readHeaders(rdr)
	body := readBody(rdr, headers)

	return &Request{
		Method:     method,
		Path:       path,
		PathParams: make(map[string]string),
		Headers:    headers,
		Body:       body,
		Version:    version,
	}
}

func readHeaders(rdr *bufio.Reader) map[string]string {
	headers := make(map[string]string)

	for {
		headerLine, err := rdr.ReadString('\n')
		if err != nil {
			break
		}
		if len(headerLine) < 2 {
			break
		}
		if headerLine[len(headerLine)-2] != '\r' || headerLine[len(headerLine)-1] != '\n' {
			break
		}
		headerLine = headerLine[:len(headerLine)-2]
		if headerLine == "" {
			break
		}

		parts := strings.SplitN(headerLine, ":", 2)
		if len(parts) != 2 {
			break
		}

		key := parts[0]
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}

	return headers
}

func readBody(rdr *bufio.Reader, headers map[string]string) []byte {
	contentLength, exists := headers["Content-Length"]
	if !exists {
		return nil
	}

	length, err := strconv.Atoi(contentLength)
	if err != nil {
		return nil
	}

	body := make([]byte, length)
	n, err := io.ReadFull(rdr, body)
	if err != nil || n != length {
		return nil
	}

	return body
}
