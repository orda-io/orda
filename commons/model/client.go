package model

import (
	"encoding/hex"
)

//GetCUIDString returns the string of CUID
func (c *Client) GetCUIDString() string {
	return hex.EncodeToString(c.Cuid)
}
