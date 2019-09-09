package model

import (
	"encoding/hex"
)

func (c *Client) GetCuidString() string {
	return hex.EncodeToString(c.Cuid)
}
