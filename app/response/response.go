package response

import (
	"fmt"
	"io"
)

type Status struct {
	Code int
	Name string
}

func (s Status) String() string {
	return fmt.Sprintf("%d %s", s.Code, s.Name)
}

var (
	OK                  = Status{Code: 200, Name: "OK"}
	NotFound            = Status{Code: 404, Name: "Not Found"}
	InternalServerError = Status{Code: 500, Name: "Internal Server Error"}
)

type Response struct {
	Status  Status
	Headers map[string]string
	Body    *[]byte
}

func NewResponse(status Status, headers map[string]string, body []byte) *Response {
	return &Response{
		Status:  status,
		Headers: headers,
		Body:    &body,
	}
}

func (r *Response) Write(writer io.Writer) error {
	response := []byte{}
	response = append(response, []byte("HTTP/1.1 "+r.Status.String()+"\r\n")...)
	for key, value := range r.Headers {
		response = append(response, []byte(key+": "+value+"\r\n")...)
	}
	response = append(response, []byte("\r\n")...)
	response = append(response, *r.Body...)

	_, err := writer.Write(response)
	if err != nil {
		return err
	}

	return nil
}
