package commons

import (
	"fmt"
	. "github.com/knowhunger/ortoo/commons/utils"
)

type Operation interface {
	executeLocal(datatype interface{}) (interface{}, error)
	executeRemote(datatype interface{})
	SetOperationID(opID *operationID)
	GetOperationID() *operationID
}

type operationT struct {
	id        *operationID
	typ       OpType
	timestamp timestamp
}

func (o *operationT) SetOperationID(opID *operationID) {
	o.id = opID
}

func (o *operationT) GetOperationID() *operationID {
	return o.id
}

func (o *operationT) executeLocal(datatype interface{}) (interface{}, error) {
	Log.Println("operation executeLocal")
	return nil, fmt.Errorf("not implemented yet")
}

func (o *operationT) executeRemote(datatype interface{}) {
	Log.Println("operation executeRemote")
}
