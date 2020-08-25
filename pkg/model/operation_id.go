package model

import (
	"bytes"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
	"strings"
)

// NewOperationID creates a new OperationID.
func NewOperationID() *OperationID {
	return &OperationID{
		Era:     0,
		Lamport: 0,
		CUID:    make([]byte, 16),
		Seq:     0,
	}
}

// NewOperationIDWithCUID creates a new OperationID with CUID.
func NewOperationIDWithCUID(cuid []byte) *OperationID {
	return &OperationID{
		Era:     0,
		Lamport: 0,
		CUID:    cuid,
		Seq:     0,
	}
}

// GetTimestamp returns Timestamp from OperationID
func (o *OperationID) GetTimestamp() *Timestamp {
	return &Timestamp{
		Era:       o.Era,
		Lamport:   o.Lamport,
		CUID:      o.CUID,
		Delimiter: 0,
	}
}

// SetOperationID sets the values of OperationID.
func (o *OperationID) SetOperationID(other *OperationID) {
	o.Era = other.Era
	o.Lamport = other.Lamport
	o.CUID = other.CUID
	o.Seq = other.Seq

}

// Next increments an OperationID
func (o *OperationID) Next() *OperationID {
	o.Lamport++
	o.Seq++
	return &OperationID{
		Era:     o.Era,
		Lamport: o.Lamport,
		CUID:    o.CUID,
		Seq:     o.Seq,
	}
}

// SyncLamport synchronizes the value of Lamport.
func (o *OperationID) SyncLamport(other uint64) uint64 {
	if o.Lamport < other {
		o.Lamport = other
	} else {
		o.Lamport++
	}
	return o.Lamport
}

// SetClient sets clientID
func (o *OperationID) SetClient(cuid []byte) {
	o.CUID = cuid
}

// Clone ...
func (o *OperationID) Clone() *OperationID {
	return &OperationID{
		Era:     o.Era,
		Lamport: o.Lamport,
		CUID:    o.CUID,
		Seq:     o.Seq,
	}
}

// ToString returns customized string
func (o *OperationID) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "[%d:%d:%s:%d]",
		o.Era, o.Lamport, types.UID(o.CUID).ShortString(), o.Seq)
	return b.String()
}

// Compare compares two operationIDs.
func (o *OperationID) Compare(other *OperationID) int {
	retEra := int32(o.Era - other.Era)
	if retEra > 0 {
		return 1
	} else if retEra < 0 {
		return -1
	}
	var diff = int64(o.Lamport - other.Lamport)
	if diff > 0 {
		return 1
	} else if diff < 0 {
		return -1
	}
	return bytes.Compare(o.CUID, other.CUID)
}
