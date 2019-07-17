package commons

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
)

type BaseDatatypeT struct {
	id     *Duid
	opID   *operationID
	typeOf DatatypeType
	state  DatatypeState
	*log.OrtooLog
}

func newBaseDatatypeT(t DatatypeType) (*BaseDatatypeT, error) {
	duid, err := newDuid()
	if err != nil {
		return nil, log.OrtooError(err, "fail to create base datatype due to duid")
	}
	return &BaseDatatypeT{
		id:       duid,
		opID:     newOperationID(),
		typeOf:   t,
		state:    StateLocallyExisted,
		OrtooLog: log.NewOrtooLog(),
	}, nil
}

func (b *BaseDatatypeT) String() string {
	return fmt.Sprintf("%s", b.id.String())
}

func (b *BaseDatatypeT) executeBase(datatype interface{}, op Operation) (interface{}, error) {
	op.SetOperationID(b.opID.Next())
	return op.executeLocal(datatype)
}
