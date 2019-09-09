package testonly

import "github.com/knowhunger/ortoo/client"

func TestOrtooClientConfig1() *client.OrtooClientConfig {
	return &client.OrtooClientConfig{
		Address:        "127.0.0.1",
		Port:           19061,
		CollectionName: "OrtooTest",
	}
}
