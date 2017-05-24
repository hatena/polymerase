package api

import "github.com/taku-k/polymerase/pkg/base"

type Config struct {
	*base.Config

	HTTPApiPrefix string
}
