package datatypes

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type baseDatatype struct {
	id     model.Duid
	opID   *model.OperationID
	typeOf model.TypeDatatype
	state  model.StateDatatype
	*log.OrtooLog
}

type PublicBaseDatatypeInterface interface {
	GetType() model.TypeDatatype
}

func newBaseDatatype(t model.TypeDatatype) (*baseDatatype, error) {
	duid, err := model.NewDuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create base datatype due to duid")
	}
	return &baseDatatype{
		id:       duid,
		opID:     model.NewOperationID(),
		typeOf:   t,
		state:    model.StateDatatype_LOCALLY_EXISTED,
		OrtooLog: log.NewOrtooLog(),
	}, nil
}

func (b *baseDatatype) String() string {
	return fmt.Sprintf("%s", b.id)
}

func (b *baseDatatype) executeLocalBase(datatype model.OperationExecuter, op model.Operation) (interface{}, error) {
	b.SetNextOpID(op)
	return op.ExecuteLocal(datatype)
}

func (b *baseDatatype) SetNextOpID(op model.Operation) {
	op.GetBase().SetOperationID(b.opID.Next())
}

func (b *baseDatatype) executeRemoteBase(datatype model.OperationExecuter, op model.Operation) {
	op.ExecuteRemote(datatype)
}

func (b *baseDatatype) GetType() model.TypeDatatype {
	return b.typeOf
}
