/*
 * smolder
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
)

const (
	optionsReqIdentifier = "OPTIONS"
)

var (
	shutdownGracefully *bool
	requestIncChan     chan int
)

func gracefulShutdownFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if *shutdownGracefully {
		var resp struct {
			Error string `json:"error"`
		}
		resp.Error = "Server is shutting down"

		log.Warn("Rejecting incoming request")
		response.WriteHeaderAndEntity(http.StatusServiceUnavailable, resp)
		return
	}

	defer func() {
		// Make sure pendingReqeusts gets decremented even if a panic was
		// thrown in ProcessFilter().
		requestIncChan <- -1
	}()

	requestIncChan <- 1
	chain.ProcessFilter(request, response)
}

func loggingFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	if req.Request.Method != optionsReqIdentifier {
		log.WithFields(log.Fields{
			"Method": req.Request.Method,
			"URL":    req.Request.URL,
		}).Info("Handling request")
	}

	resp.PrettyPrint(false)
	chain.ProcessFilter(req, resp)
	duration := time.Since(start)

	if req.Request.Method != optionsReqIdentifier {
		log.WithFields(log.Fields{
			"Method":   req.Request.Method,
			"URL":      req.Request.URL,
			"Response": resp.StatusCode(),
			"Duration": duration,
		}).Info("Finished request")
	}
}

func corsFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	response.Header().Set("Access-Control-Allow-Origin", "*")
	chain.ProcessFilter(request, response)
}

func optionsFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	if request.Request.Method == optionsReqIdentifier {
		response.AddHeader(restful.HEADER_AccessControlAllowHeaders, "authorization, content-type")
		response.AddHeader(restful.HEADER_AccessControlAllowMethods, "GET, POST, PUT, PATCH, DELETE")
	}

	chain.ProcessFilter(request, response)
}

// NewSmolderContainer initializes a new Container with all the default filters
func NewSmolderContainer(config APIConfig, _shutdownGracefully *bool, _requestIncChan chan int) *restful.Container {
	shutdownGracefully = _shutdownGracefully
	requestIncChan = _requestIncChan

	wsContainer := restful.NewContainer()
	wsContainer.Filter(gracefulShutdownFilter)
	wsContainer.Filter(loggingFilter)
	wsContainer.Filter(optionsFilter)
	wsContainer.Filter(corsFilter)
	wsContainer.Filter(wsContainer.OPTIONSFilter)

	if config.SwaggerFilePath != "" {
		wsConfig := swagger.Config{
			WebServices:     wsContainer.RegisteredWebServices(),
			WebServicesUrl:  config.BaseURL,
			ApiPath:         config.SwaggerAPIPath,
			SwaggerPath:     config.SwaggerPath,
			SwaggerFilePath: config.SwaggerFilePath,
		}
		swagger.RegisterSwaggerService(wsConfig, wsContainer)
	}

	return wsContainer
}

func init() {
	restful.SetLogger(log.StandardLogger())
}
