/*
 * smolder
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

import (
	"errors"
	"strings"

	"github.com/emicklei/go-restful"
)

func decodeParam(param string) string {
	return strings.Replace(param, "+", " ", -1)
}

// Validate is used to check input for required values
func Validate(request *restful.Request, params []*restful.Parameter) (map[string][]string, error) {
	res := make(map[string][]string)

	for _, p := range params {
		var t string
		values := []string{}

		switch p.Kind() {
		case restful.QueryParameterKind:
			t = "Query"
			if ql, ok := request.Request.URL.Query()[p.Data().Name]; ok {
				for _, q := range ql {
					values = append(values, q)
				}
			}

		case restful.PathParameterKind:
			t = "Path"
			values = append(values, request.PathParameter(p.Data().Name))
		}

		if p.Data().Required && len(values) == 0 {
			return res, errors.New(t + "-Parameter '" + p.Data().Name + "' is required but missing")
		}

		for _, v := range values {
			res[p.Data().Name] = append(res[p.Data().Name], decodeParam(v))
		}
	}

	return res, nil
}
