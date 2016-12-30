package main

import (
	"errors"

	restful "github.com/emicklei/go-restful"
	"github.com/muesli/smolder"
)

// Context is the central API context
type Context struct {
}

// NewAPIContext returns a new context
func (context *Context) NewAPIContext() smolder.APIContext {
	return &Context{}
}

// LogSummary logs out the current context stats
func (context *Context) LogSummary() {
}

// Authentication parses the request for an access-/authtoken and returns the matching user
func (context *Context) Authentication(request *restful.Request) (interface{}, error) {
	return nil, errors.New("Auth is not implemented")
}
