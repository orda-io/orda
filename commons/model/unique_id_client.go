package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

type Cuid uniqueID

func NewCuid() (Cuid, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooError(err, "fail to generate client UID")
	}
	return Cuid(u), nil
}

func newNilCuid() Cuid {
	bin := make([]byte, 16)
	return Cuid(bin)
}

func (c *Cuid) String() string {
	return uniqueID(*c).String()
}

func (c *Cuid) Compare(o *Cuid) {

}
