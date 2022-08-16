package utils

import (
	ctx "context"
	"github.com/go-redsync/redsync/v4"
	"github.com/orda-io/orda/client/pkg/context"
)

// RedisLock is used for distributed locking among multiple servers
type RedisLock struct {
	ctx      context.OrdaContext
	lockName string
	mutex    *redsync.Mutex
}

// GetRedisLock returns a RedisLock with the specified name
func GetRedisLock(ctx context.OrdaContext, lockName string, rs *redsync.Redsync) *RedisLock {
	mutex := rs.NewMutex(prefix+lockName, redsync.WithTries(100), redsync.WithExpiry(defaultExpireTime))
	return &RedisLock{
		ctx:      ctx,
		lockName: lockName,
		mutex:    mutex,
	}
}

// TryLock tries to lock by redis, and returns true if it succeeds; otherwise false
func (its *RedisLock) TryLock() bool {
	timeCtx, cancel := ctx.WithTimeout(its.ctx, defaultLeaseTime)
	defer cancel()
	if err := its.mutex.LockContext(timeCtx); err != nil {
		its.ctx.L().Warnf("[🔒] fail to lock '%v': %v", its.lockName, err.Error())
		return false
	}
	if err := timeCtx.Err(); err != nil {
		its.ctx.L().Warnf("[🔒] fail to lock '%v':%v", its.lockName, err.Error())
		return false
	}
	its.ctx.L().Infof("[🔒] lock '%v': %v", its.lockName, its.mutex.Until())
	return true
}

// Unlock unlocks the redis lock, and returns true if it succeeds; otherwise false
func (its *RedisLock) Unlock() bool {
	success, err := its.mutex.Unlock()
	if err != nil {
		its.ctx.L().Errorf("[🔒] fail to unlock '%v': %v", its.lockName, err.Error())
		return false
	}
	if success {
		its.ctx.L().Infof("[🔒] unlock '%v'", its.lockName)
		return true
	}
	its.ctx.L().Warnf("[🔒] something wrong with lock '%v'", its.lockName)
	return false
}
