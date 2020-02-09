package model

import (
	"encoding/hex"
	"fmt"
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

// NewOperationIDWithCuid creates a new OperationID with CUID.
func NewOperationIDWithCuid(cuid CUID) *OperationID {
	return &OperationID{
		Era:     0,
		Lamport: 0,
		CUID:    cuid,
		Seq:     0,
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

func (o *OperationID) ToString() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "[%d:%d:%s:%d]", o.Era, o.Lamport, hex.EncodeToString(o.CUID), o.Seq)
	return b.String()
}

// Compare compares two operationIDs.
func Compare(a, b *OperationID) int {
	retEra := a.Era - b.Era
	if retEra > 0 {
		return 1
	} else if retEra < 0 {
		return -1
	}
	diff := a.Lamport - b.Lamport
	if diff > 0 {
		return 1
	} else if diff < 0 {
		return -1
	}
	return 0
}
