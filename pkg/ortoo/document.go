package ortoo

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
)

// Document is an Ortoo datatype which provides document (JSON-like) interfaces.
type Document interface {
	Datatype
	DocumentInTxn
	DoTransaction(tag string, txFunc func(document DocumentInTxn) error) error
}

// DocumentInTxn is an Ortoo datatype which provides document (JSON-like) interfaces in a transaction.
type DocumentInTxn interface {
	PutToObject(key string, value interface{}) (Document, errors.OrtooError)
	DeleteInObject(key string) (Document, errors.OrtooError)
	GetFromObject(key string) (Document, errors.OrtooError)

	InsertToArray(pos int, value ...interface{}) (Document, errors.OrtooError)
	UpdateManyInArray(pos int, values ...interface{}) ([]Document, errors.OrtooError)
	DeleteInArray(pos int) (Document, errors.OrtooError)
	DeleteManyInArray(pos int, numOfNodes int) ([]Document, errors.OrtooError)
	GetFromArray(pos int) (Document, errors.OrtooError)
	GetManyFromArray(pos int, numOfNodes int) ([]Document, errors.OrtooError)

	GetParentDocument() Document
	GetRootDocument() Document
	GetJSONType() TypeOfJSON
	IsGarbage() bool
	Equal(o Document) bool
}

func newDocument(base *datatypes.BaseDatatype, wire iface.Wire, handlers *Handlers) (Document, errors.OrtooError) {
	doc := &document{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		SnapshotDatatype: &datatypes.SnapshotDatatype{
			Snapshot: newJSONObject(base, nil, model.OldestTimestamp()),
		},
	}
	return doc, doc.Initialize(base, wire, doc.GetSnapshot(), doc)
}

type document struct {
	*datatype
	*datatypes.SnapshotDatatype
}

/*
	document.DoTransaction(userFunc)
		-> ManageableDatatype.DoTransaction(funcWithCloneDatatype())
			-> TransactionDatatype.BeginTransaction()
			-> funcWithCloneDatatype()
				-> userFunc()
					document.Operation ->
						-> TransactionDatatype.SentenceInTransaction()
							-> TransactionDatatype.BeginTransaction(NotUserTransactionTag)
							-> BaseDatatype.executeLocalBase()
								-> op.ExecuteLocal(document)
									-> document.ExecuteLocal(op)
										-> document.snapshot.XXXX
							-> TransactionDatatype.EndTransaction()
*/
func (its *document) DoTransaction(tag string, userFunc func(document DocumentInTxn) error) error {
	return its.ManageableDatatype.DoTransaction(tag, func(txCtx *datatypes.TransactionContext) error {
		clone := &document{
			datatype:         its.newDatatype(txCtx),
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return userFunc(clone)
	})
}

func (its *document) snapshot() jsonType {
	return its.GetSnapshot().(jsonType)
}

func (its *document) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.DocPutInObjectOperation:
		return its.snapshot().PutCommonInObject(cast.C.P, cast.C.K, cast.C.V, cast.GetTimestamp())
	case *operations.DocDeleteInObjectOperation:
		return its.snapshot().DeleteCommonInObject(cast.C.P, cast.C.Key, cast.GetTimestamp(), true)
	case *operations.DocInsertToArrayOperation:
		target, parent, err := its.snapshot().InsertLocalInArray(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V...)
		if err != nil {
			return nil, err
		}
		cast.C.T = target
		return parent, nil
	case *operations.DocDeleteInArrayOperation:
		delTargets, delJSONTypes, err := its.snapshot().DeleteLocalInArray(cast.C.P, cast.Pos, cast.NumOfNodes, cast.ID.GetTimestamp())
		if err != nil {
			return nil, err
		}
		cast.C.T = delTargets
		return delJSONTypes, nil
	case *operations.DocUpdateInArrayOperation:
		uptTargets, oldOnes, err := its.snapshot().UpdateLocalInArray(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V...)
		if err != nil {
			return nil, err
		}
		cast.C.T = uptTargets
		return oldOnes, nil
	}
	return nil, errors.DatatypeIllegalParameters.New(its.L(), op.(iface.Operation).GetType())
}

func (its *document) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		err := its.ApplySnapshotOperation(
			cast.GetContent(),
			newJSONObject(its.GetBase(), nil, model.OldestTimestamp()))
		return nil, err
	case *operations.DocPutInObjectOperation:
		return its.snapshot().PutCommonInObject(cast.C.P, cast.C.K, cast.C.V, cast.GetTimestamp())
	case *operations.DocDeleteInObjectOperation:
		return its.snapshot().DeleteCommonInObject(cast.C.P, cast.C.Key, cast.GetTimestamp(), false)
	case *operations.DocInsertToArrayOperation:
		return its.snapshot().InsertRemoteInArray(cast.C.P, cast.C.T, cast.GetTimestamp(), cast.C.V...)
	case *operations.DocDeleteInArrayOperation:
		return its.snapshot().DeleteRemoteInArray(cast.C.P, cast.GetTimestamp(), cast.C.T)
	case *operations.DocUpdateInArrayOperation:
		return its.snapshot().UpdateRemoteInArray(cast.C.P, cast.GetTimestamp(), cast.C.T, cast.C.V)
	}
	return nil, errors.DatatypeIllegalParameters.New(its.L(), op)
}

// GetFromObject returns the child associated with the given key as a Document.
func (its *document) GetFromObject(key string) (Document, errors.OrtooError) {
	if err := its.assertLocalOp("GetFromObject", TypeJSONObject, true); err != nil {
		return nil, err
	}
	obj := its.snapshot().(*jsonObject)
	child := obj.getFromMap(key).(jsonType)
	if child == nil || child.isGarbage() {
		return nil, nil
	}
	return its.toDocument(child), nil
}

