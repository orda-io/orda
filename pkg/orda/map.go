package orda

import (
	"encoding/json"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/iface"
	"github.com/orda-io/orda/pkg/internal/datatypes"
	"github.com/orda-io/orda/pkg/model"
	operations "github.com/orda-io/orda/pkg/operations"
	"github.com/orda-io/orda/pkg/types"
	"strings"
)

// Map is an Orda datatype which provides the hash map interfaces.
type Map interface {
	Datatype
	MapInTx
	Transaction(tag string, txFunc func(hashMap MapInTx) error) error
}

// MapInTx is an Orda datatype which provides hash map interface in a transaction.
type MapInTx interface {
	Get(key string) interface{}
	Put(key string, value interface{}) (interface{}, errors.OrdaError)
	Remove(key string) (interface{}, errors.OrdaError)
	Size() int
}

type ordaMap struct {
	*datatype
	*datatypes.SnapshotDatatype
}

func newMap(base *datatypes.BaseDatatype, wire iface.Wire, handlers *Handlers) (Map, errors.OrdaError) {
	oMap := &ordaMap{
		datatype:         newDatatype(base, wire, handlers),
		SnapshotDatatype: datatypes.NewSnapshotDatatype(base, nil),
	}
	return oMap, oMap.init(oMap)
}

func (its *ordaMap) Transaction(tag string, txFunc func(hm MapInTx) error) error {
	return its.DoTransaction(tag, its.TxCtx, func(txCtx *datatypes.TransactionContext) error {
		clone := &ordaMap{
			datatype:         its.cloneDatatype(txCtx),
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return txFunc(clone)
	})
}

func (its *ordaMap) ResetSnapshot() {
	its.Snapshot = newMapSnapshot(its.BaseDatatype)
}

func (its *ordaMap) snapshot() *mapSnapshot {
	return its.GetSnapshot().(*mapSnapshot)
}

func (its *ordaMap) ExecuteLocal(op interface{}) (interface{}, errors.OrdaError) {
	switch cast := op.(type) {
	case *operations.PutOperation:
		return its.snapshot().putCommon(cast.GetBody().Key, cast.GetBody().Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot().removeLocal(cast.GetBody().Key, cast.GetTimestamp())
	}
	return nil, errors.DatatypeIllegalOperation.New(its.L(), its.TypeOf.String(), op)
}

func (its *ordaMap) ExecuteRemote(op interface{}) (interface{}, errors.OrdaError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		return nil, its.ApplySnapshot(cast.GetBody())
	case *operations.PutOperation:
		return its.snapshot().putCommon(cast.GetBody().Key, cast.GetBody().Value, cast.GetTimestamp())
	case *operations.RemoveOperation:
		return its.snapshot().removeRemote(cast.GetBody().Key, cast.GetTimestamp())
	}
	return nil, errors.DatatypeIllegalOperation.New(its.L(), its.TypeOf.String(), op)
}

func (its *ordaMap) Put(key string, value interface{}) (interface{}, errors.OrdaError) {
	if key == "" || value == nil {
		return nil, errors.DatatypeIllegalParameters.New(its.L(), "neither empty key nor null value is not allowed")
	}
	jsonSupportedType := types.ConvertToJSONSupportedValue(value)

	op := operations.NewPutOperation(key, jsonSupportedType)
	return its.SentenceInTx(its.TxCtx, op, true)
}

func (its *ordaMap) Get(key string) interface{} {
	return its.snapshot().get(key)
}

func (its *ordaMap) Remove(key string) (interface{}, errors.OrdaError) {
	if key == "" {
		return nil, errors.DatatypeIllegalParameters.New(its.L(), "empty key is not allowed")
	}
	op := operations.NewRemoveOperation(key)
	return its.SentenceInTx(its.TxCtx, op, true)
}

func (its *ordaMap) Size() int {
	return its.snapshot().size()
}

// ////////////////////////////////////////////////////////////////
//  mapSnapshot
// ////////////////////////////////////////////////////////////////

type mapSnapshot struct {
	iface.BaseDatatype
	Map  map[string]timedType
	Size int
}

func newMapSnapshot(base iface.BaseDatatype) *mapSnapshot {
	return &mapSnapshot{
		BaseDatatype: base,
		Map:          make(map[string]timedType),
		Size:         0,
	}
}

func (its *mapSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Map  map[string]timedType
		Size int
	}{
		Map:  its.Map,
		Size: its.Size,
	})
}

func (its *mapSnapshot) UnmarshalJSON(bytes []byte) error {
	temp := &struct {
		Map  map[string]*timedNode
		Size int
	}{}
	err := json.Unmarshal(bytes, temp)
	if err != nil {
		return errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	its.Map = make(map[string]timedType)
	for k, v := range temp.Map {
		its.Map[k] = v
	}
	its.Size = temp.Size
	return nil
}

func (its *mapSnapshot) getFromMap(key string) timedType {
	return its.Map[key]
}

func (its *mapSnapshot) putCommon(
	key string,
	value interface{},
	ts *model.Timestamp,
) (interface{}, errors.OrdaError) {
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

func (its *mapSnapshot) ToJSON() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v.getValue() != nil {
			m[k] = v.getValue()
		}
	}
	return m
}

func (its *mapSnapshot) removeLocal(key string, ts *model.Timestamp) (interface{}, errors.OrdaError) {
	_, oldV, err := its.removeLocalWithTimedType(key, ts)
	return oldV, err
}

func (its *mapSnapshot) removeRemote(key string, ts *model.Timestamp) (interface{}, errors.OrdaError) {
	_, oldV, err := its.removeRemoteWithTimedType(key, ts)
	return oldV, err
}

func (its *mapSnapshot) removeLocalWithTimedType(
	key string,
	ts *model.Timestamp,
) (timedType, types.JSONValue, errors.OrdaError) {
	if tt, ok := its.Map[key]; ok && !tt.isTomb() {
		oldV := tt.getValue()
		if tt.getTime().Compare(ts) < 0 {
			tt.makeTomb(ts) // makeTomb works differently
			its.Size--
			return tt, oldV, nil
		}
		// local remove cannot reach here; ts is always the newest;
	}
	return nil, nil, errors.DatatypeNoOp.New(its.L(), "remove the value for not existing key")
}

func (its *mapSnapshot) removeRemoteWithTimedType(
	key string,
	ts *model.Timestamp,
) (timedType, types.JSONValue, errors.OrdaError) {
	if tt, ok := its.Map[key]; ok {
		if tt.getTime().Compare(ts) < 0 {
			if !tt.isTomb() {
				its.Size--
			}
			oldV := tt.getValue()
			tt.makeTomb(ts)
			return tt, oldV, nil
		}
		return nil, nil, nil
	}
	return nil, nil, errors.DatatypeNoTarget.New(its.L(), key)
}

func (its *mapSnapshot) size() int {
	return its.Size
}

func (its *mapSnapshot) get(key string) interface{} {
	if tt, ok := its.Map[key]; ok && !tt.isTomb() {
		return tt.getValue()
	}
	return nil
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
