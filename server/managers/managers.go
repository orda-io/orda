package managers

import (
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/orda-io/orda/server/notification"
	"github.com/orda-io/orda/server/redis"
	"github.com/orda-io/orda/server/utils"
)

type Managers struct {
	Mongo    *mongodb.RepositoryMongo
	Notifier *notification.Notifier
	Redis    *redis.Client
}

func New(ctx context.OrdaContext, conf *OrdaServerConfig) (*Managers, errors.OrdaError) {
	var oErr errors.OrdaError
	clients := &Managers{}
	if clients.Mongo, oErr = mongodb.New(ctx, &conf.Mongo); oErr != nil {
		return clients, oErr
	}

	if clients.Notifier, oErr = notification.NewNotifier(ctx, conf.Notification); oErr != nil {
		return clients, oErr
	}

	if clients.Redis, oErr = redis.New(ctx, &conf.Redis); oErr != nil {
		return clients, oErr
	}
	return clients, nil
}

func (its *Managers) GetLock(ctx context.OrdaContext, lockName string) utils.Lock {
	return its.Redis.GetLock(ctx, lockName)
}

func (its *Managers) Close(ctx context.OrdaContext) {
	if err := its.Redis.Close(); err != nil {
		ctx.L().Errorf("fail to close redis: %v", err)
	}
	if err := its.Mongo.Close(ctx); err != nil {
		ctx.L().Errorf("fail to close mongo")
	}
}
