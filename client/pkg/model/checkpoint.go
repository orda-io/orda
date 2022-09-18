package model

import "fmt"

// NewCheckPoint creates a new checkpoint
func NewCheckPoint() *CheckPoint {
	return &CheckPoint{
		Sseq: 0,
		Cseq: 0,
	}
}

// NewSetCheckPoint creates a new checkpoint with set values
func NewSetCheckPoint(sseq, cseq uint64) *CheckPoint {
	return &CheckPoint{
		Sseq: sseq,
		Cseq: cseq,
	}
}

// Set sets the values of checkpoint
func (its *CheckPoint) Set(sseq, cseq uint64) *CheckPoint {
	its.Sseq = sseq
	its.Cseq = cseq
	return its
}

// SyncCseq syncs Cseq
func (its *CheckPoint) SyncCseq(cseq uint64) *CheckPoint {
	if its.Cseq < cseq {
		its.Cseq = cseq
	}
	return its
}

// Clone makes a carbon copy of this one.
func (its *CheckPoint) Clone() *CheckPoint {
	return NewCheckPoint().Set(its.Sseq, its.Cseq)
}

// Compare returns true if this CheckPoint is equal to other; otherwise, false.
func (its *CheckPoint) Compare(other *CheckPoint) bool {
	return its.Cseq == other.Cseq && its.Sseq == other.Sseq
}

// ToString returns a short string of CheckPoint
func (its *CheckPoint) ToString() string {
	if its == nil {
		return "nil"
	}
	return fmt.Sprintf("(s:%d c:%d)", its.Sseq, its.Cseq)
}
