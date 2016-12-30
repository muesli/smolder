package main

import (
	"github.com/emicklei/go-restful"
	"github.com/muesli/smolder"
)

// HelloResource is the resource responsible for /hello
type HelloResource struct {
	smolder.Resource
}

var (
	_ smolder.GetSupported = &HelloResource{}
)

// Register this resource with the container to setup all the routes
func (r *HelloResource) Register(container *restful.Container, config smolder.APIConfig, context smolder.APIContextFactory) {
	r.Name = "HelloResource"
	r.TypeName = "reply"
	r.Endpoint = "hello"
	r.Doc = "Our very own API endpoint"

	r.Config = config
	r.Context = context

	r.Init(container, r)
}
