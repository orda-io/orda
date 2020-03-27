package server

import (
	"github.com/knowhunger/ortoo/server/mongodb"
)

// OrtooServerConfig is a configuration of OrtooServer
type OrtooServerConfig struct {
	OrtooServer      string
	Mongo            *mongodb.Config
	NotificationAddr string
}
