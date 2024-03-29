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

// Lock defines the lock interfaces
type Lock interface {
	TryLock() bool
	Unlock() bool
}

// GetLockName returns a lock name with the prefix
func GetLockName(prefix string, collectionNum int32, key string) string {
	return fmt.Sprintf("%s:%d:%s", prefix, collectionNum, key)
}
