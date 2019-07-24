package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

type Duid uniqueID

func NewDuid() (Duid, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooError(err, "fail to generate datatype UID")
	}
	return Duid(u), nil
}

func (d Duid) String() string {
	return d.String()
}
