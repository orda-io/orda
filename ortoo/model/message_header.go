package model

import (
	"encoding/hex"
	"fmt"
)

// ToString returns customized string
func (m *MessageHeader) ToString() string {
	return fmt.Sprintf("v%s|%d|%s|%s|%s",
		m.Version, m.Seq, m.TypeOf.String(), m.Collection, hex.EncodeToString(m.Cuid)[0:8])
}
