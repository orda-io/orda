package client

import "fmt"

type OrtooClientConfig struct {
	Address        string
	Port           int
	CollectionName string
	Alias          string
}

func (o *OrtooClientConfig) getServiceHost() string {
	return fmt.Sprintf("%s:%d", o.Address, o.Port)
}
