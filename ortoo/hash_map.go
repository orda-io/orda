package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	operations "github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
	"strings"
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
	Size() int
}

func newHashMap(key string, cuid types.CUID, wire iface.Wire, handlers *Handlers) HashMap {
	hashMap := &hashMap{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
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
	return its.ManageableDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &hashMap{
			datatype: &datatype{
				ManageableDatatype: &datatypes.ManageableDatatype{
					TransactionDatatype: its.ManageableDatatype.TransactionDatatype,
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

func (its *hashMap) GetSnapshot() iface.Snapshot {
	return its.snapshot
}

func (its *hashMap) SetSnapshot(snapshot iface.Snapshot) {
	its.snapshot = snapshot.(*hashMapSnapshot)
}

func (its *hashMap) GetAsJSON() interface{} {
	return its.snapshot.GetAsJSON()
}

func (its *hashMap) GetMetaAndSnapshot() ([]byte, iface.Snapshot, error) {
	meta, err := its.ManageableDatatype.GetMeta()
	if err != nil {
		return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return meta, its.snapshot, nil
}

func (its *hashMap) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	if err := its.ManageableDatatype.SetMeta(meta); err != nil {
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

func (its *hashMap) Size() int {
	return its.snapshot.size()
}

// ////////////////////////////////////////////////////////////////
//  hashMapSnapshot
// ////////////////////////////////////////////////////////////////

type obj struct {
	V types.JSONType
	T *model.Timestamp
}

func (its *obj) String() string {
	if its.V == nil {
		return fmt.Sprintf("Φ|%s", its.T.ToString())
	}
	return fmt.Sprintf("%v|%s}", its.V, its.T.ToString())
}

type hashMapSnapshot struct {
	Map  map[string]*obj
	Size int
}

func newHashMapSnapshot() *hashMapSnapshot {
	return &hashMapSnapshot{
		Map:  make(map[string]*obj),
		Size: 0,
	}
}

func (its *hashMapSnapshot) CloneSnapshot() iface.Snapshot {
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
	if !ok {
		its.Map[key] = &obj{
			V: value,
			T: ts,
		}
		its.Size++
		return nil, nil
	}
	defer func() {
		log.Logger.Infof("putCommon value: %v => %v for key: %s", oldObj, its.Map[key], key)
	}()

	if oldObj.T.Compare(ts) <= 0 {
		its.Map[key] = &obj{
			V: value,
			T: ts,
		}
	}

	return oldObj.V, nil
}

func (its *hashMapSnapshot) GetAsJSON() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v.V != nil {
			m[k] = v.V
		}
	}
	return m

}

func (its *hashMapSnapshot) removeCommon(key string, ts *model.Timestamp) interface{} {
	if oldObj, ok := its.Map[key]; ok {
		if oldObj.T.Compare(ts) <= 0 {
			its.Map[key] = &obj{
				V: nil,
				T: ts,
			}
			if oldObj.V != nil {
				its.Size--
			}
			return oldObj.V
		}
	}
	return nil
}

func (its *hashMapSnapshot) size() int {
	return its.Size
}

func (its *hashMapSnapshot) String() string {
	var sb strings.Builder
	sb.WriteString("{ ")
	for k, v := range its.Map {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(v.String())
		sb.WriteString(" ")
	}
	sb.WriteString("}")
	return sb.String()
}
