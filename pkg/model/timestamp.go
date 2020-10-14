package model

import (
	"bytes"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
	"strings"
)

// OldestTimestamp returns the oldest timestamp.
func OldestTimestamp() *Timestamp {
	return &Timestamp{
		Era:       0,
		Lamport:   0,
		CUID:      make([]byte, 16),
		Delimiter: 0,
	}
}

func NewTimestamp(era uint32, lamport uint64, cuid []byte, delimiter uint32) *Timestamp {
	return &Timestamp{
		Era:       era,
		Lamport:   lamport,
		CUID:      cuid,
		Delimiter: delimiter,
	}
}

// Compare is used to compared with another Timestamp.
func (its *Timestamp) Compare(o *Timestamp) int {
	retEra := int32(its.Era - o.Era)
	if retEra > 0 {
		return 1
	} else if retEra < 0 {
		return -1
	}
	var diff = int64(its.Lamport - o.Lamport)
	if diff > 0 {
		return 1
	} else if diff < 0 {
		return -1
	}
	return bytes.Compare(its.CUID, o.CUID)
}

// ToString is used to get string for Timestamp
func (its *Timestamp) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "[%d:%d:%s:%d]", its.Era, its.Lamport,
		types.UIDtoShortString(its.CUID), its.Delimiter)
	return b.String()
}

// Hash returns the string hash of timestamp.
// DON'T change this because protocol can be broken : TODO: this can be improved.
func (its *Timestamp) Hash() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "%d%d%s%d", its.Era, its.Lamport, types.UIDtoString(its.CUID), its.Delimiter)
	return b.String()
}

// Next returns a next Timestamp having increased Lamport.
func (its *Timestamp) Next() *Timestamp {
	return &Timestamp{
		Era:       its.Era,
		Lamport:   its.Lamport + 1,
		CUID:      its.CUID,
		Delimiter: 0,
	}
}

// GetAndNextDelimiter returns a next Timestamp having increased deliminator.
func (its *Timestamp) GetAndNextDelimiter() *Timestamp {

	ts := &Timestamp{
		Era:       its.Era,
		Lamport:   its.Lamport,
		CUID:      its.CUID,
		Delimiter: its.Delimiter,
	}
	its.Delimiter++
	return ts
}

// Clone returns clone of this Timestamp.
func (its *Timestamp) Clone() *Timestamp {
	return &Timestamp{
		Era:       its.Era,
		Lamport:   its.Lamport,
		CUID:      its.CUID,
		Delimiter: its.Delimiter,
	}
}
