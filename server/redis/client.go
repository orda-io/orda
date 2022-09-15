package redis

import (
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/server/utils"
)

// Client is used to manage redis client
type Client struct {
	rs       *redsync.Redsync
	ctx      iface.OrdaContext
	client   goredislib.UniversalClient
	mutexMap map[string]*redsync.Mutex
}

// New creates a redis client
func New(ctx iface.OrdaContext, conf *Config) (*Client, errors.OrdaError) {
	if conf == nil || conf.Addrs == nil {
		ctx.L().Infof("redis is NOT initialized")
		return &Client{
			ctx: ctx,
		}, nil
	}
	options := &goredislib.UniversalOptions{
		Addrs:    conf.Addrs,
		Username: conf.Username,
		Password: conf.Password,
	}
	client := goredislib.NewUniversalClient(options)
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	mutexMap := make(map[string]*redsync.Mutex)
	ctx.L().Infof("redis is initialized")
	return &Client{
		rs:       rs,
		ctx:      ctx,
		client:   client,
		mutexMap: mutexMap,
	}, nil
}

// GetLock gets a lock of redis lock. If redis is not available, a local lock is gotten
func (its *Client) GetLock(ctx iface.OrdaContext, lockName string) utils.Lock {
	if its.rs == nil {
		return utils.GetLocalLock(ctx, lockName)
	}
	return utils.GetRedisLock(ctx, lockName, its.rs)
}

// Close closes redis Client
func (its *Client) Close() errors.OrdaError {
	if its.client == nil {
		return nil
	}
	for s, mutex := range its.mutexMap {
		if unlock, err := mutex.Unlock(); err != nil {
			its.ctx.L().Warnf("[🔒] fail to unlock '%v'", s)
		} else {
			if !unlock {

			}
		}
	}
	if err := its.client.Close(); err != nil {
		return errors.ServerInit.New(its.ctx.L(), "[🔒] fail to close redis")
	}
	return nil
}
