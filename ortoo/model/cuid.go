package model

import (
	"github.com/knowhunger/ortoo/ortoo/log"
)

// CUID is a uniqueID for a client.
type CUID UniqueID

// NewCUID creates a new CUID
func NewCUID() (CUID, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate client UID")
	}
	return CUID(u), nil
}

// NewNilCUID creates an instance of Nil CUID.
func NewNilCUID() CUID {
	bin := make([]byte, 16)
	return bin
}

func (c *CUID) String() string {
	return UniqueID(*c).String()
}
