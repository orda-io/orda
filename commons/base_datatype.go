package commons

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type BaseDatatypeT struct {
	id     *datatypeID
	opID   *operationID
	typeOf DatatypeType
	state  DatatypeState
	logger *logrus.Logger
}

func newBaseDatatypeT(t DatatypeType) *BaseDatatypeT {
	loge := logrus.New()
	loge.SetFormatter(&logrus.TextFormatter{})
	return &BaseDatatypeT{
		id:     newDatatypeID(),
		opID:   newOperationID(),
		typeOf: t,
		state:  StateLocallyExisted,
		logger: logrus.New(),
	}
}

func executeLocalBase(base *BaseDatatypeT, datatype interface{}, op Operation) (interface{}, error) {
	op.SetOperationID(base.opID.Next())
	return op.executeLocal(datatype)
}

func (c *BaseDatatypeT) String() string {
	return fmt.Sprintf("%s", c.id.String())
}
