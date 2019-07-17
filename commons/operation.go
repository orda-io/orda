package commons

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/protocols"
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

func NewOperation(opType OpType) *operationT {
	return &operationT{
		id:  newOperationID(),
		typ: opType,
	}
}

func (o *operationT) SetOperationID(opID *operationID) {
	o.id = opID
}

func (o *operationT) GetOperationID() *operationID {
	return o.id
}

func (o *operationT) executeLocal(datatype interface{}) (interface{}, error) {
	log.Logger.Infoln("operation executeLocal")
	return nil, fmt.Errorf("not implemented yet")
}

func (o *operationT) executeRemote(datatype interface{}) {
	log.Logger.Infoln("operation executeRemote")
}

func (o *operationT) toProtoBuf() *protocols.PbOperation {
	return &protocols.PbOperation{
		Id:     o.id.toProtoBuf(),
		OpType: uint32(o.typ),
	}
}

func pbToOperation(pb *protocols.PbOperation) (*operationT, error) {
	id, err := pbToOperationID(pb.Id)
	if err != nil {
		return nil,
	}
	return &operationT{
		id:        nil,
		typ:       0,
		timestamp: timestamp{},
	}
}
