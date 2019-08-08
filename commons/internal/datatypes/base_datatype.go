package datatypes

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type baseDatatypeImpl struct {
	id     model.Duid
	opID   *model.OperationID
	typeOf model.TypeDatatype
	state  model.StateDatatype
	*log.OrtooLog
}

type PublicBaseInterface interface {
	GetType() model.TypeDatatype
}

func newBaseDatatype(t model.TypeDatatype) (*baseDatatypeImpl, error) {
	duid, err := model.NewDuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create base datatype due to duid")
	}
	return &baseDatatypeImpl{
		id:       duid,
		opID:     model.NewOperationID(),
		typeOf:   t,
		state:    model.StateDatatype_LOCALLY_EXISTED,
		OrtooLog: log.NewOrtooLog(),
	}, nil
}

func (b *baseDatatypeImpl) String() string {
	return fmt.Sprintf("%s", b.id)
}

func (b *baseDatatypeImpl) executeLocalBase(datatype model.OperationExecuter, op model.Operationer) (interface{}, error) {
	op.GetBase().SetOperationID(b.opID.Next())
	return op.ExecuteLocal(datatype)
}

func (b *baseDatatypeImpl) executeRemoteBase(datatype model.OperationExecuter, op model.Operationer) {
	op.ExecuteRemote(datatype)
}

func (b *baseDatatypeImpl) GetType() model.TypeDatatype {
	return b.typeOf
}
