package commons

type OpType uint8

type operationTypeT struct {
	CreateOpType           OpType
	DeleteOpType           OpType
	ErrorOpType            OpType
	SnapshotOpType         OpType
	IntCounterIncreaseType OpType
}

var OperationTypes = &operationTypeT{
	CreateOpType:           1,
	DeleteOpType:           2,
	ErrorOpType:            3,
	SnapshotOpType:         4,
	IntCounterIncreaseType: 11,
}
