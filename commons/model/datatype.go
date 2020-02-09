package model

// FinalDatatype defines the interface of executing operations, which is implemented by every datatype.
type FinalDatatype interface {
	GetType() TypeOfDatatype         // @baseDatatype
	GetFinalDatatype() FinalDatatype // @baseDatatype
	GetKey() string                  // @baseDatatype
	GetDUID() DUID                   // @baseDatatype
	SetState(state StateOfDatatype)  // @baseDatatype
	GetCUID() string                 // @baseDatatype

	Rollback() error // @TransactionDatatype

	SubscribeOrCreate(state StateOfDatatype) error          // @CommonDatatype
	ExecuteTransactionRemote(transaction []Operation) error // @CommonDatatype

	CreatePushPullPack() *PushPullPack // @WiredDatatype
	ApplyPushPullPack(*PushPullPack)   // @WiredDatatype
	NeedSync(sseq uint64) bool         // @WiredDatatype

	ExecuteLocal(op interface{}) (interface{}, error)      // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, error)     // @Real datatype
	GetSnapshot() Snapshot                                 // @Real datatype
	GetMetaAndSnapshot() ([]byte, string, error)           // @Real datatype
	SetMetaAndSnapshot(meta []byte, snapshot string) error // @Real datatype
	HandleSubscription()                                   // @Real datatype
	HandleError(err error)                                 // @Real datatype
	HandleRemoteChange()                                   // @Real datatype
}
