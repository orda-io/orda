package model

// NewCheckPoint creates a new checkpoint
func NewCheckPoint() *CheckPoint {
	return &CheckPoint{
		Sseq: 0,
		Cseq: 0,
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
