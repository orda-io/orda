package ortoo

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
)

type Document interface {
	Datatype
	DocumentInTxn
	DoTransaction(tag string, txnFunc func(document DocumentInTxn) error) error
}

type DocumentInTxn interface {
	PutToObject(key string, value interface{}) (interface{}, error)
	InsertToArray(pos int, value ...interface{}) (interface{}, error)
	DeleteInObject(key string) (interface{}, error)
	DeleteInArray(pos int) (interface{}, error)
	DeleteManyInArray(pos int, numOfNodes int) ([]interface{}, error)
	UpdateInArray(pos int, values ...interface{}) ([]interface{}, error)
	GetFromObject(key string) (Document, error)
	GetFromArray(pos int) (Document, error)
	GetDocumentType() TypeOfJSON
}

func newDocument(key string, cuid types.CUID, wire iface.Wire, handlers *Handlers) Document {
	doc := &document{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		root:      model.OldestTimestamp,
		typeOfDoc: TypeJSONObject,
		snapshot:  newJSONObject(nil, model.OldestTimestamp),
	}
	doc.Initialize(key, model.TypeOfDatatype_DOCUMENT, cuid, wire, doc.snapshot, doc)
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

func (its *document) PutToObject(key string, value interface{}) (interface{}, error) {
	if its.typeOfDoc == TypeJSONObject {
		op := operations.NewDocPutInObjectOperation(its.root, key, value)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New()
}

func (its *document) InsertToArray(pos int, values ...interface{}) (interface{}, error) {
	if its.typeOfDoc == TypeJSONArray {
		op := operations.NewDocInsToArrayOperation(its.root, pos, values)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New()
}

func (its *document) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.DocPutInObjectOperation:
		if _, err := its.snapshot.PutCommonInObject(cast.C.P, cast.C.K, cast.C.V, cast.GetTimestamp()); err != nil {
			return nil, err
		}
		return its, nil
	case *operations.DocDeleteInObjectOperation:
		ret, err := its.snapshot.DeleteLocalInObject(cast.C.P, cast.C.Key, cast.GetTimestamp())
		if err != nil {
			return nil, err
		}
		return ret, nil
	case *operations.DocInsertToArrayOperation:
		target, ret, err := its.snapshot.InsertLocal(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V...)
		if err != nil {
			return nil, err
		}
		cast.C.T = target
		return ret, nil
	case *operations.DocDeleteInArrayOperation:
		deletedTargets, deletedValues, err := its.snapshot.DeleteLocalInArray(cast.C.P, cast.Pos, cast.NumOfNodes, cast.ID.GetTimestamp())
		if err != nil {
			return nil, err
		}
		cast.C.T = deletedTargets
		return deletedValues, nil
	case *operations.DocUpdateInArrayOperation:
		// its.snapshot.UpdateLocalInArray()
		return nil, nil
	}
	return nil, errors.ErrDatatypeIllegalOperation.New(op.(iface.Operation).GetType())
}

func (its *document) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		var newSnap = newJSONObject(nil, model.OldestTimestamp)
		if err := json.Unmarshal([]byte(cast.C.Snapshot), newSnap); err != nil {
			return nil, errors.ErrDatatypeSnapshot.New(err.Error())
		}
		its.snapshot = newSnap
		// its.datatype.SetOpID()
		return nil, nil
	case *operations.DocPutInObjectOperation:
		if _, err := its.snapshot.PutCommonInObject(cast.C.P, cast.C.K, cast.C.V, cast.GetTimestamp()); err != nil {
			return nil, err
		}
		return nil, nil
	case *operations.DocInsertToArrayOperation:
		its.snapshot.InsertRemote(cast.C.P, cast.C.T, cast.GetTimestamp(), cast.C.V...)
		return nil, nil
	case *operations.DocDeleteInObjectOperation:
		if _, err := its.snapshot.DeleteRemoteInObject(cast.C.P, cast.C.Key, cast.GetTimestamp()); err != nil {
			return nil, err
		}
		return nil, nil
	case *operations.DocDeleteInArrayOperation:
		its.snapshot.DeleteRemoteInArray(cast.C.P, cast.C.T, cast.GetTimestamp())
		return nil, nil
	}
	return nil, errors.ErrDatatypeIllegalOperation.New(op)
}

func (its *document) GetFromObject(key string) (Document, error) {
	if its.typeOfDoc == TypeJSONObject {
		if currentRoot, ok := its.snapshot.findJSONObject(its.root); ok {
			child := currentRoot.get(key).(jsonType)
			if child == nil {
				return nil, errors.ErrDatatypeNotExistChildDocument.New()
			}
			return its.getChildDocument(child), nil
		}

	}
	return nil, errors.ErrDatatypeInvalidParent.New()
}

func (its *document) getChildDocument(child jsonType) *document {
	return &document{
		datatype:  its.datatype,
		root:      child.getKey(),
		typeOfDoc: child.getType(),
		snapshot:  its.snapshot,
	}
}

func (its *document) GetFromArray(pos int) (Document, error) {
	if its.typeOfDoc == TypeJSONArray {
		if currentRoot, ok := its.snapshot.findJSONArray(its.root); ok {
			c, err := currentRoot.getPrecededType(pos)
			if err != nil {
				return nil, err
			}
			child := c.(jsonType)
			return its.getChildDocument(child), nil
		}
	}
	return nil, errors.ErrDatatypeInvalidParent.New()
}

func (its *document) DeleteInObject(key string) (interface{}, error) {
	if its.typeOfDoc == TypeJSONObject {
		op := operations.NewDocDeleteInObjectOperation(its.root, key)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New()
}

func (its *document) DeleteInArray(pos int) (interface{}, error) {
	ret, err := its.DeleteManyInArray(pos, 1)
	return ret[0], err
}

func (its *document) DeleteManyInArray(pos int, numOfNodes int) ([]interface{}, error) {
	if its.typeOfDoc == TypeJSONArray {
		op := operations.NewDocDeleteInArrayOperation(its.root, pos, numOfNodes)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret.([]interface{}), nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New()
}

func (its *document) UpdateInArray(pos int, values ...interface{}) ([]interface{}, error) {
	if its.typeOfDoc == TypeJSONArray {
		if len(values) < 1 {
			return nil, errors.ErrDatatypeIllegalOperation.New("at least one value should be inserted")
		}

		op := operations.NewDocUpdateInArrayOperation(its.root, pos, values)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret.([]interface{}), nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New()
}

func (its *document) GetDocumentType() TypeOfJSON {
	return its.typeOfDoc
}

func (its *document) GetAsJSON() interface{} {
	r, _ := its.snapshot.findJSONPrimitive(its.root)
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

func (its *document) GetMetaAndSnapshot() ([]byte, iface.Snapshot, error) {
	meta, err := its.ManageableDatatype.GetMeta()
	if err != nil {
		return nil, nil, errors.ErrDatatypeSnapshot.New(err.Error())
	}
	return meta, its.snapshot, nil
}

func (its *document) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	log.Logger.Infof("SetMetaAndSnapshot:%v", snapshot)
	if err := its.ManageableDatatype.SetMeta(meta); err != nil {
		return errors.ErrDatatypeSnapshot.New(err.Error())
	}
	if err := json.Unmarshal([]byte(snapshot), its.snapshot); err != nil {
		return errors.ErrDatatypeSnapshot.New(err.Error())
	}
	return nil
}
