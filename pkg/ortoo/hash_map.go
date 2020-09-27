package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	operations "github.com/knowhunger/ortoo/pkg/operations"
	"github.com/knowhunger/ortoo/pkg/types"
	"strings"
)

// HashMap is an Ortoo datatype which provides the hash map interfaces.
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
	base := datatypes.NewBaseDatatype(key, model.TypeOfDatatype_HASH_MAP, cuid)
	hashMap := &hashMap{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		snapshot: newHashMapSnapshot(base),
	}
	hashMap.Initialize(base, wire, hashMap.snapshot, hashMap)
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
	return nil, errors.ErrDatatypeIllegalOperation.New(its.Logger, op)
}

func (its *hashMap) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		var newSnap = newHashMapSnapshot(its.BaseDatatype)
		if err := json.Unmarshal([]byte(cast.C.Snapshot), newSnap); err != nil {
			return nil, errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
		}
		its.snapshot = newSnap
		return nil, nil
	case *operations.PutOperation:
		return its.snapshot.putCommon(cast.C.Key, cast.C.Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot.removeRemote(cast.C.Key, cast.GetTimestamp())
	}
	return nil, errors.ErrDatatypeIllegalOperation.New(its.Logger, op)
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

func (its *hashMap) GetMetaAndSnapshot() ([]byte, iface.Snapshot, errors.OrtooError) {
	meta, err := its.ManageableDatatype.GetMeta()
	if err != nil {
		return nil, nil, errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
	}
	return meta, its.snapshot, nil
}

func (its *hashMap) SetMetaAndSnapshot(meta []byte, snapshot string) errors.OrtooError {
	if err := its.ManageableDatatype.SetMeta(meta); err != nil {
		return errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
	}

	if err := its.snapshot.UnmarshalJSON([]byte(snapshot)); err != nil {
		return errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
	}

	if err := its.snapshot.UnmarshalJSON([]byte(snapshot)); err != nil {

	}
	return nil
}

func (its *hashMap) Put(key string, value interface{}) (interface{}, error) {
	if key == "" || value == nil {
		return nil, errors.ErrDatatypeIllegalOperation.New(its.Logger, "empty key or nil value is not allowed")
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
		return nil, errors.ErrDatatypeIllegalOperation.New(its.Logger, "empty key is not allowed")
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
	base *datatypes.BaseDatatype
	Map  map[string]timedType `json:"map"`
	Size int                  `json:"size"`
}

func newHashMapSnapshot(base *datatypes.BaseDatatype) *hashMapSnapshot {
	return &hashMapSnapshot{
		base: base,
		Map:  make(map[string]timedType),
		Size: 0,
	}
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
	removed, _ := its.putCommonWithTimedType(key, newTimedNode(value, ts))
	if removed != nil {
		return removed.getValue(), nil
	}
	return nil, nil
}

func (its *hashMapSnapshot) putCommonWithTimedType(key string, tt timedType) (timedType, timedType) {
	removed, ok := its.Map[key]
	if !ok { // empty
		its.Map[key] = tt
		its.Size++
		return nil, tt
	}

	if removed.getTime().Compare(tt.getTime()) <= 0 {
		its.Map[key] = tt
		return removed, tt
	}
	return tt, removed
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
	_, oldV, err := its.removeLocalWithTimedType(key, ts)
	return oldV, err
}

func (its *hashMapSnapshot) removeRemote(key string, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	_, oldV, err := its.removeRemoteWithTimedType(key, ts)
	return oldV, err
}

func (its *hashMapSnapshot) removeLocalWithTimedType(key string, ts *model.Timestamp) (timedType, types.JSONValue, errors.OrtooError) {
	if tt, ok := its.Map[key]; ok {
		if !tt.isTomb() {
			oldV := tt.getValue()
			if tt.makeTomb(ts) { // makeTomb works differently according to implementations of timedType
				its.Size--
				return tt, oldV, nil
			}
		}
	}
	return nil, nil, errors.ErrDatatypeNoOp.New(its.base.Logger)
}

func (its *hashMapSnapshot) removeRemoteWithTimedType(key string, ts *model.Timestamp) (timedType, types.JSONValue, errors.OrtooError) {
	if tt, ok := its.Map[key]; ok {
		oldV := tt.getValue()
		if tt.makeTomb(ts) {
			its.Size--
			return tt, oldV, nil
		}
		return nil, nil, nil
	}
	return nil, nil, errors.ErrDatatypeNoTarget.New(its.base.Logger, key)
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
