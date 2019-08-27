package server

import "fmt"

type OrtooServerConfig struct {
	host string
	port int
}

func (o *OrtooServerConfig) getHostAddress() string {
	return fmt.Sprintf("%s:%d", o.host, o.port)
}

func DefaultConfig() *OrtooServerConfig {
	return &OrtooServerConfig{
		host: "0.0.0.0",
		port: 19061,
	}
}
