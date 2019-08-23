package model

func NewOperationID() *OperationID {
	return &OperationID{
		Era:     0,
		Lamport: 0,
		Cuid:    make([]byte, 16),
		Seq:     0,
	}
}

func NewOperationIDWithCuid(cuid *Cuid) *OperationID {
	return &OperationID{
		Era:     0,
		Lamport: 0,
		Cuid:    []byte(*cuid),
		Seq:     0,
	}
}

func (o *OperationID) SetOperationID(other *OperationID) {
	o.Era = other.Era
	o.Lamport = other.Lamport
	o.Cuid = other.Cuid
	o.Seq = other.Seq

}

func (o *OperationID) Next() *OperationID {
	o.Lamport++
	o.Seq++
	return &OperationID{
		Era:     o.Era,
		Lamport: o.Lamport,
		Cuid:    o.Cuid,
		Seq:     o.Seq,
	}
}

func (o *OperationID) SyncLamport(other uint64) uint64 {
	if o.Lamport < other {
		o.Lamport = other
	} else {
		o.Lamport++
	}
	return o.Lamport
}

func (o *OperationID) SetClient(cuid []byte) {
	o.Cuid = cuid
}

func (o *OperationID) Clone() *OperationID {
	return &OperationID{
		Era:     o.Era,
		Lamport: o.Lamport,
		Cuid:    o.Cuid,
		Seq:     o.Seq,
	}
}

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
