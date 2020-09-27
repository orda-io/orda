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
	base := datatypes.NewBaseDatatype(key, model.TypeOfDatatype_DOCUMENT, cuid)
	doc := &document{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		root:      model.OldestTimestamp,
		typeOfDoc: TypeJSONObject,
		snapshot:  newJSONObject(base, nil, model.OldestTimestamp),
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

func (its *document) PutToObject(key string, value interface{}) (interface{}, error) {
	if its.typeOfDoc == TypeJSONObject {
		op := operations.NewDocPutInObjectOperation(its.root, key, value)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "PutToObject not on JSONObject")
}

func (its *document) InsertToArray(pos int, values ...interface{}) (interface{}, error) {
	if its.typeOfDoc == TypeJSONArray {
		op := operations.NewDocInsertToArrayOperation(its.root, pos, values)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "InsertToArray not on JSONArray")
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
		target, ret, err := its.snapshot.InsertLocalInArray(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V...)
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
	return nil, errors.ErrDatatypeIllegalOperation.New(its.Logger, op.(iface.Operation).GetType())
}

func (its *document) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		var newSnap = newJSONObject(its.BaseDatatype, nil, model.OldestTimestamp)
		if err := json.Unmarshal([]byte(cast.C.Snapshot), newSnap); err != nil {
			return nil, errors.ErrDatatypeSnapshot.New(its.Logger, err.Error())
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
		its.snapshot.InsertRemoteInArray(cast.C.P, cast.C.T, cast.GetTimestamp(), cast.C.V...)
		return nil, nil
	case *operations.DocDeleteInObjectOperation:
		if _, err := its.snapshot.DeleteRemoteInObject(cast.C.P, cast.C.Key, cast.GetTimestamp()); err != nil {
			return nil, err
		}
		return nil, nil
	case *operations.DocDeleteInArrayOperation:
		errs := its.snapshot.DeleteRemoteInArray(cast.C.P, cast.GetTimestamp(), cast.C.T)
		if len(errs) > 0 {
			// TODO: have to deliver handler
		}
		return nil, nil
	}
	return nil, errors.ErrDatatypeIllegalOperation.New(its.Logger, op)
}

func (its *document) GetFromObject(key string) (Document, error) {
	if its.typeOfDoc == TypeJSONObject {
		if currentRoot, ok := its.snapshot.findJSONObject(its.root); ok {
			child := currentRoot.get(key).(jsonType)
			if child == nil {
				return nil, errors.ErrDatatypeNotExistChildDocument.New(its.Logger)
			}
			return its.getChildDocument(child), nil
		}

	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "GetFromObject not on JSONObject")
}

func (its *document) getChildDocument(child jsonType) *document {
	return &document{
		datatype:  its.datatype,
		root:      child.getKeyTime(),
		typeOfDoc: child.getType(),
		snapshot:  its.snapshot,
	}
}

func (its *document) GetFromArray(pos int) (Document, error) {
	if its.typeOfDoc == TypeJSONArray {
		if currentRoot, ok := its.snapshot.findJSONArray(its.root); ok {
			c, err := currentRoot.getTimedType(pos)
			if err != nil {
				return nil, err
			}
			child := c.(jsonType)
			return its.getChildDocument(child), nil
		}
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "GetFromArray not on JSONArray")
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
	return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "DeleteInObject not on JSONObject")
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
	return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "DeleteManyInArray not on JSONArray")
}

func (its *document) UpdateInArray(pos int, values ...interface{}) ([]interface{}, error) {
	if its.typeOfDoc == TypeJSONArray {
		if len(values) < 1 {
			return nil, errors.ErrDatatypeIllegalOperation.New(its.Logger, "at least one value should be inserted")
		}

		op := operations.NewDocUpdateInArrayOperation(its.root, pos, values)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret.([]interface{}), nil
	}
	return nil, errors.ErrDatatypeInvalidParent.New(its.Logger, "UpdateInArray not on JSONArray")
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
