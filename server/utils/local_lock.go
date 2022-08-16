package utils

import (
	ctx "context"
	"github.com/orda-io/orda/client/pkg/context"
	golock "github.com/viney-shih/go-lock"
	"sync"
)

var localLockMap sync.Map

// LocalLock is used for local locking in a single server
type LocalLock struct {
	ctx      context.OrdaContext
	mutex    *golock.CASMutex
	lockName string
}

// GetLocalLock returns a LocalLock with the specified name
func GetLocalLock(ctx context.OrdaContext, lockName string) *LocalLock {

	value, loaded := localLockMap.LoadOrStore(lockName, &LocalLock{
		ctx:      ctx,
		mutex:    golock.NewCASMutex(),
		lockName: lockName,
	})
	if loaded {
		ctx.L().Infof("[🔒] load lock '%v'", lockName)
	} else {
		ctx.L().Infof("[🔒] create lock '%v'", lockName)
	}
	return value.(*LocalLock)
}

// TryLock tries to a local lock, and returns true if it succeeds; otherwise false
func (its *LocalLock) TryLock() bool {
	timeCtx, cancel := ctx.WithTimeout(its.ctx, defaultLeaseTime)
	defer cancel()
	if !its.mutex.TryLockWithContext(timeCtx) {
		if err := timeCtx.Err(); err != nil {
			its.ctx.L().Warnf("[🔒] fail to lock '%v':%v", its.lockName, err.Error())
			return false
		}
		its.ctx.L().Warnf("[🔒] fail to lock '%v'", its.lockName)
		return false
	}

	ts, _ := timeCtx.Deadline()
	its.ctx.L().Infof("[🔒] lock '%v': %v", its.lockName, ts)
	return true
}

// Unlock unlocks the local lock, and returns true if it succeeds; otherwise false
func (its *LocalLock) Unlock() bool {
	its.mutex.Unlock()
	its.ctx.L().Infof("[🔒] unlock '%v'", its.lockName)
	return true
}
