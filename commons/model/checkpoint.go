package model

//NewCheckPoint creates a new checkpoint
func NewCheckPoint() *CheckPoint {
	return &CheckPoint{
		Sseq: 0,
		Cseq: 0,
	}
}

//Set sets the values of checkpoint
func (c *CheckPoint) Set(sseq, cseq uint64) *CheckPoint {
	c.Sseq = sseq
	c.Cseq = cseq
	return c
}
