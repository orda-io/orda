package orda

import "github.com/orda-io/orda/pkg/model"

// ClientConfig is a configuration for OrdaClient
type ClientConfig struct {
	ServerAddr       string
	NotificationAddr string
	CollectionName   string
	SyncType         model.SyncType
}

// NewLocalClientConfig makes a new local client which do not synchronize with OrdaServer
func NewLocalClientConfig(collectionName string) *ClientConfig {
	return &ClientConfig{
		ServerAddr:       "",
		NotificationAddr: "",
		CollectionName:   collectionName,
		SyncType:         model.SyncType_LOCAL_ONLY,
	}
}
