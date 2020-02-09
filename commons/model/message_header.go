package model

import (
	"encoding/hex"
	"fmt"
)

func (m *MessageHeader) ToString() string {
	return fmt.Sprintf("v%s|%d|%s|%s|%s", m.Version, m.Seq, m.TypeOf.String(), m.Collection, hex.EncodeToString(m.Cuid))
}
