package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/protocols"
)

type era uint32
type timeSeq uint64

func newOperationID() *OperationID {
	cuid := newNilCuid()
	return NewOperationIdWithCuid(cuid)
}

func NewOperationIdWithCuid(cuid *Cuid) *operationID {
	return &protocols.OperationId{
		Era:     0,
		Lamport: 0,
		Cuid:    []byte(cuid),
		Seq:     0,
	}
	return &operationID()
	//	era:     0,
	//	lamport: 0,
	//	cuid:    cuid,
	//	seq:     0,
	//}
}

func (o *protocols.OperationID) SetClient(cuid *Cuid) {
	o.cuid = cuid
}

func (o *OperationID) Next() *operationID {
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

func (o *operationID) toProtoBuf() *protocols.PbOperationId {
	return &protocols.PbOperationId{
		Era:     uint32(o.era),
		Lamport: uint64(o.lamport),
		Cuid:    nil,
		Seq:     uint64(o.seq),
	}
}

func pbToOperationID(pb *protocols.PbOperationId) (*operationID, error) {
	cuid, err := pbToUniqueID(pb.Cuid)
	if err != nil {
		return nil, log.OrtooError(err, "fail to decode PbOperationId.Cuid")
	}
	return &operationID{
		era:     era(pb.Era),
		lamport: timeSeq(pb.Lamport),
		cuid:    &Cuid{cuid},
		seq:     0,
	}, nil

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
