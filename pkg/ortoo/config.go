package ortoo

import "github.com/knowhunger/ortoo/pkg/model"

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
		NotificationAddr: "",
		CollectionName:   collectionName,
		SyncType:         model.SyncType_LOCAL_ONLY,
	}
}
