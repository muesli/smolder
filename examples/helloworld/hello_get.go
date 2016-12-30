package main

import (
	"github.com/emicklei/go-restful"
	"github.com/muesli/smolder"
)

// GetAuthRequired returns true because all requests need authentication
func (r *HelloResource) GetAuthRequired() bool {
	return false
}

// GetDoc returns the description of this API endpoint
func (r *HelloResource) GetDoc() string {
	return "request a hello"
}

// GetParams returns the parameters supported by this API endpoint
func (r *HelloResource) GetParams() []*restful.Parameter {
	params := []*restful.Parameter{}
	params = append(params, restful.QueryParameter("name", "your name").
		DataType("string").
		AllowMultiple(false).
		Required(true))

	return params
}

// Get sends out items matching the query parameters
func (r *HelloResource) Get(context smolder.APIContext, request *restful.Request, response *restful.Response, params map[string][]string) {
	resp := HelloResponse{}
	resp.Init(context)

	name := params["name"]
	if len(name) > 0 {
		resp.Reply = "Hello " + name[0]
	}

	resp.Send(response)
}
