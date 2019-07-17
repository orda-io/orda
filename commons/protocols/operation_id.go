package protocols

func NewOperationID() *OperationID {

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
