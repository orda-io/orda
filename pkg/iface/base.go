package iface

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/types"
)

// PublicBaseDatatype is a public interface for a datatype.
type PublicBaseDatatype interface {
	GetType() model.TypeOfDatatype
	GetState() model.StateOfDatatype
	GetKey() string // @baseDatatype
}

// BaseDatatype defines a base operations for datatype
type BaseDatatype interface {
	PublicBaseDatatype
	GetDatatype() Datatype                // @baseDatatype
	GetDUID() types.DUID                  // @baseDatatype
	GetCUID() string                      // @baseDatatype
	SetState(state model.StateOfDatatype) // @baseDatatype
	SetLogger(l *log.OrtooLog)            // @baseDatatype
	GetMeta() ([]byte, errors.OrtooError)
	SetMeta(meta []byte) errors.OrtooError
	GetLogger() *log.OrtooLog
	GetOpID() *model.OperationID
}
