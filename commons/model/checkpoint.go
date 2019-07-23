package model

func NewCheckPoint() *CheckPoint {
	return &CheckPoint{
		Sseq: 0,
		Cseq: 0,
	}
}

func (c *CheckPoint) Set(sseq, cseq uint64) {
	c.Sseq = sseq
	c.Cseq = cseq
}
