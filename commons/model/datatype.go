package model

// FinalDatatype defines the interface of executing operations, which is implemented by every datatype.
type FinalDatatype interface {
	ExecuteLocal(op interface{}) (interface{}, error)  // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, error) // @Real datatype
	Rollback() error                                   // @TransactionDatatype
	GetType() TypeOfDatatype                           // @baseDatatype
	GetFinalDatatype() FinalDatatype                   // @baseDatatype
	GetKey() string                                    // @baseDatatype
	GetDUID() DUID                                     // @baseDatatype
	GetSnapshot() Snapshot                             // @Real datatype
	SubscribeOrCreate(state StateOfDatatype) error     // @CommonDatatype
	ExecuteTransactionRemote(transaction []Operation) error
	CreatePushPullPack() *PushPullPack // @WiredDatatype
	ApplyPushPullPack(*PushPullPack)   // @WiredDatatype
	SetState(state StateOfDatatype)    // @baseDatatype
	GetCUID() string                   // @baseDatatype
	GetMetaAndSnapshot() ([]byte, string)
}
