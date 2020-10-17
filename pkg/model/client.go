package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
	"github.com/knowhunger/ortoo/pkg/utils"
)

// ToString returns customized string
func (its *Client) ToString() string {
	return fmt.Sprintf("%s(%s)|%s|%s", its.Alias, types.UIDtoShortString(its.CUID), its.SyncType.String(), its.Collection)
}

func (its *Client) GetSummary() string {
	return utils.MakeSummary(its.Alias, its.CUID, true)
}
