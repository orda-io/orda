package types

import "github.com/knowhunger/ortoo/ortoo/model"

// Datatype defines the interface of executing operations, which is implemented by every datatype.
type Datatype interface {
	GetType() model.TypeOfDatatype        // @baseDatatype
	GetDatatype() Datatype                // @baseDatatype
	GetKey() string                       // @baseDatatype
	GetDUID() DUID                        // @baseDatatype
	SetState(state model.StateOfDatatype) // @baseDatatype
	GetCUID() string                      // @baseDatatype

	// Rollback() error // @TransactionDatatype

	SubscribeOrCreate(state model.StateOfDatatype) error                                             // @FinalDatatype
	ExecuteTransactionRemote(transaction []*model.Operation, obtainList bool) ([]interface{}, error) // @FinalDatatype

	CreatePushPullPack() *model.PushPullPack // @WiredDatatype
	ApplyPushPullPack(*model.PushPullPack)   // @WiredDatatype
	NeedSync(sseq uint64) bool               // @WiredDatatype

	ExecuteLocal(op interface{}) (interface{}, error)      // @Real datatype
	ExecuteRemote(op interface{}) (interface{}, error)     // @Real datatype
	GetSnapshot() Snapshot                                 // @Real datatype
	GetMetaAndSnapshot() ([]byte, Snapshot, error)         // @Real datatype
	SetMetaAndSnapshot(meta []byte, snapshot string) error // @Real datatype

	HandleStateChange(old, new model.StateOfDatatype) // @ortoo.Datatype
	HandleErrors(errs ...error)                       // @ortoo.Datatype
	HandleRemoteOperations(operations []interface{})  // @ortoo.Datatype
}
