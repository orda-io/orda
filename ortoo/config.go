package ortoo

import "github.com/knowhunger/ortoo/ortoo/model"

// OrtooClientConfig is a configuration for OrtooClient
type ClientConfig struct {
	Address          string
	CollectionName   string
	NotificationAddr string
	SyncType         model.SyncType
}

func NewLocalClientConfig(collectionName string) *ClientConfig {
	return &ClientConfig{
		Address:          "",
		CollectionName:   collectionName,
		NotificationAddr: "",
		SyncType:         0,
	}
}
