package utils

import (
	ctx "context"
	"github.com/orda-io/orda/pkg/context"
	golock "github.com/viney-shih/go-lock"
	"sync"
)

var localLockMap sync.Map

type LocalLock struct {
	ctx      context.OrdaContext
	mutex    *golock.CASMutex
	lockName string
}

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

func (its *LocalLock) Unlock() bool {
	its.mutex.Unlock()
	its.ctx.L().Infof("[🔒] unlock '%v'", its.lockName)
	return true
}
