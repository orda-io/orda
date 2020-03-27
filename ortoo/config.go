package ortoo

import "github.com/knowhunger/ortoo/ortoo/model"

// ClientConfig is a configuration for OrtooClient
type ClientConfig struct {
	ServerAddr       string
	NotificationAddr string
	CollectionName   string
	SyncType         model.SyncType
}

// NewLocalClientConfig makes a new local client which do not synchronize with OrtooServer
func NewLocalClientConfig(collectionName string) *ClientConfig {
	return &ClientConfig{
		ServerAddr:       "",
		CollectionName:   collectionName,
		NotificationAddr: "",
		SyncType:         0,
	}
}
