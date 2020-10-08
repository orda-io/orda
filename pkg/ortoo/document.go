package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
	"github.com/knowhunger/ortoo/pkg/types"
)

// Document is an Ortoo datatype which provides document (JSON-like) interfaces.
type Document interface {
	Datatype
	DocumentInTxn
	DoTransaction(tag string, txnFunc func(document DocumentInTxn) error) error
}

// DocumentInTxn is an Ortoo datatype which provides document (JSON-like) interfaces in a transaction.
type DocumentInTxn interface {
	PutToObject(key string, value interface{}) (Document, error)
	DeleteInObject(key string) (Document, error)
	GetFromObject(key string) (Document, error)

	InsertToArray(pos int, value ...interface{}) (Document, error)
	UpdateInArray(pos int, values ...interface{}) ([]Document, error)
	DeleteInArray(pos int) (Document, error)
	DeleteManyInArray(pos int, numOfNodes int) ([]Document, error)
	GetFromArray(pos int) (Document, error)

	GetJSONType() TypeOfJSON
}

func newDocument(key string, cuid types.CUID, wire iface.Wire, handlers *Handlers) Document {
	base := datatypes.NewBaseDatatype(key, model.TypeOfDatatype_DOCUMENT, cuid)
	doc := &document{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		root:      model.OldestTimestamp(),
		typeOfDoc: TypeJSONObject,
		snapshot:  newJSONObject(base, nil, model.OldestTimestamp()),
	}
	doc.Initialize(base, wire, doc.snapshot, doc)
	return doc
}

type document struct {
	*datatype
	root      *model.Timestamp
	typeOfDoc TypeOfJSON
	snapshot  *jsonObject
}

func (its *document) DoTransaction(tag string, txnFunc func(document DocumentInTxn) error) error {
	return its.ManageableDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &document{
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

func (its *document) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.DocPutInObjectOperation:
		return its.snapshot.PutCommonInObject(cast.C.P, cast.C.K, cast.C.V, cast.GetTimestamp())
	case *operations.DocDeleteInObjectOperation:
		return its.snapshot.DeleteLocalInObject(cast.C.P, cast.C.Key, cast.GetTimestamp())
	case *operations.DocInsertToArrayOperation:
		target, parent, err := its.snapshot.InsertLocalInArray(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V...)
		if err != nil {
			return nil, err
		}
		cast.C.T = target
		return parent, nil
	case *operations.DocDeleteInArrayOperation:
		delTargets, delJSONTypes, err := its.snapshot.DeleteLocalInArray(cast.C.P, cast.Pos, cast.NumOfNodes, cast.ID.GetTimestamp())
		if err != nil {
			return nil, err
		}
		cast.C.T = delTargets
		return delJSONTypes, nil
	case *operations.DocUpdateInArrayOperation:
		uptTargets, oldOnes, err := its.snapshot.UpdateLocalInArray(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V)
		if err != nil {
			return nil, err
		}
		cast.C.T = uptTargets
		return oldOnes, nil
	}
	return nil, errors.ErrDatatypeIllegalParameters.New(its.Logger, op.(iface.Operation).GetType())
}

func (its *document) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		var newSnap = newJSONObject(its.BaseDatatype, nil, model.OldestTimestamp())
		if err := json.Unmarshal([]byte(cast.C.Snapshot), newSnap); err != nil {
			return nil, errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
		}
		its.snapshot = newSnap
		return nil, nil
	case *operations.DocPutInObjectOperation:
		return its.snapshot.PutCommonInObject(cast.C.P, cast.C.K, cast.C.V, cast.GetTimestamp())
	case *operations.DocDeleteInObjectOperation:
		return its.snapshot.DeleteRemoteInObject(cast.C.P, cast.C.Key, cast.GetTimestamp())
	case *operations.DocInsertToArrayOperation:
		return its.snapshot.InsertRemoteInArray(cast.C.P, cast.C.T, cast.GetTimestamp(), cast.C.V...)
	case *operations.DocDeleteInArrayOperation:
		return its.snapshot.DeleteRemoteInArray(cast.C.P, cast.GetTimestamp(), cast.C.T)
	case *operations.DocUpdateInArrayOperation:
		return its.snapshot.UpdateRemoteInArray(cast.C.P, cast.GetTimestamp(), cast.C.T, cast.C.V)
	}
	return nil, errors.ErrDatatypeIllegalParameters.New(its.Logger, op)
}

// GetFromObject returns the child associated with the given key as a Document.
func (its *document) GetFromObject(key string) (Document, error) {
	jt, err := its.allowLocalOperation("GetFromObject", TypeJSONObject, true)
	if err != nil {
		return nil, err
	}
	currentRoot := jt.(*jsonObject)
	child := currentRoot.getFromMap(key).(jsonType)
	if child == nil {
		return nil, errors.ErrDatatypeNotExist.New(its.Logger)
	}
	return its.toDocument(child), nil
}

// PutToObject associates a new value with the given key, and returns the old value as a Document
func (its *document) PutToObject(key string, value interface{}) (Document, error) {
	if _, err := its.allowLocalOperation("PutToObject", TypeJSONObject, false); err != nil {
		return nil, err
	}
	op := operations.NewDocPutInObjectOperation(its.root, key, value)
	removed, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	if removed != nil {
		return its.toDocument(removed.(jsonType)), nil
	}
	return nil, nil
}

