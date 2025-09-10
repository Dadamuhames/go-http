package headers

import (
	"bytes"
	"fmt"
	"maps"
	"strconv"
	"strings"
)

type Headers struct {
	headers map[string]string
}

func (h *Headers) Get(key string) string {
	return h.headers[strings.ToLower(key)]
}

func (h *Headers) Set(key string, value string) {
	h.headers[strings.ToLower(key)] = value
}

func (h *Headers) Contains(key string) bool {
	return h.headers[strings.ToLower(key)] != ""
}

func (h *Headers) Extend(headers Headers) {
	for k := range maps.Keys(headers.GetHeaders()) {
		h.Set(k, headers.Get(k))
	}
}

func (h Headers) GetHeaders() map[string]string {
	return h.headers
}

func (h *Headers) Delete(key string) {
	delete(h.headers, strings.ToLower(key))
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func GetDefaultHeaders(contentLen int) Headers {
	headers := NewHeaders()

	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return *headers
}

func (h Headers) ToString() string {
	headersString := ""

	for key, value := range h.headers {
		headersString += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	headersString += "\r\n"

	return headersString
}

func isValidToken(key []byte) bool {
	for _, ch := range key {
		present := false

		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			present = true
		} else {
			switch ch {
			case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
				present = true
			}
		}

		if !present {
			return false
		}
	}

	return true
}

var CRLF = "\r\n"

var MALFORMED_FIELD_LINE = fmt.Errorf("Malformed field line")
var MALFORMED_FIELD_NAME = fmt.Errorf("Malformed field name")

func parseHeader(fieldLine []byte) (string, string, error) {
	colonIdx := bytes.Index(fieldLine, []byte(":"))

	if colonIdx == -1 {
		return "", "", MALFORMED_FIELD_LINE
	}

	fieldName := fieldLine[:colonIdx]
	fieldValue := bytes.TrimSpace(fieldLine[colonIdx+len([]byte(":")):])

	if bytes.HasSuffix(fieldName, []byte(" ")) {
		return "", "", MALFORMED_FIELD_NAME
	}

	return string(fieldName), string(fieldValue), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], []byte(CRLF))

		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(CRLF)
			break
		}

		fieldLine := data[read : read+idx]

		name, value, err := parseHeader(fieldLine)

		if err != nil {
			return 0, false, err
		}

		if !isValidToken([]byte(name)) {
			return 0, false, MALFORMED_FIELD_NAME
		}

		if h.Contains(name) {
			h.Set(name, fmt.Sprintf("%s, %s", h.Get(name), value))
		} else {
			h.Set(name, value)
		}

		read += idx + len(CRLF)
	}

	return read, done, nil
}
