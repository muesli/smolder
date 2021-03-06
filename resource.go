/*
 * smolder
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import (
	"net/http"
	"reflect"

	"github.com/emicklei/go-restful"
	log "github.com/sirupsen/logrus"
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
	GetByIDsAuthRequired() bool
	Returns() interface{}
}

// GetSupported is the interface Resources need to fulfill to respond to generic GET requests
type GetSupported interface {
	Get(context APIContext, request *restful.Request, response *restful.Response, params map[string][]string)
	GetAuthRequired() bool
	GetDoc() string
	GetParams() []*restful.Parameter
	Returns() interface{}
}

// PostSupported is the interface Resources need to fulfill to respond to generic POST requests
type PostSupported interface {
	Post(context APIContext, data interface{}, request *restful.Request, response *restful.Response)
	PostAuthRequired() bool
	PostDoc() string
	PostParams() []*restful.Parameter
	Validate(context APIContext, data interface{}, request *restful.Request) error
	Reads() interface{}
	Returns() interface{}
}

// PutSupported is the interface Resources need to fulfill to respond to generic PUT requests
type PutSupported interface {
	Put(context APIContext, data interface{}, request *restful.Request, response *restful.Response)
	PutAuthRequired() bool
	PutDoc() string
	PutParams() []*restful.Parameter
	Validate(context APIContext, data interface{}, request *restful.Request) error
	Reads() interface{}
	Returns() interface{}
}

// PatchSupported is the interface Resources need to fulfill to respond to generic PATCH requests
type PatchSupported interface {
	Patch(context APIContext, data interface{}, request *restful.Request, response *restful.Response)
	PatchAuthRequired() bool
	PatchDoc() string
	PatchParams() []*restful.Parameter
	Validate(context APIContext, data interface{}, request *restful.Request) error
	Reads() interface{}
	Returns() interface{}
}

// DeleteSupported is the interface Resources need to fulfill to respond to generic DELETE requests
type DeleteSupported interface {
	Delete(context APIContext, request *restful.Request, response *restful.Response)
	DeleteAuthRequired() bool
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
	if resource, ok := resource.(GetIDSupported); ok {
		isDatabaseItem = true
		route := ws.GET("/{id:*}").To(r.GetByIDs).
			Doc("get item by id").
			Returns(http.StatusOK, "OK", resource.Returns()).
			Returns(http.StatusNotFound, "Not found", ErrorResponse{}).
			Param(ws.PathParameter("id", "ID of "+r.TypeName).
				DataType("string").
				Required(true).
				AllowMultiple(false))

		if resource.GetByIDsAuthRequired() {
			route.Returns(http.StatusUnauthorized, "Authorization required", ErrorResponse{})
			route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
				DataType("string").
				Required(true).
				AllowMultiple(false))
		}

		ws.Route(route)
	}

	isGetSupported := false
	if resource, ok := resource.(GetSupported); ok {
		isGetSupported = true
		route := ws.GET("").To(r.Get).
			Doc(resource.GetDoc()).
			Returns(http.StatusOK, "OK", resource.Returns()).
			Returns(http.StatusNotFound, "Not found", ErrorResponse{})

		if resource.GetAuthRequired() {
			route.Returns(http.StatusUnauthorized, "Authorization required", ErrorResponse{})
			route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
				DataType("string").
				Required(true).
				AllowMultiple(false))
		}

		for _, p := range resource.GetParams() {
			route.Param(p)
		}
		if isDatabaseItem {
			route.Param(ws.QueryParameter("ids[]", "IDs of "+r.TypeName+"s").
				DataType("string").
				// Required(true).
				AllowMultiple(true))
		}
		ws.Route(route)
	}

	if isDatabaseItem && !isGetSupported {
		route := ws.GET("").To(r.GetByIDs).
			Doc("get "+r.TypeName+" by ids").
			Returns(http.StatusNotFound, "Not found", ErrorResponse{}).
			Param(ws.QueryParameter("ids[]", "IDs of "+r.TypeName+"s").
				DataType("string").
				// Required(true).
				AllowMultiple(true))

		ws.Route(route)
	}

	if resource, ok := resource.(PostSupported); ok {
		route := ws.POST("").To(r.Post).
			Doc(resource.PostDoc()).
			Reads(reflect.Indirect(reflect.ValueOf(resource.Reads())).Interface()).
			Returns(http.StatusOK, "OK", resource.Returns()).
			Returns(http.StatusBadRequest, "Invalid post data", ErrorResponse{})

		if resource.PostAuthRequired() {
			route.Returns(http.StatusUnauthorized, "Authorization required", ErrorResponse{})
			route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
				DataType("string").
				Required(true).
				AllowMultiple(false))
		}

		for _, p := range resource.PostParams() {
			route.Param(p)
		}

		ws.Route(route)
	}

	if resource, ok := resource.(PutSupported); ok {
		route := ws.PUT("/{"+r.TypeName+"-id:*}").To(r.Put).
			Doc(resource.PutDoc()).
			Reads(reflect.Indirect(reflect.ValueOf(resource.Reads())).Interface()).
			Returns(http.StatusOK, "OK", resource.Returns()).
			Returns(http.StatusNotFound, "Not found", ErrorResponse{}).
			Returns(http.StatusBadRequest, "Invalid put data", ErrorResponse{})

		if resource.PutAuthRequired() {
			route.Returns(http.StatusUnauthorized, "Authorization required", ErrorResponse{})
			route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
				DataType("string").
				Required(true).
				AllowMultiple(false))
		}

		for _, p := range resource.PutParams() {
			route.Param(p)
		}

		route.Param(restful.PathParameter(r.TypeName+"-id", "ID of a "+r.TypeName).
			Required(true).
			AllowMultiple(false))

		ws.Route(route)
	}

	if resource, ok := resource.(PatchSupported); ok {
		route := ws.PATCH("/{"+r.TypeName+"-id:*}").To(r.Patch).
			Doc(resource.PatchDoc()).
			Reads(reflect.Indirect(reflect.ValueOf(resource.Reads())).Interface()).
			Returns(http.StatusOK, "OK", resource.Returns()).
			Returns(http.StatusNotFound, "Not found", ErrorResponse{}).
			Returns(http.StatusBadRequest, "Invalid patch data", ErrorResponse{})

		if resource.PatchAuthRequired() {
			route.Returns(http.StatusUnauthorized, "Authorization required", ErrorResponse{})
			route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
				DataType("string").
				Required(true).
				AllowMultiple(false))
		}

		for _, p := range resource.PatchParams() {
			route.Param(p)
		}

		route.Param(restful.PathParameter(r.TypeName+"-id", "ID of a "+r.TypeName).
			Required(true).
			AllowMultiple(false))

		ws.Route(route)
	}

	if resource, ok := resource.(DeleteSupported); ok {
		route := ws.DELETE("/{"+r.TypeName+"-id:*}").To(r.Delete).
			Doc(resource.DeleteDoc()).
			Returns(http.StatusNotFound, "Not found", ErrorResponse{})

		if resource.DeleteAuthRequired() {
			route.Returns(http.StatusUnauthorized, "Authorization required", ErrorResponse{})
			route.Param(restful.QueryParameter("accesstoken", "accesstoken required for auth").
				DataType("string").
				Required(true).
				AllowMultiple(false))
		}

		for _, p := range resource.DeleteParams() {
			route.Param(p)
		}

		route.Param(restful.PathParameter(r.TypeName+"-id", "ID of a "+r.TypeName).
			Required(true).
			AllowMultiple(false))

		ws.Route(route)
	}

	container.Add(ws)
}

// Get responds to GET requests
func (r Resource) Get(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(GetSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if resource.GetAuthRequired() {
			if err != nil || auth == nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusUnauthorized,
					"Invalid accesstoken",
					"GET"))
				return
			}
		}
		context.SetAuth(auth)

		params, err := Validate(request, resource.GetParams())
		if err != nil {
			ErrorResponseHandler(request, response, err, NewErrorResponse(
				http.StatusBadRequest,
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

		resource.Get(context, request, response, params)
		request.SetAttribute("context", context)
	}
}

// Post responds to POST requests
func (r Resource) Post(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(PostSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if resource.PostAuthRequired() {
			if err != nil || auth == nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusUnauthorized,
					"Invalid accesstoken",
					"POST"))
				return
			}
		}
		context.SetAuth(auth)

		ps := resource.Reads()
		if ps != nil {
			err := request.ReadEntity(&ps)
			if err != nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusBadRequest,
					"Can't parse request data",
					"POST Data Validation"))
				return
			}

			err = resource.Validate(context, ps, request)
			if err != nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusBadRequest,
					err,
					"POST Data Validation"))
				return
			}
		}

		resource.Post(context, ps, request, response)
		request.SetAttribute("context", context)
	}
}

// Put responds to PUT requests
func (r Resource) Put(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(PutSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if resource.PutAuthRequired() {
			if err != nil || auth == nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusUnauthorized,
					"Invalid accesstoken",
					"PUT"))
				return
			}
		}
		context.SetAuth(auth)

		ps := resource.Reads()
		if ps != nil {
			err := request.ReadEntity(&ps)
			if err != nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusBadRequest,
					"Can't parse request data",
					"PUT Data Validation"))
				return
			}

			err = resource.Validate(context, ps, request)
			if err != nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusBadRequest,
					err,
					"PUT Data Validation"))
				return
			}
		}

		resource.Put(context, ps, request, response)
		request.SetAttribute("context", context)
	}
}

// Patch responds to PATCH requests
func (r Resource) Patch(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(PatchSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if resource.PatchAuthRequired() {
			if err != nil || auth == nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusUnauthorized,
					"Invalid accesstoken",
					"PATCH"))
				return
			}
		}
		context.SetAuth(auth)

		ps := resource.Reads()
		if ps != nil {
			err := request.ReadEntity(&ps)
			if err != nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusBadRequest,
					"Can't parse request data",
					"PATCH Data Validation"))
				return
			}

			err = resource.Validate(context, ps, request)
			if err != nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusBadRequest,
					err,
					"PATCH Data Validation"))
				return
			}
		}

		resource.Patch(context, ps, request, response)
		request.SetAttribute("context", context)
	}
}

// Delete responds to DELETE requests
func (r Resource) Delete(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(DeleteSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if resource.DeleteAuthRequired() {
			if err != nil || auth == nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusUnauthorized,
					"Invalid accesstoken",
					"DELETE"))
				return
			}
		}
		context.SetAuth(auth)

		resource.Delete(context, request, response)
		request.SetAttribute("context", context)
	}
}

// GetByIDs handles GET requests which want to retrieve one or more IDs
func (r Resource) GetByIDs(request *restful.Request, response *restful.Response) {
	if resource, ok := r.Parent.(GetIDSupported); ok {
		context := r.Context.NewAPIContext()
		auth, err := context.Authentication(request)
		if resource.GetByIDsAuthRequired() {
			if err != nil || auth == nil {
				ErrorResponseHandler(request, response, err, NewErrorResponse(
					http.StatusUnauthorized,
					"Invalid accesstoken",
					"GET"))
				return
			}
		}
		context.SetAuth(auth)

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
			ErrorResponseHandler(request, response, nil, NewErrorResponse(
				http.StatusBadRequest,
				"No item-id(s) specified",
				"validate"))
			return
		}

		resource.GetByIDs(context, request, response, ids)
		request.SetAttribute("context", context)
	}
}

// NotFound is the default 404 response
func (r Resource) NotFound(request *restful.Request, response *restful.Response) {
	ErrorResponseHandler(request, response, nil, NewErrorResponse(
		http.StatusNotFound,
		"This "+r.TypeName+" does not exist.",
		r.TypeName))
}
