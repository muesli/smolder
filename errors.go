/*
 * smolder
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import (
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
)

// APIError describes an API error
type APIError struct {
	Code          int    `json:"statusCode"`
	InternalError bool   `json:"internalerror,omitempty"`
	Msg           string `json:"description"`
	Context       string `json:"context,omitempty"`
}

// ErrorResponse is the default error handling response
type ErrorResponse struct {
	Err APIError `json:"error"`
}

func (err *ErrorResponse) Error() string {
	return err.Err.Msg
}

// NewErrorResponse creates a new ErrorResponse with the provided values
func NewErrorResponse(code int, internal bool, err interface{}, context string) *ErrorResponse {
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
		APIError{
			Code:          code,
			InternalError: internal,
			Msg:           msg,
			Context:       context,
		},
	}
}

// ErrorResponseHandler is the default error response handler
func ErrorResponseHandler(request *restful.Request, response *restful.Response, err *ErrorResponse) {
	fields := log.Fields{
		"Internal":     err.Err.InternalError,
		"Description":  err.Err.Msg,
		"Context":      err.Err.Context,
		"URL":          request.Request.URL.String(),
		"Method":       request.Request.Method,
		"http_request": request.Request,
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
	log.WithFields(fields).Error(err)

	if response != nil {
		response.WriteHeaderAndEntity(err.Err.Code, err)
	}
}
