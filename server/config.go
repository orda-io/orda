package server

import (
	"github.com/knowhunger/ortoo/server/mongodb"
)

// OrtooServerConfig is a configuration of OrtooServer
type OrtooServerConfig struct {
	Host       string
	Mongo      *mongodb.Config
	PubSubAddr string
}
