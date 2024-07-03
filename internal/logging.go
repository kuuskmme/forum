package internal

import (
	"bytes"
	"log"
	"net/http"
)

var Logger *log.Logger

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       bytes.Buffer
}

type HttpError struct {
	Code int
	Msg  string
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	// Default to 200 OK in case WriteHeader is not called
	return &LoggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK, headers: make(http.Header)}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	// Note: Do not call lrw.ResponseWriter.WriteHeader here.
}

func (lrw *LoggingResponseWriter) StatusCode() int {
	return lrw.statusCode
}

func (lrw *LoggingResponseWriter) Header() http.Header {
	return lrw.headers
}

func (lrw *LoggingResponseWriter) Write(data []byte) (int, error) {
	return lrw.body.Write(data) // Write to the buffer instead of the underlying response.
}

func (lrw *LoggingResponseWriter) Flush() {
	// Write the buffered headers
	for name, values := range lrw.headers {
		for _, value := range values {
			lrw.ResponseWriter.Header().Add(name, value)
		}
	}
	// Write the status code
	lrw.ResponseWriter.WriteHeader(lrw.statusCode)
	// Write the buffered body
	lrw.body.WriteTo(lrw.ResponseWriter)
}