// PutToObject associates a new value with the given key, and returns the old value as a Document
func (its *document) PutToObject(key string, value interface{}) (Document, errors.OrtooError) {
	if err := its.assertLocalOp("PutToObject", TypeJSONObject, false); err != nil {
		return nil, err
	}
	op := operations.NewDocPutInObjectOperation(its.snapshot().getCreateTime(), key, value)
	removed, err := its.SentenceInTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	if removed != nil {
		return its.toDocument(removed.(jsonType)), nil
	}
	return nil, nil
}

// DeleteInObject removes the value associated with the given key, and returns the removed value as a Document.
func (its *document) DeleteInObject(key string) (Document, errors.OrtooError) {
	if err := its.assertLocalOp("DeleteInObject", TypeJSONObject, false); err != nil {
		return nil, err
	}
	op := operations.NewDocDeleteInObjectOperation(its.snapshot().getCreateTime(), key)
	removed, err := its.SentenceInTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocument(removed.(jsonType)), nil
}

// GetFromArray returns the element of the JSONArray Document at the given position.
func (its *document) GetFromArray(pos int) (Document, errors.OrtooError) {
	ret, err := its.GetManyFromArray(pos, 1)
	if err != nil {
		return nil, err
	}
	return ret[0], nil
}

func (its *document) GetManyFromArray(pos int, numOfNodes int) ([]Document, errors.OrtooError) {
	if err := its.assertLocalOp("GetManyFromArray", TypeJSONArray, true); err != nil {
		return nil, err
	}
	arr := its.snapshot().(*jsonArray)
	if err := arr.validateGetRange(pos, numOfNodes); err != nil {
		return nil, err
	}
	children := arr.getManyJSONTypes(pos, numOfNodes)
	return its.toDocuments(children), nil
}

// InsertToArray inserts given values at the next of the given position.
// It returns the current JSONArray Document.
func (its *document) InsertToArray(pos int, values ...interface{}) (Document, errors.OrtooError) {
	if err := its.assertLocalOp("InsertToArray", TypeJSONArray, false); err != nil {
		return its, err
	}
	arr := its.snapshot().(*jsonArray)
	if err := arr.validateInsertPosition(pos); err != nil {
		return its, err
	}
	op := operations.NewDocInsertToArrayOperation(its.snapshot().getCreateTime(), pos, values)
	if _, err := its.SentenceInTransaction(its.TransactionCtx, op, true); err != nil {
		return its, err
	}
	return its, nil
}

// DeleteInArray deletes a value at the given position, and returns the deleted Document.
func (its *document) DeleteInArray(pos int) (Document, errors.OrtooError) {
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
func (its *document) DeleteManyInArray(pos int, numOfNodes int) ([]Document, errors.OrtooError) {
	if err := its.assertLocalOp("DeleteManyInArray", TypeJSONArray, false); err != nil {
		return nil, err
	}
	arr := its.snapshot().(*jsonArray)
	if err := arr.validateGetRange(pos, numOfNodes); err != nil {
		return nil, err
	}
	op := operations.NewDocDeleteInArrayOperation(its.snapshot().getCreateTime(), pos, numOfNodes)
	delJSONTypes, err := its.SentenceInTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocuments(delJSONTypes.([]jsonType)), nil
}

// UpdateManyInArray updates the child from the given position, and returns the previous child Documents
func (its *document) UpdateManyInArray(pos int, values ...interface{}) ([]Document, errors.OrtooError) {
	if err := its.assertLocalOp("UpdateManyInArray", TypeJSONArray, false); err != nil {
		return nil, err
	}
	arr := its.snapshot().(*jsonArray)
	if err := arr.validateGetRange(pos, len(values)); err != nil {
		return nil, err
	}
	op := operations.NewDocUpdateInArrayOperation(its.snapshot().getCreateTime(), pos, values)
	oldOnes, err := its.SentenceInTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocuments(oldOnes.([]jsonType)), nil
}

func (its *document) GetJSONType() TypeOfJSON {
	return its.snapshot().getType()
}

func (its *document) IsGarbage() bool {
	return its.snapshot().isGarbage()
}

func (its *document) GetAsJSON() interface{} {
	return its.snapshot().GetAsJSONCompatible()
}

func (its *document) GetParentDocument() Document {
	return its.toDocument(its.snapshot().getParent())
}
func (its *document) GetRootDocument() Document {
	if its.snapshot().getRoot() == its.snapshot() {
		return its
	}
	return its.toDocument(its.snapshot().getRoot())
}

func (its *document) ResetSnapshot() {
	its.SnapshotDatatype.SetSnapshot(newJSONObject(its.BaseDatatype, nil, model.OldestTimestamp()))
}

func (its *document) Equal(o Document) bool {
	other := o.(*document)
	if its.datatype != other.datatype {
		return false
	}
	if its.snapshot() != other.snapshot() {
		return false
	}
	return true
}

func (its *document) toDocuments(children []jsonType) (ret []Document) {
	for _, child := range children {
		ret = append(ret, its.toDocument(child))
	}
	return
}

func (its *document) toDocument(child jsonType) Document {
	return &document{
		datatype: its.datatype,
		SnapshotDatatype: &datatypes.SnapshotDatatype{
			Snapshot: child,
		},
	}
}
func (its *document) assertLocalOp(opName string, ofJSON TypeOfJSON, workOnGarbage bool) errors.OrtooError {
	if its.GetJSONType() != ofJSON {
		return errors.DatatypeInvalidParent.New(its.L(), opName, " is not allowed to ")
	}
	if !workOnGarbage && its.snapshot().isGarbage() {
		return errors.DatatypeNoOp.New(its.L(), "already deleted from the root Document")
	}
	return nil
}
