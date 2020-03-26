package ortoo

import "github.com/knowhunger/ortoo/ortoo/model"

// ClientConfig is a configuration for OrtooClient
type ClientConfig struct {
	Address          string
	CollectionName   string
	NotificationAddr string
	SyncType         model.SyncType
}

// NewLocalClientConfig makes a new local client which do not synchronize with OrtooServer
func NewLocalClientConfig(collectionName string) *ClientConfig {
	return &ClientConfig{
		Address:          "",
		CollectionName:   collectionName,
		NotificationAddr: "",
		SyncType:         0,
	}
}
