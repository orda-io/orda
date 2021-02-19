package server

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
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

// LoadOrtooServerConfig loads config from file.
func LoadOrtooServerConfig(filePath string) (*OrtooServerConfig, errors.OrtooError) {
	conf := &OrtooServerConfig{}
	if err := conf.loadConfig(filePath); err != nil {
		return nil, err
	}
	return conf, nil
}

func (its *OrtooServerConfig) loadConfig(filepath string) errors.OrtooError {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.ServerInit.New(log.Logger, "cannot read server config file "+filepath)
	}
	if err := json.Unmarshal(data, its); err != nil {
		return errors.ServerInit.New(log.Logger, "cannot unmarshal server config file:"+err.Error())
	}
	return nil
}

func (its *OrtooServerConfig) GetRPCServerAddr() string {
	return fmt.Sprintf(":%d", its.RPCServerPort)
}

func (its *OrtooServerConfig) GetRestfulAddr() string {
	return fmt.Sprintf(":%d", its.RestfulPort)
}
