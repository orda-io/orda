package ortoo

import "github.com/knowhunger/ortoo/ortoo/model"

// OrtooClientConfig is a configuration for OrtooClient
type OrtooClientConfig struct {
	Address          string
	CollectionName   string
	NotificationAddr string
	SyncType         model.SyncType
}
