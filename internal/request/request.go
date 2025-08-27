package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type parserState string

const (
	stateInit    parserState = "init"
	stateDone    parserState = "done"
	stateBody    parserState = "body"
	stateHeaders parserState = "headers"
	stateError   parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string

	state parserState
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func newRequest() *Request {
	return &Request{
		state:   stateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

var (
	ErrorMalformedRequestLine   = fmt.Errorf("malformed request-line")
	ErrorUnsupportedHTTPVersion = fmt.Errorf("unsupported http version")
	ErrorRequestInErrorState    = fmt.Errorf("request in error state")
	SEPARATOR                   = []byte("\r\n")
)

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}
	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
dance:
	for {
		currentData := data[read:]
		// if len(currentData) == 0 {
		// 	break dance
		// }

		switch r.state {
		case stateError:
			return 0, ErrorRequestInErrorState

		case stateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = stateError
				return 0, err
			}
			if n == 0 {
				break dance
			}
			r.RequestLine = *rl
			read += n

			r.state = stateHeaders

		case stateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = stateError
				return 0, err
			}
			if n == 0 {
				break dance
			}

			read += n

			if done {
				// if r.hasBody() {
				// 	r.state = stateBody
				// } else {
				// 	r.state = stateDone
				// }
				r.state = stateBody
			}

		case stateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				// panic("chunck not implemented")
				r.state = stateDone
				break dance

			}

			remaining := min(length-len(r.Body), len(currentData))
			if remaining == 0 {
				break dance
			}
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = stateDone
			}

		case stateDone:
			break dance

		default:
			panic("somehow we have programmed poorly")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == stateDone || r.state == stateError
}

// func (r *Request) hasBody() bool {
// 	length := getInt(r.Headers, "content-length", 0)
// 	return length > 0
// }

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
