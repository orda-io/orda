package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
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
	Put(key string, value interface{}) (interface{}, errors.OrtooError)
	Remove(key string) (interface{}, errors.OrtooError)
	Size() int
}

func newHashMap(base *datatypes.BaseDatatype, wire iface.Wire, handlers *Handlers) (HashMap, errors.OrtooError) {
	hashMap := &hashMap{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		SnapshotDatatype: &datatypes.SnapshotDatatype{
			Snapshot: newHashMapSnapshot(base),
		},
	}
	return hashMap, hashMap.Initialize(base, wire, hashMap.GetSnapshot(), hashMap)
}

type hashMap struct {
	*datatype
	*datatypes.SnapshotDatatype
}

func (its *hashMap) DoTransaction(tag string, txnFunc func(hm HashMapInTxn) error) error {
	return its.ManageableDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &hashMap{
			datatype:         its.newDatatype(txnCtx),
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return txnFunc(clone)
	})
}

func (its *hashMap) ResetSnapshot() {
	its.SnapshotDatatype.SetSnapshot(newHashMapSnapshot(its.BaseDatatype))
}

func (its *hashMap) snapshot() *hashMapSnapshot {
	return its.GetSnapshot().(*hashMapSnapshot)
}

func (its *hashMap) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.PutOperation:
		return its.snapshot().putCommon(cast.C.Key, cast.C.Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot().removeLocal(cast.C.Key, cast.GetTimestamp())
	}
	return nil, errors.DatatypeIllegalParameters.New(its.Logger, op)
}

func (its *hashMap) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		err := its.ApplySnapshotOperation(cast.GetContent(), newHashMapSnapshot(its.BaseDatatype))
		return nil, err
	case *operations.PutOperation:
		return its.snapshot().putCommon(cast.C.Key, cast.C.Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot().removeRemote(cast.C.Key, cast.GetTimestamp())
	}
	return nil, errors.DatatypeIllegalParameters.New(its.Logger, op)
}

func (its *hashMap) Put(key string, value interface{}) (interface{}, errors.OrtooError) {
	if key == "" || value == nil {
		return nil, errors.DatatypeIllegalParameters.New(its.Logger, "empty key or nil value is not allowed")
	}
	jsonSupportedType := types.ConvertToJSONSupportedValue(value)

	op := operations.NewPutOperation(key, jsonSupportedType)
	return its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
}

func (its *hashMap) Get(key string) interface{} {
	if obj, ok := its.snapshot().Map[key]; ok {
		return obj.getValue()
	}
	return nil
}

func (its *hashMap) Remove(key string) (interface{}, errors.OrtooError) {
	if key == "" {
		return nil, errors.DatatypeIllegalParameters.New(its.Logger, "empty key is not allowed")
	}
	op := operations.NewRemoveOperation(key)
	return its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
}

func (its *hashMap) Size() int {
	return its.snapshot().size()
}

// ////////////////////////////////////////////////////////////////
//  hashMapSnapshot
// ////////////////////////////////////////////////////////////////

type hashMapSnapshot struct {
	base
	Map  map[string]timedType `json:"map"`
	Size int                  `json:"size"`
}

func newHashMapSnapshot(base iface.BaseDatatype) *hashMapSnapshot {
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
		return errors.DatatypeMarshal.New(its.GetLogger(), err.Error())
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

func (its *hashMapSnapshot) GetBase() iface.BaseDatatype {
	return its.base
}

func (its *hashMapSnapshot) SetBase(base iface.BaseDatatype) {
	its.base = base
}

func (its *hashMapSnapshot) getFromMap(key string) timedType {
	return its.Map[key]
}

func (its *hashMapSnapshot) putCommon(key string, value interface{}, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	removed, _ := its.putCommonWithTimedType(key, newTimedNode(value, ts))
	if removed != nil {
		return removed.getValue(), nil
	}
	return nil, nil
}

func (its *hashMapSnapshot) putCommonWithTimedType(key string, newOne timedType) (o timedType, n timedType) {
	oldOne, ok := its.Map[key]
	if !ok { // empty
		its.Map[key] = newOne
		its.Size++
		return nil, newOne
	}

	if oldOne.getTime().Compare(newOne.getTime()) < 0 {
		its.Map[key] = newOne
		return oldOne, newOne
	}
	return newOne, oldOne
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

func (its *hashMapSnapshot) removeLocalWithTimedType(
	key string,
	ts *model.Timestamp,
) (timedType, types.JSONValue, errors.OrtooError) {
	if tt, ok := its.Map[key]; ok {
		if !tt.isTomb() {
			oldV := tt.getValue()
			if tt.getTime().Compare(ts) < 0 {
				tt.makeTomb(ts) // makeTomb works differently
				its.Size--
				return tt, oldV, nil
			}
		}
	}
	return nil, nil, errors.DatatypeNoOp.New(its.GetLogger(), "remove the value for not existing key")
}

func (its *hashMapSnapshot) removeRemoteWithTimedType(
	key string,
	ts *model.Timestamp,
) (timedType, types.JSONValue, errors.OrtooError) {
	if tt, ok := its.Map[key]; ok {
		oldV := tt.getValue()
		if tt.getTime().Compare(ts) < 0 {
			if !tt.isTomb() {
				its.Size--
			}
			tt.makeTomb(ts)
			return tt, oldV, nil
		}
		return nil, nil, nil
	}
	return nil, nil, errors.DatatypeNoTarget.New(its.GetLogger(), key)
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
