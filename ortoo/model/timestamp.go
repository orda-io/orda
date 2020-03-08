package model

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

func (t *Timestamp) Compare(o *Timestamp) int {
	retEra := int32(t.Era - o.Era)
	if retEra > 0 {
		return 1
	} else if retEra < 0 {
		return -1
	}
	var diff = int64(t.Lamport - o.Lamport)
	if diff > 0 {
		return 1
	} else if diff < 0 {
		return -1
	}
	return bytes.Compare(t.CUID, o.CUID)
}

func (t *Timestamp) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "[%d:%d:%s:%d]", t.Era, t.Lamport, hex.EncodeToString(t.CUID), t.Delimiter)
	return b.String()
}
