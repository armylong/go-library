package longgin

import (
	"context"
)

// ErrorUnknown
const (
	ErrorOK      = ErrorCode(0)
	ErrorUnknown = ErrorCode(-1)
)

type ErrorCode int

type ErrorHandler func(ctx context.Context, err error, code ErrorCode, data ...any) *Response

type Response struct {
	ErrorCode    ErrorCode `json:"errorCode"`
	ErrorMsg     string    `json:"errorMsg"`
	ResponseData any       `json:"responseData,omitempty"`
}

// ErrorWithContext
func ErrorWithContext(ctx context.Context, err error, code ErrorCode, data ...any) *Response {
	msg := err.Error()
	resp := &Response{
		ErrorCode:    code,
		ErrorMsg:     msg,
		ResponseData: nil,
	}
	return resp
}

// RawResponse
type RawResponse interface {
	_raw()
}

// Success
func Success(data any, msg ...string) *Response {
	resp := &Response{
		ErrorCode:    ErrorOK,
		ErrorMsg:     "ok",
		ResponseData: data,
	}
	if len(msg) > 0 {
		resp.ErrorMsg = msg[0]
	}
	return resp
}
