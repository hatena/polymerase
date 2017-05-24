package api

import "github.com/taku-k/xtralab/pkg/base"

type Config struct {
	*base.Config

	HTTPApiPrefix string
}
