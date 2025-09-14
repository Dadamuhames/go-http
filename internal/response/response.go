package response

import (
	"crypto/sha256"
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
	WriteTrailers   WriterState = 2
	WriteDone       WriterState = 3
)

type ResponseWriter struct {
	writer     io.Writer
	statusCode StatusCode
	headers    headers.Headers
	body       []byte
	state      WriterState
	trailers   headers.Headers
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

func (w ResponseWriter) SendFromStream(
	statusCode StatusCode,
	reader io.ReadCloser) {

	// write status line
	w.SetStatusCode(statusCode)
	w.writeStatusLine()

	// write headers
	w.headers.Delete("Content-Length")
	w.headers.Set("Transfer-Encoding", "chunked")
	w.headers.Set("Content-Type", "text/plain")
	w.headers.Set("Trailers", "X-Content-SHA256, X-Content-Length")
	w.writeHeaders()

	// write body
	data := make([]byte, 1024)
	hasher := sha256.New()
	contentLen := 0

	for {
		n, err := reader.Read(data)

		if err != nil {
			fmt.Println(err)
			break
		}

		chunk := data[:n]

		w.writeBodyFromByteArray([]byte(fmt.Sprintf("%x\r\n", n)))
		w.writeBodyFromByteArray(chunk)
		w.writeBodyFromByteArray([]byte("\r\n"))

		hasher.Write(chunk)
		contentLen += n
	}

	w.writeBodyFromByteArray([]byte("0\r\n\r\n"))

	trailers := headers.NewHeaders()

	sum := hasher.Sum(nil)
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sum))
	trailers.Set("X-Content-Length", strconv.Itoa(contentLen))

	w.state = WriteTrailers
	w.trailers = *trailers

	w.writeTrailers()
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

func (w *ResponseWriter) writeBodyFromByteArray(body []byte) (int, error) {
	if w.state != WriteBody {
		return 0, fmt.Errorf("Invalid state to write body")
	}

	return w.writer.Write(body)
}

func (w *ResponseWriter) writeTrailers() error {
	if w.state != WriteTrailers {
		return fmt.Errorf("Invalid state to write trailers")
	}

	if len(w.trailers.GetHeaders()) == 0 {
		w.state = WriteDone
		return nil
	}

	trailerString := w.trailers.ToString()

	_, err := w.writer.Write([]byte(trailerString))

	w.state = WriteDone

	return err

}
