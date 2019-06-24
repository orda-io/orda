package commons

import (
	"github.com/google/uuid"
)

type operationID struct {
	era     uint32
	lamport uint64
	cuid    *CUID
	seq     uint64
}

func NewOperationId() *operationID {
	cuid := CUID(uuid.Nil)
	return NewOperationIdWithCuid(cuid)
}

func NewOperationIdWithCuid(cuid CUID) *operationID {
	return &operationID{
		era:     0,
		lamport: 0,
		cuid:    &cuid,
		seq:     0,
	}
}

func (c *operationID) SetClient(cuid *CUID) {
	c.cuid = cuid
}

func (c *operationID) Next() operationID {
	c.lamport++
	c.seq++
	return operationID{c.era, c.lamport, c.cuid, c.seq}
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
