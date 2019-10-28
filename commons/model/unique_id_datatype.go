package model

import (
	"github.com/knowhunger/ortoo/commons/log"
)

//DUID is the unique ID of datatype
type Duid uniqueID

//NewDuid creates a new DUID
func NewDuid() (Duid, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate datatype UID")
	}
	return Duid(u), nil
}

func (d Duid) String() string {
	return uniqueID(d).String()
}
