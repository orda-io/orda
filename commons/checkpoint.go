package commons

import . "github.com/knowhunger/ortoo/commons/protocols"

type CheckPoint struct {
	PbCheckPoint
}

func NewCheckPoint() *CheckPoint {
	return &CheckPoint{
		PbCheckPoint: PbCheckPoint{Sseq: 0, Cseq: 0}}
}

func (c *CheckPoint) GetSseq() uint64 {
	return c.GetSseq()
}
