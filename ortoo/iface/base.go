package iface

import (
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
)

// PublicBaseDatatype is a public interface for a datatype.
type PublicBaseDatatype interface {
	GetType() model.TypeOfDatatype
	GetState() model.StateOfDatatype
	GetKey() string // @baseDatatype
}

type BaseDatatype interface {
	PublicBaseDatatype
	GetDatatype() Datatype                // @baseDatatype
	GetDUID() types.DUID                  // @baseDatatype
	GetCUID() string                      // @baseDatatype
	SetState(state model.StateOfDatatype) // @baseDatatype
}
