package datatypes

import (
	"bytes"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type baseDatatype struct {
	id         model.Duid
	opID       *model.OperationID
	typeOf     model.TypeDatatype
	state      model.StateDatatype
	opExecuter model.OperationExecuter
	Logger     *log.OrtooLog
}

//PublicBaseDatatypeInterface is a public interface for a datatype.
type PublicBaseDatatypeInterface interface {
	GetType() model.TypeDatatype
}

func newBaseDatatype(t model.TypeDatatype) (*baseDatatype, error) {
	duid, err := model.NewDuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create base datatype due to duid")
	}
	return &baseDatatype{
		id:     duid,
		opID:   model.NewOperationID(),
		typeOf: t,
		state:  model.StateDatatype_LOCALLY_EXISTED,
		Logger: log.NewOrtooLogWithTag(fmt.Sprintf("%s", duid)[:8]),
	}, nil
}

func (b *baseDatatype) String() string {
	return fmt.Sprintf("%s", b.id)
}

func (b *baseDatatype) executeLocalBase(op model.Operation) (interface{}, error) {
	b.SetNextOpID(op)
	return op.ExecuteLocal(b.opExecuter)
}

func (b *baseDatatype) Replay(op model.Operation) error {
	if bytes.Compare(b.opID.Cuid, op.GetBase().Id.Cuid) == 0 {
		_, err := b.executeLocalBase(op)
		if err != nil {
			return log.OrtooError(err, "fail to replay local operation")
		}
	} else {
		b.executeRemoteBase(op)
	}
	return nil
}

func (b *baseDatatype) SetNextOpID(op model.Operation) {
	op.GetBase().SetOperationID(b.opID.Next())
}

func (b *baseDatatype) executeRemoteBase(op model.Operation) {
	op.ExecuteRemote(b.opExecuter)
}

func (b *baseDatatype) GetType() model.TypeDatatype {
	return b.typeOf
}

func (b *baseDatatype) SetOperationExecuter(opExecuter model.OperationExecuter) {
	b.opExecuter = opExecuter
}

func (b *baseDatatype) SetOpID(opID *model.OperationID) {
	b.opID = opID
}
