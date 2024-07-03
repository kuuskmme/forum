package internal

import (
	"github.com/microcosm-cc/bluemonday"
)

func sanitizeInput(input string) string {
	// Using bluemonday lib for basic data sanitization
	return bluemonday.UGCPolicy().Sanitize(input)
}
