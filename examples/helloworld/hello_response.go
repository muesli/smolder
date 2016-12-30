package main

import "github.com/muesli/smolder"

// HelloResponse is the common response to 'user' requests
type HelloResponse struct {
	smolder.Response

	Reply string `json:"reply"`
}

// Init a new response
func (r *HelloResponse) Init(context smolder.APIContext) {
	r.Parent = r
	r.Context = context
}
