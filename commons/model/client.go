package model

import (
	"encoding/hex"
	"fmt"
)

// GetCUIDString returns the string of CUID
func (c *Client) GetCUIDString() string {
	return hex.EncodeToString(c.CUID)
}

// ToString returns customized string
func (c *Client) ToString() string {
	return fmt.Sprintf("%s(%s)|%s|%s", c.Alias, hex.EncodeToString(c.CUID), c.SyncType.String(), c.Collection)
}
