package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
)

type Duid uniqueID

func newDuid() (Duid, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooError(err, "fail to generate datatype UID")
	}
	return Duid(u), nil
}
