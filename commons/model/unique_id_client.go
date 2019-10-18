package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

//CUID is a uniqueID
type CUID uniqueID

//NewCuid creates a new CUID
func NewCuid() (CUID, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate client UID")
	}
	return CUID(u), nil
}

func newNilCuid() CUID {
	bin := make([]byte, 16)
	return CUID(bin)
}

func (c *CUID) String() string {
	return uniqueID(*c).String()
}

//Compare makes two CUID compared
func (c *CUID) Compare(o *CUID) {

}
