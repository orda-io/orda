package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/types"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// HashMap is an Ortoo datatype which provides hash map interfaces.
type HashMap interface {
	datatypes.PublicWiredDatatypeInterface
	HashMapInTxn
	DoTransaction(tag string, txnFunc func(hashMap HashMapInTxn) error) error
}

// HashMapInTxn is an Ortoo datatype which provides hash map interface in a transaction.
type HashMapInTxn interface {
	Get(key string) interface{}
	Put(key string, value interface{}) (interface{}, error)
	Remove(key string) (interface{}, error)
}

func newHashMap(key string, cuid model.CUID, wire datatypes.Wire) (HashMap, error) {
	hashMap := &hashMap{
		FinalDatatype: &datatypes.FinalDatatype{},
		snapshot:      newHashMapSnapshot(),
	}
	if err := hashMap.Initialize(key, model.TypeOfDatatype_HASH_MAP, cuid, wire, hashMap.snapshot, hashMap); err != nil {
		return nil, err
	}
	return hashMap, nil
}

type hashMap struct {
	*datatypes.FinalDatatype
	snapshot *hashMapSnapshot
}

func (h *hashMap) DoTransaction(tag string, txnFunc func(hm HashMapInTxn) error) error {
	return h.FinalDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &hashMap{
			FinalDatatype: &datatypes.FinalDatatype{
				TransactionDatatype: h.FinalDatatype.TransactionDatatype,
				TransactionCtx:      txnCtx,
			},
			snapshot: h.snapshot,
		}
		return txnFunc(clone)
	})
}

func (h *hashMap) ExecuteLocal(op interface{}) (interface{}, error) {
	switch op.(type) {
	case *model.PutOperation:
		put := op.(*model.PutOperation)
		return h.snapshot.putCommon(put.Key, put.Value, put.Base.GetTimestamp())
	case *model.RemoveOperation:
		remove := op.(*model.RemoveOperation)
		return h.snapshot.remove(remove.Key, remove.Base.GetTimestamp()), nil
	}
	return nil, nil
}

func (h *hashMap) ExecuteRemote(op interface{}) (interface{}, error) {
	panic("implement me")
}

func (h *hashMap) GetSnapshot() model.Snapshot {
	return h.snapshot
}

func (h *hashMap) SetSnapshot(snapshot model.Snapshot) {
	h.snapshot = snapshot.(*hashMapSnapshot)
}

func (h *hashMap) GetMetaAndSnapshot() ([]byte, string, error) {
	panic("implement me")
}

func (h *hashMap) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	panic("implement me")
}

func (h *hashMap) HandleStateChange(old, new model.StateOfDatatype) {
	panic("implement me")
}

func (h *hashMap) HandleError(errs []error) {
	panic("implement me")
}

func (h *hashMap) HandleRemoteOperations(operations []interface{}) {
	panic("implement me")
}

func (h *hashMap) Put(key string, value interface{}) (interface{}, error) {
	val, err := model.ConvertType(value)
	if err != nil {
		return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidType, err.Error())
	}
	op := model.NewPutOperation(key, val)
	return h.ExecuteOperationWithTransaction(h.TransactionCtx, op, true)
}

func (h *hashMap) Get(key string) interface{} {
	if obj, ok := h.snapshot.Map[key]; ok {
		return obj.V
	}
	return nil
}

func (h *hashMap) Remove(key string) (interface{}, error) {
	op := model.NewRemoveOperation(key)
	return h.ExecuteOperationWithTransaction(h.TransactionCtx, op, true)
}

// ////////////////////////////////////////////////////////////////
//  HashMapHandlers
// ////////////////////////////////////////////////////////////////

// ////////////////////////////////////////////////////////////////
//  hashMapSnapshot
// ////////////////////////////////////////////////////////////////

type obj struct {
	V model.OrtooType
	T *model.Timestamp
}

func (o *obj) String() string {
	return fmt.Sprintf("{V:%v,T:%s}", o.V, o.T.ToString())
}

type hashMapSnapshot struct {
	Map map[string]*obj
}

func newHashMapSnapshot() *hashMapSnapshot {
	return &hashMapSnapshot{
		Map: make(map[string]*obj),
	}
}

func (h *hashMapSnapshot) CloneSnapshot() model.Snapshot {
	var cloneMap = make(map[string]*obj)
	for k, v := range h.Map {
		cloneMap[k] = v
	}
	return &hashMapSnapshot{
		Map: cloneMap,
	}
}

func (h *hashMapSnapshot) GetTypeAny() (*types.Any, error) {
	bin, err := json.Marshal(h)
	if err != nil {
		return nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	log.Logger.Infof("%s", h)
	return &types.Any{
		TypeUrl: h.GetTypeURL(),
		Value:   bin,
	}, nil
}

func (h *hashMapSnapshot) GetTypeURL() string {
	return "github.com/knowhunger/ortoo/ortoo/hashMapSnapshot"
}

func (h *hashMapSnapshot) get(key string) interface{} {
	return h.Map[key]
}

func (h *hashMapSnapshot) putCommon(key string, value interface{}, ts *model.Timestamp) (interface{}, error) {
	oldObj, ok := h.Map[key]
	defer func() {
		log.Logger.Infof("putCommon value: %v => %v for key: %s", oldObj, h.Map[key], key)
	}()
	if !ok {
		h.Map[key] = &obj{
			V: value,
			T: ts,
		}
		return nil, nil
	}

	if oldObj.T.Compare(ts) <= 0 {
		h.Map[key] = &obj{
			V: value,
			T: ts,
		}
	}

	return oldObj.V, nil
}

func (h *hashMapSnapshot) remove(key string, ts *model.Timestamp) interface{} {
	if oldObj, ok := h.Map[key]; ok {
		if oldObj.T.Compare(ts) <= 0 {
			h.Map[key] = &obj{
				V: nil,
				T: ts,
			}
			return oldObj.V
		}
	}
	return nil
}
