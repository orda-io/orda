package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/utils"
)

// ToString returns customized string
func (its *Client) ToString() string {
	return fmt.Sprintf("%s(%s)|%s|%s", its.Alias, its.CUID, its.SyncType.String(), its.Collection)
}

func (its *Client) GetSummary() string {
	return fmt.Sprintf("%s|%s(%s)", its.Collection, utils.MakeDefaultShort(its.Alias), its.CUID)
}
