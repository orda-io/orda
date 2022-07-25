package redis

import (
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/orda-io/orda/client/pkg/context"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/server/utils"
)

type Client struct {
	rs       *redsync.Redsync
	ctx      context.OrdaContext
	client   goredislib.UniversalClient
	mutexMap map[string]*redsync.Mutex
}

func New(ctx context.OrdaContext, conf *Config) (*Client, errors2.OrdaError) {
	if conf == nil {
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
	return &Client{
		rs:       rs,
		ctx:      ctx,
		client:   client,
		mutexMap: mutexMap,
	}, nil
}

func (its *Client) GetLock(ctx context.OrdaContext, lockName string) utils.Lock {
	if its.rs == nil {
		return utils.GetLocalLock(ctx, lockName)
	}
	return utils.GetRedisLock(ctx, lockName, its.rs)
}

func (its *Client) Close() errors2.OrdaError {
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
		return errors2.ServerInit.New(its.ctx.L(), "[🔒] fail to close redis")
	}
	return nil
}
