package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/internal/types"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	operations "github.com/knowhunger/ortoo/ortoo/operations"
)

// HashMap is an Ortoo datatype which provides hash map interfaces.
type HashMap interface {
	Datatype
	HashMapInTxn
	DoTransaction(tag string, txnFunc func(hashMap HashMapInTxn) error) error
}

// HashMapInTxn is an Ortoo datatype which provides hash map interface in a transaction.
type HashMapInTxn interface {
	Get(key string) interface{}
	Put(key string, value interface{}) (interface{}, error)
	Remove(key string) (interface{}, error)
}

func newHashMap(key string, cuid model.CUID, wire datatypes.Wire, handlers *Handlers) HashMap {
	hashMap := &hashMap{
		datatype: &datatype{
			FinalDatatype: &datatypes.FinalDatatype{},
			handlers:      handlers,
		},
		snapshot: newHashMapSnapshot(),
	}
	hashMap.Initialize(key, model.TypeOfDatatype_HASH_MAP, cuid, wire, hashMap.snapshot, hashMap)
	return hashMap
}

type hashMap struct {
	*datatype
	snapshot *hashMapSnapshot
}

func (its *hashMap) DoTransaction(tag string, txnFunc func(hm HashMapInTxn) error) error {
	return its.FinalDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &hashMap{
			datatype: &datatype{
				FinalDatatype: &datatypes.FinalDatatype{
					TransactionDatatype: its.FinalDatatype.TransactionDatatype,
					TransactionCtx:      txnCtx,
				},
				handlers: its.handlers,
			},
			snapshot: its.snapshot,
		}
		return txnFunc(clone)
	})
}

func (its *hashMap) ExecuteLocal(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.PutOperation:
		return its.snapshot.putCommon(cast.C.Key, cast.C.Value, cast.ID.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot.removeCommon(cast.C.Key, cast.ID.GetTimestamp()), nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *hashMap) ExecuteRemote(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		var newSnap = newHashMapSnapshot()
		if err := json.Unmarshal([]byte(cast.C.Snapshot), newSnap); err != nil {
			return nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
		}
		its.snapshot = newSnap
		return nil, nil
	case *operations.PutOperation:
		return its.snapshot.putCommon(cast.C.Key, cast.C.Value, cast.ID.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot.removeCommon(cast.C.Key, cast.ID.GetTimestamp()), nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *hashMap) GetSnapshot() model.Snapshot {
	return its.snapshot
}

func (its *hashMap) SetSnapshot(snapshot model.Snapshot) {
	its.snapshot = snapshot.(*hashMapSnapshot)
}

func (its *hashMap) GetAsJSON() (string, error) {
	return its.snapshot.GetAsJSON()
}

func (its *hashMap) GetMetaAndSnapshot() ([]byte, model.Snapshot, error) {
	meta, err := its.FinalDatatype.GetMeta()
	if err != nil {
		return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return meta, its.snapshot, nil
}

func (its *hashMap) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	if err := its.FinalDatatype.SetMeta(meta); err != nil {
		return errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	if err := json.Unmarshal([]byte(snapshot), its.snapshot); err != nil {
		return errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return nil
}

func (its *hashMap) Put(key string, value interface{}) (interface{}, error) {
	if key == "" || value == nil {
		return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "empty key or nil value is not allowed")
	}
	jsonSupportedType := types.ConvertToJSONSupportedType(value)

	op := operations.NewPutOperation(key, jsonSupportedType)
	return its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
}

func (its *hashMap) Get(key string) interface{} {
	if obj, ok := its.snapshot.Map[key]; ok {
		return obj.V
	}
	return nil
}

func (its *hashMap) Remove(key string) (interface{}, error) {
	if key == "" {
		return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "empty key is not allowed")
	}
	op := operations.NewRemoveOperation(key)
	return its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
}

// ////////////////////////////////////////////////////////////////
//  hashMapSnapshot
// ////////////////////////////////////////////////////////////////

type obj struct {
	V types.JSONType
	T *model.Timestamp
}

func (its *obj) String() string {
	return fmt.Sprintf("{V:%v,T:%s}", its.V, its.T.ToString())
}

type hashMapSnapshot struct {
	Map map[string]*obj
}

func newHashMapSnapshot() *hashMapSnapshot {
	return &hashMapSnapshot{
		Map: make(map[string]*obj),
	}
}

func (its *hashMapSnapshot) CloneSnapshot() model.Snapshot {
	var cloneMap = make(map[string]*obj)
	for k, v := range its.Map {
		cloneMap[k] = v
	}
	return &hashMapSnapshot{
		Map: cloneMap,
	}
}

func (its *hashMapSnapshot) get(key string) interface{} {
	return its.Map[key]
}

func (its *hashMapSnapshot) putCommon(key string, value interface{}, ts *model.Timestamp) (interface{}, error) {
	oldObj, ok := its.Map[key]
	defer func() {
		log.Logger.Infof("putCommon value: %v => %v for key: %s", oldObj, its.Map[key], key)
	}()
	if !ok {
		its.Map[key] = &obj{
			V: value,
			T: ts,
		}
		return nil, nil
	}

	if oldObj.T.Compare(ts) <= 0 {
		its.Map[key] = &obj{
			V: value,
			T: ts,
		}
	}

	return oldObj.V, nil
}

func (its *hashMapSnapshot) GetAsJSON() (string, error) {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v.V != nil {
			m[k] = v.V
		}
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "", errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return string(data), nil

}

func (its *hashMapSnapshot) removeCommon(key string, ts *model.Timestamp) interface{} {
	if oldObj, ok := its.Map[key]; ok {
		if oldObj.T.Compare(ts) <= 0 {
			its.Map[key] = &obj{
				V: nil,
				T: ts,
			}
			return oldObj.V
		}
	}
	return nil
}
