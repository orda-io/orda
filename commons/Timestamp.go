package commons

type timestamp struct {
	era     era
	lamport timeSeq
	cuid    *Cuid
}

func NewTimestamp(era era, lamport timeSeq, cuid *Cuid) *timestamp {
	return &timestamp{
		era:     era,
		lamport: lamport,
		cuid:    cuid,
	}
}
