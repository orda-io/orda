package utils

import (
	"fmt"
	"time"
)

const (
	prefix            = "orda:__lock__:"
	defaultLeaseTime  = 5 * time.Second
	defaultExpireTime = 10 * time.Second
)

type Lock interface {
	TryLock() bool
	Unlock() bool
}

func GetLockName(prefix string, collectionNum uint32, key string) string {
	return fmt.Sprintf("%s:%d:%s", prefix, collectionNum, key)
}
