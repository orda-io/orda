package commons

type OpType uint8

type operationTypeT struct {
	CreateOpType   OpType
	DeleteOpType   OpType
	ErrorOpType    OpType
	SnapshotOpType OpType
}

var OperationTypes = &operationTypeT{
	CreateOpType:   0,
	DeleteOpType:   1,
	ErrorOpType:    2,
	SnapshotOpType: 3,
}
