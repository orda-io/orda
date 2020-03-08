package serverside

import (
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// NewDatatype generates a Datatype with the specified key and type, which is used in the server side.
func NewDatatype(key string, typeOf model.TypeOfDatatype) (model.Datatype, error) {
	var internal model.Datatype
	switch typeOf {
	case model.TypeOfDatatype_INT_COUNTER:
		ic, err := ortoo.NewIntCounter(key, model.NewNilCUID(), nil, nil)
		icImpl := ic.(datatypes.FinalDatatypeInterface)

		if err != nil {
			return nil, log.OrtooError(err)
		}
		internal = icImpl.GetFinal().GetDatatype()
	}
	return internal, nil
}

// SetSnapshot sets the snapshot for the given Datatype.
func SetSnapshot(datatype model.Datatype, meta []byte, snap string) error {
	return datatype.SetMetaAndSnapshot(meta, snap)
}
