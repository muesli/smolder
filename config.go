/*
 * smolder
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package smolder

// APIConfig contains all parameters required to set up a new smolder API
type APIConfig struct {
	BaseURL    string
	PathPrefix string
}
