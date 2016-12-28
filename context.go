/*
 * smolder
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import "github.com/emicklei/go-restful"

// APIContextFactory allows you to retrieve a new APIContext
type APIContextFactory interface {
	NewAPIContext() APIContext
}

// APIContext contains all the functions required to interact with the API user
type APIContext interface {
	Authentication(request *restful.Request) (interface{}, error)
}
