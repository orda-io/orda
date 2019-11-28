package serverside

import (
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/model"
)

type ServerSideDatatype struct {
	*datatypes.CommonDatatype
	model.FinalDatatype
}

func NewFinalDatatype(key string, datatype model.TypeOfDatatype) model.FinalDatatype {
	data, err := commons.NewIntCounter(key, model.NewNilCUID(), nil)
	if err != nil {

	}
	internal := data.(*datatypes.IntCounter)
	return internal.GetFinalDatatype()
}
