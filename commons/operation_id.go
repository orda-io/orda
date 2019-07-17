package commons

import "github.com/knowhunger/ortoo/commons/protocols"

type era uint32
type timeSeq uint64

type operationID struct {
	era     era
	lamport timeSeq
	cuid    *Cuid
	seq     timeSeq
}

func newOperationID() *operationID {
	cuid := newNilCuid()
	return NewOperationIdWithCuid(cuid)
}

func NewOperationIdWithCuid(cuid *Cuid) *operationID {
	return &operationID{
		era:     0,
		lamport: 0,
		cuid:    cuid,
		seq:     0,
	}
}

func (o *operationID) SetClient(cuid *Cuid) {
	o.cuid = cuid
}

func (o *operationID) Next() *operationID {
	o.lamport++
	o.seq++
	return &operationID{o.era, o.lamport, o.cuid, o.seq}
}

func (o *operationID) GetTimestamp() *timestamp {
	return &timestamp{
		era:     o.era,
		lamport: o.lamport,
		cuid:    o.cuid,
	}
}

func (o *operationID) syncLamport(other timeSeq) timeSeq {
	if o.lamport < other {
		o.lamport = other
	} else {
		o.lamport++
	}
	return o.lamport
}

func (o *operationID) getPB() *protocols.PbOperationId {
	return &protocols.PbOperationId{
		Era:     uint32(o.era),
		Lamport: uint64(o.lamport),
		Cuid:    nil,
		Seq:     uint64(o.seq),
	}
}

func Compare(a, b *operationID) int {
	retEra := a.era - b.era
	if retEra > 0 {
		return 1
	} else if retEra < 0 {
		return -1
	}
	retLamport := a.lamport - b.lamport
	if retLamport > 0 {
		return 1
	} else if retLamport < 0 {
		return -1
	}
	return 0
}
