package model

//FinalDatatype defines the interface of executing operations, which is implemented by every datatype.
type FinalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, error)
	ExecuteRemote(op interface{}) (interface{}, error)
	Rollback() error
	GetType() TypeOfDatatype
	GetFinalDatatype() FinalDatatype
	GetKey() string
	GetSnapshot() Snapshot
	SubscribeOrCreate() error
	CreatePushPullPack() *PushPullPack
}
