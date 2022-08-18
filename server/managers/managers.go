package managers

import (
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/orda-io/orda/server/notification"
	"github.com/orda-io/orda/server/redis"
	"github.com/orda-io/orda/server/utils"
)

// Managers are a bundle of infra
type Managers struct {
	Mongo    *mongodb.RepositoryMongo
	Notifier *notification.Notifier
	Redis    *redis.Client
}

// New creates Managers with context and config
func New(ctx context.OrdaContext, conf *OrdaServerConfig) (*Managers, errors.OrdaError) {
	var oErr errors.OrdaError
	clients := &Managers{}
	if clients.Mongo, oErr = mongodb.New(ctx, conf.Mongo); oErr != nil {
		return clients, oErr
	}

	if clients.Notifier, oErr = notification.NewNotifier(ctx, conf.Notification); oErr != nil {
		return clients, oErr
	}

	if clients.Redis, oErr = redis.New(ctx, conf.Redis); oErr != nil {
		return clients, oErr
	}
	return clients, nil
}

// GetLock returns either a local or redis lock
func (its *Managers) GetLock(ctx context.OrdaContext, lockName string) utils.Lock {
	return its.Redis.GetLock(ctx, lockName)
}

// Close closes this Managers
func (its *Managers) Close(ctx context.OrdaContext) {
	if err := its.Redis.Close(); err != nil {
		ctx.L().Errorf("fail to close redis: %v", err)
	}
	if err := its.Mongo.Close(ctx); err != nil {
		ctx.L().Errorf("fail to close mongo")
	}
}
