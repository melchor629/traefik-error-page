package helpers

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// CodeModifier forwards a response back to the client,
// while enforcing a given response code.
type CodeModifier struct {
	code int // the code enforced in the response.

	// headerSent is whether the headers have already been sent,
	// either through Write or WriteHeader.
	headerSent bool
	headerMap  http.Header // the HTTP response headers from the backend.

	responseWriter http.ResponseWriter
}

// NewCodeModifier returns a codeModifier that enforces the given code.
func NewCodeModifier(rw http.ResponseWriter, code int) *CodeModifier {
	return &CodeModifier{
		headerMap:      make(http.Header),
		code:           code,
		responseWriter: rw,
	}
}

// Header returns the response headers.
func (r *CodeModifier) Header() http.Header {
	if r.headerSent {
		return r.responseWriter.Header()
	}

	if r.headerMap == nil {
		r.headerMap = make(http.Header)
	}

	return r.headerMap
}

// Write calls WriteHeader to send the enforced code,
// then writes the data directly to r.responseWriter.
func (r *CodeModifier) Write(buf []byte) (int, error) {
	r.WriteHeader(r.code)
	return r.responseWriter.Write(buf)
}

// WriteHeader sends the headers, with the enforced code (the code in argument is always ignored),
// if it hasn't already been done.
// WriteHeader is, in the specific case of 1xx status codes, a direct call to the wrapped ResponseWriter, without marking headers as sent,
// allowing so further calls.
func (r *CodeModifier) WriteHeader(code int) {
	if r.headerSent {
		return
	}

	// Handling informational headers.
	if code >= 100 && code <= 199 {
		// Multiple informational status codes can be used,
		// so here the copy is not appending the values to not repeat them.
		CopyHeaders(r.responseWriter.Header(), r.headerMap)
		r.responseWriter.WriteHeader(code)
		return
	}

	CopyHeaders(r.responseWriter.Header(), r.headerMap)
	r.responseWriter.WriteHeader(r.code)
	r.headerSent = true
}

// Hijack hijacks the connection.
func (r *CodeModifier) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.responseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.responseWriter)
	}
	return hijacker.Hijack()
}

// Flush sends any buffered data to the client.
func (r *CodeModifier) Flush() {
	r.WriteHeader(r.code)

	if flusher, ok := r.responseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
