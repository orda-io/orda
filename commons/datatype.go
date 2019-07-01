package commons

type BaseDataType struct {
	id     datatypeID
	opID   operationID
	typeOf DatatypeType
	state  DatatypeState
}

func (c *BaseDataType) execute(op operation) {

}

type TransferableDataType struct {
	BaseDataType
	checkpoint checkpoint
}
