package model

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/log"
)

// ToString returns customized string
func (its *Client) ToString() string {
	return fmt.Sprintf("%s(%s)|%s|%s", its.Alias, its.CUID, its.SyncType.String(), its.Collection)
}

// GetSummary returns the summary of client
func (its *Client) GetSummary() string {
	return fmt.Sprintf("%s|%s(%s)", its.Collection, log.MakeDefaultShort(its.Alias), its.CUID)
}
