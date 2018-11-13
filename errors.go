/*
 * smolder
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import (
	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
)

// APIError describes an API error
type APIError struct {
	Code          int    `json:"statusCode"`
	InternalError bool   `json:"internalerror,omitempty"`
	Msg           string `json:"detail"`
	Source        struct {
		Pointer string `json:"pointer"`
	} `json:"source"`
	Context string `json:"context,omitempty"`
}

// ErrorResponse is the default error handling response
type ErrorResponse struct {
	Err []APIError `json:"errors"`
}

func (err *ErrorResponse) Error() string {
	return err.Err[0].Msg
}

// NewErrorResponse creates a new ErrorResponse with the provided values
func NewErrorResponse(code int, err interface{}, context string) *ErrorResponse {
	var msg string

	switch err := err.(type) {
	case error:
		msg = err.Error()
	case string:
		msg = err
	default:
		return nil
	}

	return &ErrorResponse{
		Err: []APIError{
			{
				Code:          code,
				InternalError: code == 500,
				Msg:           msg,
				Context:       context,
			},
		},
	}
}

// ErrorResponseHandler is the default error response handler
func ErrorResponseHandler(request *restful.Request, response *restful.Response, origin error, err *ErrorResponse) {
	fields := log.Fields{
		"Internal":    err.Err[0].InternalError,
		"Description": err.Err[0].Msg,
		"Context":     err.Err[0].Context,
		"URL":         request.Request.URL.String(),
		"Method":      request.Request.Method,
	}

	for k, vs := range request.Request.Form {
		var out string
		for i, v := range vs {
			if i > 0 {
				out += ","
			}
			out += v
		}
		fields[k] = out
	}
	log.WithFields(fields).Error(origin)

	if response != nil {
		response.WriteHeaderAndEntity(err.Err[0].Code, err)
	}
}
