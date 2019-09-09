package integration

import (
	"github.com/knowhunger/ortoo/client"
	"github.com/knowhunger/ortoo/server"
	"github.com/knowhunger/ortoo/server/mongodb"
)

func NewTestOrtooClientConfig() *client.OrtooClientConfig {
	return &client.OrtooClientConfig{
		Address:        "127.0.0.1",
		Port:           19061,
		CollectionName: "",
	}
}

func NewTestMongoDBConfig() *mongodb.Config {
	return &mongodb.Config{
		Host:    "mongodb://root:ortoo-test@localhost:27017",
		OrtooDB: "ortoo_test",
	}
}

func NewTestOrtooServerConfig() *server.OrtooServerConfig {
	return &server.OrtooServerConfig{
		Host:  "127.0.0.1",
		Port:  19061,
		Mongo: NewTestMongoDBConfig(),
	}
}
