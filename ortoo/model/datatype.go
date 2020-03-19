package model

// Datatype defines the interface of executing operations, which is implemented by every datatype.
type Datatype interface {
	GetType() TypeOfDatatype        // @baseDatatype
	GetDatatype() Datatype          // @baseDatatype
	GetKey() string                 // @baseDatatype
	GetDUID() DUID                  // @baseDatatype
	SetState(state StateOfDatatype) // @baseDatatype
	GetCUID() string                // @baseDatatype

	Rollback() error // @TransactionDatatype

	SubscribeOrCreate(state StateOfDatatype) error          // @FinalDatatype
	ExecuteTransactionRemote(transaction []Operation) error // @FinalDatatype

	CreatePushPullPack() *PushPullPack // @WiredDatatype
	ApplyPushPullPack(*PushPullPack)   // @WiredDatatype
	NeedSync(sseq uint64) bool         // @WiredDatatype

	ExecuteLocal(op interface{}) (interface{}, error)        // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, error)       // @Real datatype
	GetSnapshot() Snapshot                                   // @Real datatype
	GetMetaAndSnapshot() ([]byte, Snapshot, error)           // @Real datatype
	SetMetaAndSnapshot(meta []byte, snapshot Snapshot) error // @Real datatype

	HandleStateChange(old, new StateOfDatatype)      // @ortoo.Datatype
	HandleErrors(errs ...error)                      // @ortoo.Datatype
	HandleRemoteOperations(operations []interface{}) // @ortoo.Datatype
}
