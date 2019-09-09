package server

import (
	"fmt"
	"github.com/knowhunger/ortoo/server/mongodb"
)

type OrtooServerConfig struct {
	Host  string
	Port  int
	Mongo *mongodb.Config
}

func (o *OrtooServerConfig) getHostAddress() string {
	return fmt.Sprintf("%s:%d", o.Host, o.Port)
}
