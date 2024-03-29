package managers

import (
	"encoding/json"
	"fmt"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/server/redis"
	"io/ioutil"

	"github.com/orda-io/orda/server/mongodb"
)

// OrdaServerConfig is a configuration of OrdaServer
type OrdaServerConfig struct {
	RPCServerPort   int             `json:"RPCServerPort"`
	RestfulPort     int             `json:"RestfulPort"`
	SwaggerBasePath string          `json:"SwaggerBasePath"`
	SwaggerJSON     string          `json:"SwaggerJSON"`
	Notification    string          `json:"Notification"`
	Mongo           *mongodb.Config `json:"Mongo"`
	Redis           *redis.Config   `json:"Redis,omitempty"`
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
		return errors.ServerInit.New(log.Logger, fmt.Sprintf("cannot read config file: %v", err.Error()))
	}
	if err := json.Unmarshal(data, its); err != nil {
		return errors.ServerInit.New(log.Logger, "cannot unmarshal server config file:"+err.Error())
	}
	return nil
}

// GetRPCServerAddr returns RPC Server Address
func (its *OrdaServerConfig) GetRPCServerAddr() string {
	return fmt.Sprintf(":%d", its.RPCServerPort)
}

// GetRestfulAddr returns Restful Server Address
func (its *OrdaServerConfig) GetRestfulAddr() string {
	return fmt.Sprintf(":%d", its.RestfulPort)
}

// String returns a marshaled string
func (its *OrdaServerConfig) String() string {
	b, _ := json.Marshal(its)
	return string(b)
}
