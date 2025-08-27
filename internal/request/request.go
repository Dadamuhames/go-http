package request

import (
	"bytes"
	"fmt"
	"go-http/internal/headers"
	"io"
	"strings"
)

type parserState int

const (
	StateInitialized    parserState = 0
	StateParsingHeaders parserState = 1
	StateDone           parserState = 2
)

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       parserState
}

func newRequest() *Request {
	return &Request{
		Headers: *headers.NewHeaders(),
		state:   StateInitialized,
	}
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]

		switch r.state {
		case StateInitialized:
			rLine, readN, err := parseRequestLine(currentData)

			if err != nil {
				return 0, err
			}

			if rLine != nil {
				r.RequestLine = *rLine
				read += readN
				r.state = StateParsingHeaders
			}

			if readN == 0 {
				break outer // if there is not \r\n break out to outer loop
			}

		case StateParsingHeaders:
			readN, done, err := r.Headers.Parse(currentData)

			if err != nil {
				return 0, err
			}

			if readN == 0 {
				break outer
			}

			read += readN

			if done {
				r.state = StateDone
			}

		case StateDone:
			break outer
		}
	}

	return read, nil
}

var ERROR_BAD_START_LINE = fmt.Errorf("Invalid start line")
var ERROR_HTTP_VERSION_NOT_SUPPORTED = fmt.Errorf("HTTP version not supported")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, SEPARATOR)

	if idx == -1 {
		return nil, 0, nil
	}

	rawLine := data[:idx]
	read := idx + len(SEPARATOR)

	rawLineArr := bytes.Split(rawLine, []byte(" "))

	if len(rawLineArr) != 3 {
		return nil, read, ERROR_BAD_START_LINE
	}

	method := string(rawLineArr[0])

	if strings.ToUpper(method) != method {
		return nil, read, ERROR_BAD_START_LINE
	}

	path := string(rawLineArr[1])

	httpVersion := string(rawLineArr[2])

	if httpVersion != "HTTP/1.1" {
		return nil, read, ERROR_HTTP_VERSION_NOT_SUPPORTED
	}

	versionNumber := strings.Split(httpVersion, "/")[1]

	return &RequestLine{Method: method, RequestTarget: path, HttpVersion: versionNumber}, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufIdx := 0

	for !request.done() {
		n, err := reader.Read(buf[bufIdx:])

		if err != nil {
			return nil, err
		}

		bufIdx += n

		readN, err := request.parse(buf[:bufIdx])

		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}

	return request, nil
}
