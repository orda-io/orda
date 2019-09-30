package datatypes

import (
	"bytes"
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type baseDatatype struct {
	Key           string
	id            model.Duid
	opID          *model.OperationID
	TypeOf        model.TypeOfDatatype
	state         model.StateOfDatatype
	finalDatatype model.FinalDatatype
	Logger        *log.OrtooLog
}

//PublicBaseDatatypeInterface is a public interface for a datatype.
type PublicBaseDatatypeInterface interface {
	GetType() model.TypeOfDatatype
}

func newBaseDatatype(key string, t model.TypeOfDatatype, cuid model.Cuid) (*baseDatatype, error) {
	duid, err := model.NewDuid()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to create base datatype due to duid")
	}
	return &baseDatatype{
		Key:    key,
		id:     duid,
		TypeOf: t,
		opID:   model.NewOperationIDWithCuid(cuid),
		state:  model.StateOfDatatype_LOCALLY_EXISTED,
		Logger: log.NewOrtooLogWithTag(fmt.Sprintf("%s", duid)[:8]),
	}, nil
}

func (b *baseDatatype) GetEra() uint32 {
	return b.opID.GetEra()
}

func (b *baseDatatype) String() string {
	return fmt.Sprintf("%s", b.id)
}

func (b *baseDatatype) executeLocalBase(op model.Operation) (interface{}, error) {
	b.SetNextOpID(op)
	return op.ExecuteLocal(b.finalDatatype)
}

func (b *baseDatatype) Replay(op model.Operation) error {
	if bytes.Compare(b.opID.Cuid, op.GetBase().Id.Cuid) == 0 {
		_, err := b.executeLocalBase(op)
		if err != nil {
			return log.OrtooErrorf(err, "fail to replay local operation")
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
	op.ExecuteRemote(b.finalDatatype)
}

func (b *baseDatatype) GetType() model.TypeOfDatatype {
	return b.TypeOf
}

func (b *baseDatatype) SetFinalDatatype(finalDatatype model.FinalDatatype) {
	b.finalDatatype = finalDatatype
}

func (b *baseDatatype) GetFinalDatatype() model.FinalDatatype {
	return b.finalDatatype
}

func (b *baseDatatype) SetOpID(opID *model.OperationID) {
	b.opID = opID
}

func (b *baseDatatype) GetKey() string {
	return b.Key
}
