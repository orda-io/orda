package commons

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/sirupsen/logrus"
)

type BaseDatatypeT struct {
	id     *datatypeUID
	opID   *operationID
	typeOf DatatypeType
	state  DatatypeState
	*log.OrtooLog
}

func newBaseDatatypeT(t DatatypeType) *BaseDatatypeT {
	loge := logrus.New()
	loge.SetFormatter(&logrus.TextFormatter{})
	return &BaseDatatypeT{
		id:       newDatatypeUID(),
		opID:     newOperationID(),
		typeOf:   t,
		state:    StateLocallyExisted,
		OrtooLog: log.NewOrtooLog(),
	}
}

func (b *BaseDatatypeT) String() string {
	return fmt.Sprintf("%s", b.id.String())
}

func (b *BaseDatatypeT) executeBase(datatype interface{}, op Operation) (interface{}, error) {
	op.SetOperationID(b.opID.Next())
	return op.executeLocal(datatype)
}
