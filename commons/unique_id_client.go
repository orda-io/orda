package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
)

type Cuid uniqueID

func newCuid() (Cuid, error) {
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
