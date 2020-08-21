package ortoo

import (
	"encoding/json"
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

func (its *hashMap) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.PutOperation:
		return its.snapshot.putCommon(cast.C.Key, cast.C.Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot.removeLocal(cast.C.Key, cast.GetTimestamp())
	}
	return nil, errors.ErrDatatypeIllegalOperation.New(op)
}

func (its *hashMap) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		var newSnap = newHashMapSnapshot()
		if err := json.Unmarshal([]byte(cast.C.Snapshot), newSnap); err != nil {
			return nil, errors.ErrDatatypeSnapshot.New(err.Error())
		}
		its.snapshot = newSnap
		return nil, nil
	case *operations.PutOperation:
		return its.snapshot.putCommon(cast.C.Key, cast.C.Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot.removeRemote(cast.C.Key, cast.GetTimestamp()), nil
	}
	return nil, errors.ErrDatatypeIllegalOperation.New(op)
}

func (its *hashMap) GetSnapshot() iface.Snapshot {
	return its.snapshot
}

func (its *hashMap) SetSnapshot(snapshot iface.Snapshot) {
	its.snapshot = snapshot.(*hashMapSnapshot)
}

func (its *hashMap) GetAsJSON() interface{} {
	return its.snapshot.GetAsJSONCompatible()
}

func (its *hashMap) GetMetaAndSnapshot() ([]byte, iface.Snapshot, error) {
	meta, err := its.ManageableDatatype.GetMeta()
	if err != nil {
		return nil, nil, errors.ErrDatatypeSnapshot.New(err.Error())
	}
	return meta, its.snapshot, nil
}

func (its *hashMap) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	if err := its.ManageableDatatype.SetMeta(meta); err != nil {
		return errors.ErrDatatypeSnapshot.New(err.Error())
	}

	if err := its.snapshot.UnmarshalJSON([]byte(snapshot)); err != nil {
		return errors.ErrDatatypeSnapshot.New(err.Error())
	}

	if err := its.snapshot.UnmarshalJSON([]byte(snapshot)); err != nil {

	}
	return nil
}

func (its *hashMap) Put(key string, value interface{}) (interface{}, error) {
	if key == "" || value == nil {
		return nil, errors.ErrDatatypeIllegalOperation.New("empty key or nil value is not allowed")
	}
	jsonSupportedType := types.ConvertToJSONSupportedValue(value)

	op := operations.NewPutOperation(key, jsonSupportedType)
	return its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
}

func (its *hashMap) Get(key string) interface{} {
	if obj, ok := its.snapshot.Map[key]; ok {
		return obj.getValue()
	}
	return nil
}

func (its *hashMap) Remove(key string) (interface{}, error) {
	if key == "" {
		return nil, errors.ErrDatatypeIllegalOperation.New("empty key is not allowed")
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

type hashMapSnapshot struct {
	Map  map[string]timedType `json:"map"`
	Size int                  `json:"size"`
}

func (its *hashMapSnapshot) UnmarshalJSON(bytes []byte) error {
	var temp = struct {
		Map  map[string]*timedNode `json:"map"`
		Size int                   `json:"size"`
	}{}
	err := json.Unmarshal(bytes, &temp)
	if err != nil {
		return log.OrtooError(err)
	}
	its.Map = make(map[string]timedType)
	for k, v := range temp.Map {
		its.Map[k] = v
	}
	its.Size = temp.Size
	return nil
}

func newHashMapSnapshot() *hashMapSnapshot {
	return &hashMapSnapshot{
		Map:  make(map[string]timedType),
		Size: 0,
	}
}

func (its *hashMapSnapshot) CloneSnapshot() iface.Snapshot {
	var cloneMap = make(map[string]timedType)
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

func (its *hashMapSnapshot) putCommon(key string, value interface{}, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	removed, _ := its.putCommonWithTimedValue(key, &timedNode{
		V: value,
		T: ts,
	})
	if removed != nil {
		return removed.getValue(), nil
	}
	return nil, nil
}

func (its *hashMapSnapshot) putCommonWithTimedValue(key string, tv timedType) (removed timedType, put timedType) {
	removed, ok := its.Map[key]
	if !ok { // empty
		its.Map[key] = tv
		its.Size++
		return nil, tv
	}

	if removed.getTime().Compare(tv.getTime()) <= 0 {
		its.Map[key] = tv
		return removed, tv
	}
	return tv, removed
}

func (its *hashMapSnapshot) GetAsJSONCompatible() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v.getValue() != nil {
			m[k] = v.getValue()
		}
	}
	return m
}

func (its *hashMapSnapshot) removeLocal(key string, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	if tv, ok := its.Map[key]; ok {
		if !tv.isTomb() {
			oldVal := tv.getValue()
			if tv.makeTomb(ts) {
				its.Size--
				return oldVal, nil
			}
		}
	}
	return nil, errors.ErrDatatypeNoOp.New()
}

func (its *hashMapSnapshot) removeRemote(key string, ts *model.Timestamp) interface{} {
	if tv, ok := its.Map[key]; ok {
		oldVal := tv.getValue()
		if tv.makeTomb(ts) {
			its.Size--
			return oldVal
		}
		return nil
	}
	log.Logger.Errorf("No key '%s' exists", key)
	return nil
}

func (its *hashMapSnapshot) size() int {
	return its.Size
}

func (its *hashMapSnapshot) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	for k, v := range its.Map {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(v.String())
		sb.WriteString(" ")
	}
	sb.WriteString("]")
	return sb.String()
}
