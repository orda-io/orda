package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/types"
)

// ToString returns customized string
func (its *Client) ToString() string {
	return fmt.Sprintf("%s(%s)|%s|%s", its.Alias, types.ToShortUID(its.CUID), its.SyncType.String(), its.Collection)
}