// DeleteInObject removes the value associated with the given key, and returns the removed value as a Document.
func (its *document) DeleteInObject(key string) (Document, error) {
	if _, err := its.allowLocalOperation("DeleteInObject", TypeJSONObject, false); err != nil {
		return nil, err
	}
	op := operations.NewDocDeleteInObjectOperation(its.root, key)
	removed, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocument(removed.(jsonType)), nil
}

// GetFromArray returns the element of the JSONArray Document at the given position.
func (its *document) GetFromArray(pos int) (Document, error) {
	jt, err := its.allowLocalOperation("GetFromArray", TypeJSONArray, true)
	if err != nil {
		return nil, err
	}
	currentRoot := jt.(*jsonArray)
	c, err := currentRoot.findTimedType(pos)
	if err != nil {
		return nil, err
	}
	child := c.(jsonType)
	return its.toDocument(child), nil
}

// InsertToArray inserts given values at the next of the given position.
// It returns the current JSONArray Document.
func (its *document) InsertToArray(pos int, values ...interface{}) (Document, error) {
	if _, err := its.allowLocalOperation("InsertToArray", TypeJSONArray, false); err != nil {
		return nil, err
	}
	op := operations.NewDocInsertToArrayOperation(its.root, pos, values)
	_, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its, nil
}

// DeleteInArray deletes a value at the given position, and returns the deleted Document.
func (its *document) DeleteInArray(pos int) (Document, error) {
	ret, err := its.DeleteManyInArray(pos, 1)
	if err != nil {
		return nil, err
	}
	if ret != nil {
		return ret[0], err
	}
	return nil, nil
}

// DeleteManyInArray deletes values of the given range, and returns the deleted Documents.
func (its *document) DeleteManyInArray(pos int, numOfNodes int) ([]Document, error) {
	if numOfNodes < 1 {
		return nil, errors.ErrDatatypeIllegalParameters.New(its.Logger, "numOfNodes >= 1 is allowed")
	}
	if _, err := its.allowLocalOperation("DeleteManyInArray", TypeJSONArray, false); err != nil {
		return nil, err
	}
	op := operations.NewDocDeleteInArrayOperation(its.root, pos, numOfNodes)
	delJSONTypes, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocuments(delJSONTypes.([]jsonType)), nil

}

// UpdateInArray updates the child from the given position, and returns the previous child Documents
func (its *document) UpdateInArray(pos int, values ...interface{}) ([]Document, error) {
	if len(values) < 1 {
		return nil, errors.ErrDatatypeIllegalParameters.New(its.Logger, "at least one value is required")
	}
	if _, err := its.allowLocalOperation("UpdateInArray", TypeJSONArray, false); err != nil {
		return nil, err
	}

	op := operations.NewDocUpdateInArrayOperation(its.root, pos, values)
	oldOnes, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocuments(oldOnes.([]jsonType)), nil
}

func (its *document) GetJSONType() TypeOfJSON {
	return its.typeOfDoc
}

func (its *document) GetAsJSON() interface{} {
	r, _ := its.snapshot.findJSONType(its.root)
	switch cast := r.(type) {
	case *jsonObject:
		return cast.GetAsJSONCompatible()
	case *jsonArray:
		return cast.GetAsJSONCompatible()
	case *jsonElement:
		return cast.getValue()
	}
	return nil
}

func (its *document) SetSnapshot(snapshot iface.Snapshot) {
	panic("implement me")
}

func (its *document) GetSnapshot() iface.Snapshot {
	return its.snapshot
}

func (its *document) GetMetaAndSnapshot() ([]byte, iface.Snapshot, errors.OrtooError) {
	meta, err := its.ManageableDatatype.GetMeta()
	if err != nil {
		return nil, nil, errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
	}
	return meta, its.snapshot, nil
}

func (its *document) SetMetaAndSnapshot(meta []byte, snapshot string) errors.OrtooError {
	log.Logger.Infof("SetMetaAndSnapshot:%v", snapshot)
	if err := its.ManageableDatatype.SetMeta(meta); err != nil {
		return errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
	}
	if err := json.Unmarshal([]byte(snapshot), its.snapshot); err != nil {
		return errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
	}
	return nil
}

func (its *document) toDocuments(children []jsonType) (ret []Document) {
	for _, child := range children {
		ret = append(ret, its.toDocument(child))
	}
	return
}

func (its *document) toDocument(child jsonType) Document {
	return &document{
		datatype:  its.datatype,
		root:      child.getCreateTime(),
		typeOfDoc: child.getType(),
		snapshot:  its.snapshot,
	}
}
func (its *document) allowLocalOperation(opName string, ofJSON TypeOfJSON, workOnOrphan bool) (jsonType, errors.OrtooError) {
	if its.GetJSONType() != ofJSON {
		return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, opName, " is not allowed to ")
	}
	jt, ok := its.snapshot.findJSONType(its.root)
	if !ok {
		return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "not exist in the root Document")
	}
	if !workOnOrphan && !jt.isTomb() { // FIXME: should check if there exist tombstones of its ancestor
		return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "already deleted from the root Document")
	}
	return jt, nil
}
