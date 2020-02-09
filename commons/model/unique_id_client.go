package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

// CUID is a uniqueID
type CUID UniqueID

// NewCUID creates a new CUID
func NewCUID() (CUID, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate client UID")
	}
	return CUID(u), nil
}

func NewNilCUID() CUID {
	bin := make([]byte, 16)
	return bin
}

func (c *CUID) String() string {
	return UniqueID(*c).String()
}
