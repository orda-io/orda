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

// Map is an Ortoo datatype which provides the hash map interfaces.
type Map interface {
	Datatype
	MapInTxn
	DoTransaction(tag string, txFunc func(hashMap MapInTxn) error) error
}

// MapInTxn is an Ortoo datatype which provides hash map interface in a transaction.
type MapInTxn interface {
	Get(key string) interface{}
	Put(key string, value interface{}) (interface{}, errors.OrtooError)
	Remove(key string) (interface{}, errors.OrtooError)
	Size() int
}

func newMap(base *datatypes.BaseDatatype, wire iface.Wire, handlers *Handlers) (Map, errors.OrtooError) {
	oMap := &ortooMap{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		SnapshotDatatype: &datatypes.SnapshotDatatype{
			Snapshot: newMapSnapshot(base),
		},
	}
	return oMap, oMap.Initialize(base, wire, oMap.GetSnapshot(), oMap)
}

type ortooMap struct {
	*datatype
	*datatypes.SnapshotDatatype
}

func (its *ortooMap) DoTransaction(tag string, txFunc func(hm MapInTxn) error) error {
	return its.ManageableDatatype.DoTransaction(tag, func(txCtx *datatypes.TransactionContext) error {
		clone := &ortooMap{
			datatype:         its.newDatatype(txCtx),
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return txFunc(clone)
	})
}

func (its *ortooMap) ResetSnapshot() {
	its.SnapshotDatatype.SetSnapshot(newMapSnapshot(its.BaseDatatype))
}

func (its *ortooMap) snapshot() *mapSnapshot {
	return its.GetSnapshot().(*mapSnapshot)
}

func (its *ortooMap) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.PutOperation:
		return its.snapshot().putCommon(cast.C.Key, cast.C.Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot().removeLocal(cast.C.Key, cast.GetTimestamp())
	}
	return nil, errors.DatatypeIllegalParameters.New(its.Logger, op)
}

func (its *ortooMap) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		err := its.ApplySnapshotOperation(cast.GetContent(), newMapSnapshot(its.BaseDatatype))
		return nil, err
	case *operations.PutOperation:
		return its.snapshot().putCommon(cast.C.Key, cast.C.Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot().removeRemote(cast.C.Key, cast.GetTimestamp())
	}
	return nil, errors.DatatypeIllegalParameters.New(its.Logger, op)
}

func (its *ortooMap) Put(key string, value interface{}) (interface{}, errors.OrtooError) {
	if key == "" || value == nil {
		return nil, errors.DatatypeIllegalParameters.New(its.Logger, "empty key or nil value is not allowed")
	}
	jsonSupportedType := types.ConvertToJSONSupportedValue(value)

	op := operations.NewPutOperation(key, jsonSupportedType)
	return its.SentenceInTransaction(its.TransactionCtx, op, true)
}

func (its *ortooMap) Get(key string) interface{} {
	if obj, ok := its.snapshot().Map[key]; ok {
		return obj.getValue()
	}
	return nil
}

func (its *ortooMap) Remove(key string) (interface{}, errors.OrtooError) {
	if key == "" {
		return nil, errors.DatatypeIllegalParameters.New(its.Logger, "empty key is not allowed")
	}
	op := operations.NewRemoveOperation(key)
	return its.SentenceInTransaction(its.TransactionCtx, op, true)
}

func (its *ortooMap) Size() int {
	return its.snapshot().size()
}

// ////////////////////////////////////////////////////////////////
//  mapSnapshot
// ////////////////////////////////////////////////////////////////

type mapSnapshot struct {
	base
	Map  map[string]timedType `json:"map"`
	Size int                  `json:"size"`
}

func newMapSnapshot(base iface.BaseDatatype) *mapSnapshot {
	return &mapSnapshot{
		base: base,
		Map:  make(map[string]timedType),
		Size: 0,
	}
}

func (its *mapSnapshot) UnmarshalJSON(bytes []byte) error {
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

func (its *mapSnapshot) CloneSnapshot() iface.Snapshot {
	var cloneMap = make(map[string]timedType)
	for k, v := range its.Map {
		cloneMap[k] = v
	}
	return &mapSnapshot{
		Map: cloneMap,
	}
}

func (its *mapSnapshot) GetBase() iface.BaseDatatype {
	return its.base
}

func (its *mapSnapshot) SetBase(base iface.BaseDatatype) {
	its.base = base
}

func (its *mapSnapshot) getFromMap(key string) timedType {
	return its.Map[key]
}

func (its *mapSnapshot) putCommon(key string, value interface{}, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	removed, _ := its.putCommonWithTimedType(key, newTimedNode(value, ts))
	if removed != nil {
		return removed.getValue(), nil
	}
	return nil, nil
}

func (its *mapSnapshot) putCommonWithTimedType(key string, newOne timedType) (o timedType, n timedType) {
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

func (its *mapSnapshot) GetAsJSONCompatible() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v.getValue() != nil {
			m[k] = v.getValue()
		}
	}
	return m
}

func (its *mapSnapshot) removeLocal(key string, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	_, oldV, err := its.removeLocalWithTimedType(key, ts)
	return oldV, err
}

func (its *mapSnapshot) removeRemote(key string, ts *model.Timestamp) (interface{}, errors.OrtooError) {
	_, oldV, err := its.removeRemoteWithTimedType(key, ts)
	return oldV, err
}

func (its *mapSnapshot) removeLocalWithTimedType(
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

func (its *mapSnapshot) removeRemoteWithTimedType(
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

func (its *mapSnapshot) size() int {
	return its.Size
}

func (its *mapSnapshot) String() string {
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
