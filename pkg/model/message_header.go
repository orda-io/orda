package model

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
)

// ToString returns customized string
func (m *MessageHeader) ToString() string {
	return fmt.Sprintf("v%s|%d|%s|%s|%s",
		m.Version, m.Seq, m.TypeOf.String(), m.Collection, types.UID(m.Cuid).ShortString())
}
