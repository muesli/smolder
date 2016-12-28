/*
 * smolder
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import (
	"net/http"

	"github.com/emicklei/go-restful"
)

// Response provides convenience methods to respond to requests
type Response struct {
	Context APIContext  `json:"-"`
	Parent  interface{} `json:"-"`
}

// SupportsEmptyResponse is an interface to provide empty-result responses
type SupportsEmptyResponse interface {
	EmptyResponse() interface{}
}

// Send responds to a request with http.StatusOK
func (rc *Response) Send(response *restful.Response) {
	rc.SendWithHeader(http.StatusOK, response)
}

// SendWithHeader responds to a request with a custom HTTP status code
func (rc *Response) SendWithHeader(status int, response *restful.Response) {
	if rc.Parent != nil {
		if resp, ok := rc.Parent.(SupportsEmptyResponse); ok {
			out := resp.EmptyResponse()
			if out != nil {
				response.WriteHeaderAndEntity(status, out)
				return
			}
		}

		response.WriteHeaderAndEntity(status, rc.Parent)
		return
	}

	response.WriteHeaderAndEntity(status, rc)
}
