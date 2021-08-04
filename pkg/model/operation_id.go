package model

import (
	"fmt"
	"github.com/orda-io/orda/pkg/types"
	"strings"
)

// NewOperationID creates a new OperationID.
func NewOperationID() *OperationID {
	return &OperationID{
		Era:     0,
		Lamport: 0,
		CUID:    types.NewNilUID(),
		Seq:     0,
	}
}

// NewOperationIDWithCUID creates a new OperationID with CUID.
func NewOperationIDWithCUID(cuid string) *OperationID {
	return &OperationID{
		Era:     0,
		Lamport: 0,
		CUID:    cuid,
		Seq:     0,
	}
}

// GetTimestamp returns Timestamp from OperationID
func (its *OperationID) GetTimestamp() *Timestamp {
	return &Timestamp{
		Era:       its.Era,
		Lamport:   its.Lamport,
		CUID:      its.CUID,
		Delimiter: 0,
	}
}

// SetOperationID sets the values of OperationID.
func (its *OperationID) SetOperationID(other *OperationID) {
	its.Era = other.Era
	its.Lamport = other.Lamport
	its.CUID = other.CUID
	its.Seq = other.Seq

}

// Next increments an OperationID
func (its *OperationID) Next() *OperationID {
	its.Lamport++
	its.Seq++
	return &OperationID{
		Era:     its.Era,
		Lamport: its.Lamport,
		CUID:    its.CUID,
		Seq:     its.Seq,
	}
}

func (its *OperationID) RollBack() {
	its.Lamport--
	its.Seq--
}

// SyncLamport synchronizes the value of Lamport.
func (its *OperationID) SyncLamport(other uint64) uint64 {
	if its.Lamport < other {
		its.Lamport = other
	} else {
		its.Lamport++
	}
	return its.Lamport
}

// Clone ...
func (its *OperationID) Clone() *OperationID {
	return &OperationID{
		Era:     its.Era,
		Lamport: its.Lamport,
		CUID:    its.CUID,
		Seq:     its.Seq,
	}
}

// ToString returns customized string
func (its *OperationID) ToString() string {
	if its == nil {
		return ""
	}
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "[%d:%d:%s:%d]",
		its.Era, its.Lamport, its.CUID, its.Seq)
	return b.String()
}

func (its *OperationID) ToJSON() interface{} {
	return struct {
		Era     uint32
		Lamport uint64
		CUID    string
		Seq     uint64
	}{
		Era:     its.Era,
		Lamport: its.Lamport,
		CUID:    its.CUID,
		Seq:     its.Seq,
	}
}

// Compare compares two operationIDs.
func (its *OperationID) Compare(other *OperationID) int {
	retEra := int32(its.Era - other.Era)
	if retEra > 0 {
		return 1
	} else if retEra < 0 {
		return -1
	}
	var diff = int64(its.Lamport - other.Lamport)
	if diff > 0 {
		return 1
	} else if diff < 0 {
		return -1
	}
	return strings.Compare(its.CUID, other.CUID)
}
