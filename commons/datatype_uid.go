package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
)

type datatypeUID struct {
	*uniqueID
}

func newDatatypeUID() (*datatypeUID, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooError(err, "cannot generate datatype UID")
	}
	return &datatypeUID{u}, nil
}
