// Package api contains bounded protocol decoders and protocol-neutral errors.
package api

import "fmt"

type Error struct {
	Status  int
	Type    string
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Type, e.Message) }
func BadRequest(message string) *Error {
	return &Error{Status: 400, Type: "invalid_request_error", Message: message}
}
func NotFound(message string) *Error {
	return &Error{Status: 404, Type: "model_not_found", Message: message}
}
