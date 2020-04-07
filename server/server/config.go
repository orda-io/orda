package server

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server/mongodb"
	"io/ioutil"
)

// OrtooServerConfig is a configuration of OrtooServer
type OrtooServerConfig struct {
	OrtooServer  string          `json:"OrtooServer"`
	Notification string          `json:"Notification"`
	Mongo        *mongodb.Config `json:"Mongo"`
}

func LoadOrtooServerConfig(filePath string) (*OrtooServerConfig, error) {
	conf := &OrtooServerConfig{}
	if err := conf.loadConfig(filePath); err != nil {
		return nil, err
	}
	return conf, nil
}

func (its *OrtooServerConfig) loadConfig(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return log.OrtooErrorf(err, "fail to read server config file: %s", filepath)
	}
	if err := json.Unmarshal(data, its); err != nil {
		return log.OrtooErrorf(err, "fail to unmarshal server config file")
	}
	return nil
}
