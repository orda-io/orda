package server

import "fmt"

type OrtooConfig struct {
	host string
	port int
}

func (o *OrtooConfig) getHostAddress() string {
	return fmt.Sprintf("%s:%d", o.host, o.port)
}

func DefaultConfig() *OrtooConfig {
	return &OrtooConfig{
		host: "0.0.0.0",
		port: 19061,
	}
}
