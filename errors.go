package herdrsock

import (
	"errors"
	"fmt"
)

var ErrEmptyResponse = errors.New("herdrsock: empty api response")

// WireError is the error object Herdr returned on the socket.
type WireError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps a Herdr API error response.
type ErrorResponse struct {
	ID   string
	Body WireError
}

func (e *ErrorResponse) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Body.Code == "" {
		return e.Body.Message
	}
	return fmt.Sprintf("%s: %s", e.Body.Code, e.Body.Message)
}

// IDMismatchError reports a malformed response with the wrong request id.
type IDMismatchError struct {
	Want string
	Got  string
}

func (e *IDMismatchError) Error() string {
	return fmt.Sprintf("herdrsock: response id mismatch: want %q, got %q", e.Want, e.Got)
}

// ProtocolMismatchError reports a running Herdr server with a different
// protocol version.
type ProtocolMismatchError struct {
	Required uint32
	Actual   uint32
	Version  string
}

func (e *ProtocolMismatchError) Error() string {
	return fmt.Sprintf("herdrsock: herdr server protocol mismatch: need %d, got %d (server version %s)", e.Required, e.Actual, e.Version)
}
