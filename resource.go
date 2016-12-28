/*
 * smolder
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
)

// APIResource contains all the functions required to register a new API resource
type APIResource interface {
	Register(container *restful.Container, config APIConfig, context APIContextFactory)
}

// Resource describes an API resource/endpoint
type Resource struct {
	Name     string
	TypeName string
	Endpoint string
	Doc      string

	Config  APIConfig
	Context APIContextFactory

	Parent interface{}
}

// GetIDSupported is the interface Resources need to fulfill to respond to GET-by-ID requests
type GetIDSupported interface {
	GetByIDs(context APIContext, request *restful.Request, response *restful.Response, ids []string)
}

// GetSupported is the interface Resources need to fulfill to respond to generic GET requests
type GetSupported interface {
	Get(context APIContext, request *restful.Request, response *restful.Response, params map[string][]string)
	GetDoc() string
	GetParams() []*restful.Parameter
}

// PostSupported is the interface Resources need to fulfill to respond to generic POST requests
type PostSupported interface {
	Post(context APIContext, request *restful.Request, response *restful.Response, auth interface{})
	PostDoc() string
	PostParams() []*restful.Parameter
}

// PutSupported is the interface Resources need to fulfill to respond to generic PUT requests
type PutSupported interface {
	Put(context APIContext, request *restful.Request, response *restful.Response, auth interface{})
	PutDoc() string
	PutParams() []*restful.Parameter
}

// PatchSupported is the interface Resources need to fulfill to respond to generic PATCH requests
type PatchSupported interface {
	Patch(context APIContext, request *restful.Request, response *restful.Response, auth interface{})
	PatchDoc() string
	PatchParams() []*restful.Parameter
}

// DeleteSupported is the interface Resources need to fulfill to respond to generic DELETE requests
type DeleteSupported interface {
	Delete(context APIContext, request *restful.Request, response *restful.Response, auth interface{})
	DeleteDoc() string
	DeleteParams() []*restful.Parameter
}

// Init registers a resource with the Container and sets up all the supported routes
func (r Resource) Init(container *restful.Container, resource interface{}) {
	log.WithField("Resource", r.Name).Info("Registering Resource")
	ws := new(restful.WebService)
	r.Parent = resource

	ws.Path("/" + r.Config.PathPrefix + r.Endpoint).
		Doc(r.Doc).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	isDatabaseItem := false
	if _, ok := resource.(GetIDSupported); ok {
		isDatabaseItem = true
		route := ws.GET("/{id}").To(r.GetByIDs).
			Doc("get item by id").
			Param(ws.PathParameter("id", "ID of "+r.TypeName).
				DataType("string").
				Required(true).
				AllowMultiple(false))

		ws.Route(route)
	}

	isGetSupported := false
	if resource, ok := resource.(GetSupported); ok {
		isGetSupported = true
		route := ws.GET("").To(r.Get).
			Doc(resource.GetDoc())

		for _, p := range resource.GetParams() {
			route.Param(p)
		}
		if isDatabaseItem {
			route.Param(ws.QueryParameter("ids[]", "IDs of "+r.TypeName+"s").
				DataType("string").
				Required(true).
				AllowMultiple(true))
		}
		ws.Route(route)
	}

	if isDatabaseItem && !isGetSupported {
		route := ws.GET("").To(r.GetByIDs).
			Doc("get " + r.TypeName + " by ids").
			Param(ws.QueryParameter("ids[]", "IDs of "+r.TypeName+"s").
				DataType("string").
				Required(true).
				AllowMultiple(true))

		ws.Route(route)
	}

	if resource, ok := resource.(PostSupported); ok {
		route := ws.POST("").Filter(r.authenticate).To(r.Post).
			Doc(resource.PostDoc())

		for _, p := range resource.PostParams() {
			route.Param(p)
		}
		route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
			DataType("string").
			Required(true).
			AllowMultiple(false))

		ws.Route(route)
	}

	if resource, ok := resource.(PutSupported); ok {
		route := ws.PUT("/{" + r.TypeName + "-id}").Filter(r.authenticate).To(r.Put).
			Doc(resource.PutDoc())

		for _, p := range resource.PutParams() {
			route.Param(p)
		}

		route.Param(restful.PathParameter(r.TypeName+"-id", "ID of a "+r.TypeName).
			Required(true).
			AllowMultiple(false))

		route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
			DataType("string").
			Required(true).
			AllowMultiple(false))

		ws.Route(route)
	}

	if resource, ok := resource.(PatchSupported); ok {
		route := ws.PATCH("/{" + r.TypeName + "-id").Filter(r.authenticate).To(r.Patch).
			Doc(resource.PatchDoc())

		for _, p := range resource.PatchParams() {
			route.Param(p)
		}

		route.Param(restful.PathParameter(r.TypeName+"-id", "ID of a "+r.TypeName).
			Required(true).
			AllowMultiple(false))

		route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
			DataType("string").
			Required(true).
			AllowMultiple(false))

		ws.Route(route)
	}

	if resource, ok := resource.(DeleteSupported); ok {
		route := ws.DELETE("/{" + r.TypeName + "-id}").Filter(r.authenticate).To(r.Delete).
			Doc(resource.DeleteDoc())

		for _, p := range resource.DeleteParams() {
			route.Param(p)
		}

		route.Param(restful.PathParameter(r.TypeName+"-id", "ID of a "+r.TypeName).
			Required(true).
			AllowMultiple(false))

		route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
			DataType("string").
			Required(true).
			AllowMultiple(false))

		ws.Route(route)
	}

	container.Add(ws)
}

// Get responds to GET requests
func (r Resource) Get(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(GetSupported); ok {
		params, err := Validate(request, resource.GetParams())
		if err != nil {
			ErrorResponseHandler(request, response, NewErrorResponse(
				http.StatusBadRequest,
				false,
				err,
				"validate"))
			return
		}

		if _, ok := r.Parent.(GetIDSupported); ok {
			//if _, ok := params["ids[]"]; ok { //FIXME
			if _, ok := request.Request.URL.Query()["ids[]"]; ok {
				r.GetByIDs(request, response)
				return
			}
		}

		context := r.Context.NewAPIContext()
		resource.Get(context, request, response, params)
		request.SetAttribute("context", context)
	}
}

// Post responds to POST requests
func (r Resource) Post(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(PostSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if err != nil {
			ErrorResponseHandler(request, response, NewErrorResponse(
				http.StatusUnauthorized,
				false,
				"Invalid accesstoken",
				"POST"))
			return
		}

		resource.Post(context, request, response, auth)
		request.SetAttribute("context", context)
	}
}

// Put responds to PUT requests
func (r Resource) Put(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(PutSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if err != nil {
			ErrorResponseHandler(request, response, NewErrorResponse(
				http.StatusUnauthorized,
				false,
				"Invalid accesstoken",
				"PUT"))
			return
		}

		resource.Put(context, request, response, auth)
		request.SetAttribute("context", context)
	}
}

// Patch responds to PATCH requests
func (r Resource) Patch(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(PatchSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if err != nil {
			ErrorResponseHandler(request, response, NewErrorResponse(
				http.StatusUnauthorized,
				false,
				"Invalid accesstoken",
				"PATCH"))
			return
		}

		resource.Patch(context, request, response, auth)
		request.SetAttribute("context", context)
	}
}

// Delete responds to DELETE requests
func (r Resource) Delete(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(DeleteSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if err != nil {
			ErrorResponseHandler(request, response, NewErrorResponse(
				http.StatusUnauthorized,
				false,
				"Invalid accesstoken",
				"DELETE"))
			return
		}

		resource.Delete(context, request, response, auth)
		request.SetAttribute("context", context)
	}
}

// GetByIDs handles GET requests which want to retrieve one or more IDs
func (r Resource) GetByIDs(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(GetIDSupported); ok {
		ids := []string{}
		if ql, ok := request.Request.URL.Query()["ids[]"]; ok {
			for _, q := range ql {
				if len(q) > 0 {
					ids = append(ids, q)
				}
			}
		}
		pathID := request.PathParameter("id")
		if len(pathID) > 0 {
			ids = append(ids, pathID)
		}

		if len(ids) == 0 {
			ErrorResponseHandler(request, response, NewErrorResponse(
				http.StatusBadRequest,
				false,
				"No item-id(s) specified",
				"validate"))
			return
		}

		context := r.Context.NewAPIContext()
		resource.GetByIDs(context, request, response, ids)
		request.SetAttribute("context", context)
	}
}

// NotFound is the default 404 response
func (r Resource) NotFound(request *restful.Request, response *restful.Response) {
	ErrorResponseHandler(request, response, NewErrorResponse(
		http.StatusNotFound,
		false,
		"This "+r.TypeName+" does not exist.",
		r.TypeName))
}

func (r Resource) authenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	context := r.Context.NewAPIContext()
	_, err := context.Authentication(request)
	if err != nil {
		ErrorResponseHandler(request, response, NewErrorResponse(
			http.StatusUnauthorized,
			false,
			err,
			"authenticate"))
		return
	}

	chain.ProcessFilter(request, response)
}
