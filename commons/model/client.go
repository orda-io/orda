package model

import (
	"encoding/hex"
)

//GetCuidString returns the string of CUID
func (c *Client) GetCuidString() string {
	return hex.EncodeToString(c.Cuid)
}
