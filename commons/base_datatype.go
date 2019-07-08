package commons

type BaseDataType struct {
	id     *datatypeID
	opID   *operationID
	typeOf DatatypeType
	state  DatatypeState
}

type BaseDataTyper interface {
	execute(op *operation)
}

func (c *BaseDataType) execute(op operationer) {
	op.executeLocal()
}

type TransferableDataType struct {
	BaseDataType
	Checkpoint CheckPoint
}
