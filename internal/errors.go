package internal

import (
	"log"
	"net/http"
)

const (
	ErrMsgHTTPRequestFailed = "Error making HTTP GET request to the API: %v"
	ErrMsgInvalidStatusCode = "API request error, Status code: %d"
	ErrMsgJSONDecodeFailed  = "Error decoding JSON response: %v"
	ErrMsgFetch             = "Error fetching data from the API: %v"
)

type httpError struct { // Extra error handling struct
	StatusCode int
	Message    string
}

func (e *httpError) Error() string { // Custom error message
	return e.Message
}

func NewHTTPError(statusCode int, message string) *httpError { // This generates the error object for our purposes
	return &httpError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func err500(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		log.Fatalln("ERROR:", err)
		http.Redirect(w, r, "/oops", http.StatusFound)
		spooky("The server failed to fulfill an apparently valid request", err) // triggered some sort of internal error
		return
	}
}

func spooky(args ...interface{}) {
	Logger.SetPrefix("ERROR ")
	Logger.Println(args...)
}
