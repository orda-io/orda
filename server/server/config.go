package server

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/server/mongodb"
	"io/ioutil"
)

// OrtooServerConfig is a configuration of OrtooServer
type OrtooServerConfig struct {
	RPCServerPort int            `json:"RPCServerPort"`
	RestfulPort   int            `json:"RestfulPort"`
	Notification  string         `json:"Notification"`
	Mongo         mongodb.Config `json:"Mongo"`
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

func (its *OrtooServerConfig) getRPCServerAddr() string {
	return fmt.Sprintf(":%d", its.RPCServerPort)
}

func (its *OrtooServerConfig) getRestfulAddr() string {
	return fmt.Sprintf(":%d", its.RestfulPort)
}
