package server

import (
	"encoding/json"
	"fmt"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/server/mongodb"
	"io/ioutil"
)

// OrdaServerConfig is a configuration of OrdaServer
type OrdaServerConfig struct {
	RPCServerPort int            `json:"RPCServerPort"`
	RestfulPort   int            `json:"RestfulPort"`
	Notification  string         `json:"Notification"`
	Mongo         mongodb.Config `json:"Mongo"`
}

// LoadOrdaServerConfig loads config from file.
func LoadOrdaServerConfig(filePath string) (*OrdaServerConfig, errors.OrdaError) {
	conf := &OrdaServerConfig{}
	if err := conf.loadConfig(filePath); err != nil {
		return nil, err
	}
	return conf, nil
}

func (its *OrdaServerConfig) loadConfig(filepath string) errors.OrdaError {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.ServerInit.New(log.Logger, "cannot read server config file "+filepath)
	}
	if err := json.Unmarshal(data, its); err != nil {
		return errors.ServerInit.New(log.Logger, "cannot unmarshal server config file:"+err.Error())
	}
	return nil
}

func (its *OrdaServerConfig) GetRPCServerAddr() string {
	return fmt.Sprintf(":%d", its.RPCServerPort)
}

func (its *OrdaServerConfig) GetRestfulAddr() string {
	return fmt.Sprintf(":%d", its.RestfulPort)
}
