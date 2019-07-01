package commons

type era uint32
type timeSeq uint64

type operationID struct {
	era     era
	lamport timeSeq
	cuid    *Cuid
	seq     timeSeq
}

func NewOperationId() *operationID {
	cuid := NewNilCuid()
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

func (c *operationID) SetClient(cuid *Cuid) {
	c.cuid = cuid
}

func (c *operationID) Next() operationID {
	c.lamport++
	c.seq++
	return operationID{c.era, c.lamport, c.cuid, c.seq}
}

func (c *operationID) GetTimestamp() *timestamp {
	return &timestamp{
		era:     c.era,
		lamport: c.lamport,
		cuid:    c.cuid,
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
