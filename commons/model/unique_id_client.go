package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

//Cuid is a uniqueID
type Cuid uniqueID

//NewCuid creates a new CUID
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

//Compare makes two CUID compared
func (c *Cuid) Compare(o *Cuid) {

}
