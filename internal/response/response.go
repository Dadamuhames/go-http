package response

import (
	"fmt"
	"go-http/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	HTTP_STATUS_OK                    StatusCode = 200
	HTTP_STATUS_BAD_REQUEST           StatusCode = 400
	HTTP_STATUS_INTERNAL_SERVER_ERROR StatusCode = 500
)

type WriterState int

const (
	WriteStatusLine WriterState = 0
	WriteHeaders    WriterState = 1
	WriteBody       WriterState = 2
	WriteDone       WriterState = 3
)

type ResponseWriter struct {
	writer     io.Writer
	statusCode StatusCode
	headers    headers.Headers
	body       []byte
	state      WriterState
}

func NewResponseWriter(writer io.Writer) *ResponseWriter {
	return &ResponseWriter{
		writer:  writer,
		headers: headers.GetDefaultHeaders(0),
		state:   WriteStatusLine,
	}
}

func (w *ResponseWriter) SetStatusCode(statusCode StatusCode) {
	w.statusCode = statusCode
}

func (w *ResponseWriter) SetHeader(key, value string) {
	w.headers.Set(key, value)
}

func (w *ResponseWriter) SetBody(body []byte) {
	w.body = body
}

func (w ResponseWriter) Send(statusCode StatusCode, hdrs headers.Headers, body []byte) {
	w.SetStatusCode(statusCode)

	w.headers.Extend(hdrs)
	w.headers.Set("Content-Length", strconv.Itoa(len(body)))

	w.SetBody(body)
	w.writeAll()
}

func (w ResponseWriter) SendBodyWithDefaultHeaders(statusCode StatusCode, body []byte) {
	w.SetStatusCode(statusCode)

	w.headers.Set("Content-Length", strconv.Itoa(len(body)))

	w.SetBody(body)
	w.writeAll()
}

func (w ResponseWriter) SendEmptyResponse(statusCode StatusCode) {
	w.SetStatusCode(statusCode)

	w.writeAll()
}

func (w ResponseWriter) writeAll() {
	w.writeStatusLine()
	w.writeHeaders()
	w.writeBody()
}

func (w *ResponseWriter) writeStatusLine() error {
	if w.state != WriteStatusLine {
		return fmt.Errorf("Invalid state to write status line")
	}

	reasonPhrase := ""

	switch w.statusCode {
	case HTTP_STATUS_OK:
		reasonPhrase = "OK"
	case HTTP_STATUS_BAD_REQUEST:
		reasonPhrase = "Bad Request"
	case HTTP_STATUS_INTERNAL_SERVER_ERROR:
		reasonPhrase = "Internal Server Error"
	}

	w.state = WriteHeaders

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", w.statusCode, reasonPhrase)

	_, err := w.writer.Write([]byte(statusLine))

	return err
}

func (w *ResponseWriter) writeHeaders() error {
	if w.state != WriteHeaders {
		return fmt.Errorf("Invalid state to write headers")
	}

	headerString := w.headers.ToString()

	_, err := w.writer.Write([]byte(headerString))

	w.state = WriteBody

	return err
}

func (w *ResponseWriter) writeBody() (int, error) {
	if w.state != WriteBody {
		return 0, fmt.Errorf("Invalid state to write body")
	}

	w.state = WriteDone

	return w.writer.Write(w.body)
}
