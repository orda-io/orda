package commons

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

type baseDatatypeT struct {
	id     model.Duid
	opID   *model.OperationID
	typeOf DatatypeType
	state  DatatypeState
	*log.OrtooLog
}

func newBaseDatatypeT(t DatatypeType) (*baseDatatypeT, error) {
	duid, err := model.NewDuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create base datatype due to duid")
	}
	return &baseDatatypeT{
		id:       duid,
		opID:     model.NewOperationID(),
		typeOf:   t,
		state:    StateLocallyExisted,
		OrtooLog: log.NewOrtooLog(),
	}, nil
}

func (b *baseDatatypeT) String() string {
	return fmt.Sprintf("%s", b.id)
}

func (b *baseDatatypeT) executeBase(datatype model.OperationExecuter, op model.Operationer) (interface{}, error) {
	op.GetBase().SetOperationID(b.opID.Next())
	return op.ExecuteLocal(datatype)
}
