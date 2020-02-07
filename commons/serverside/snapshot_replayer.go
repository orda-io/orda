package serverside

import (
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type ServerSideDatatype struct {
	*datatypes.CommonDatatype
	model.FinalDatatype
}

func NewFinalDatatype(key string, typeOf model.TypeOfDatatype) (model.FinalDatatype, error) {
	var internal model.FinalDatatype
	switch typeOf {
	case model.TypeOfDatatype_INT_COUNTER:
		data, err := commons.NewIntCounter(key, model.NewNilCUID(), nil)
		if err != nil {
			return nil, log.OrtooError(err)
		}
		internal = data.GetFinalDatatype()
	}
	return internal, nil
}

func SetSnapshot(datatype model.FinalDatatype, meta []byte, snap string) error {
	return datatype.SetMetaAndSnapshot(meta, snap)
}
