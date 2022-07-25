package orda

import (
	"encoding/json"
	"fmt"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	datatypes2 "github.com/orda-io/orda/client/pkg/internal/datatypes"
	"github.com/orda-io/orda/client/pkg/model"
	operations2 "github.com/orda-io/orda/client/pkg/operations"
	"github.com/orda-io/orda/client/pkg/utils"
	"github.com/wI2L/jsondiff"
	"strconv"
	"strings"
)

// Document is an Orda datatype which provides document (JSON-like) interfaces.
type Document interface {
	Datatype
	DocumentInTx
	Transaction(tag string, txFunc func(document DocumentInTx) error) error
}

// DocumentInTx is an Orda datatype which provides document (JSON-like) interfaces in a transaction.
type DocumentInTx interface {
	PutToObject(key string, value interface{}) (Document, errors2.OrdaError)
	DeleteInObject(key string) (Document, errors2.OrdaError)

	InsertToArray(pos int, value ...interface{}) (Document, errors2.OrdaError)
	UpdateManyInArray(pos int, values ...interface{}) ([]Document, errors2.OrdaError)
	DeleteInArray(pos int) (Document, errors2.OrdaError)
	DeleteManyInArray(pos int, numOfNodes int) ([]Document, errors2.OrdaError)

	GetByPath(path string) (Document, errors2.OrdaError)

	GetFromObject(key string) (Document, errors2.OrdaError)
	GetFromArray(pos int) (Document, errors2.OrdaError)
	GetManyFromArray(pos int, numOfNodes int) ([]Document, errors2.OrdaError)
	GetValue() interface{}

	Patch(ops ...jsondiff.Operation) errors2.OrdaError
	PatchByJSON(jsonStr string) ([]jsondiff.Operation, errors2.OrdaError)

	GetParentDocument() Document
	GetRootDocument() Document
	GetTypeOfJSON() TypeOfJSON
	IsGarbage() bool
	Equal(o Document) bool

	ToJSONBytes() []byte
}

func newDocument(base *datatypes2.BaseDatatype, wire iface.Wire, handlers *Handlers) (Document, errors2.OrdaError) {
	doc := &document{
		datatype:         newDatatype(base, wire, handlers),
		SnapshotDatatype: datatypes2.NewSnapshotDatatype(base, nil),
	}
	return doc, doc.init(doc)
}

type document struct {
	*datatype
	*datatypes2.SnapshotDatatype
}

func (its *document) PatchByJSON(jsonStr string) ([]jsondiff.Operation, errors2.OrdaError) {

	if !json.Valid([]byte(jsonStr)) {
		return []jsondiff.Operation{}, errors2.DatatypeInvalidPatch.New(its.L(), "invalid json string:"+jsonStr)
	}
	patches, err := jsondiff.CompareJSON(its.ToJSONBytes(), []byte(jsonStr))
	if err != nil {
		return []jsondiff.Operation{}, errors2.DatatypeInvalidPatch.New(its.L(), "")
	}
	if len(patches) == 0 {
		return patches, nil
	}
	return patches, its.Patch(patches...)
}

func (its *document) ToJSONBytes() []byte {
	m, _ := json.Marshal(its.ToJSON())
	return m
}

func (its *document) Patch(patches ...jsondiff.Operation) errors2.OrdaError {
	if len(patches) == 1 {
		return its.patchEach(patches[0])
	} else {

		tag := fmt.Sprintf("%d patches-%s", len(patches), utils.HashSum(patches))
		if err := its.Transaction(tag, func(doc DocumentInTx) error {
			for _, patch := range patches {
				if err := doc.(*document).patchEach(patch); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			if ordaErr, ok := err.(errors2.OrdaError); ok {
				return ordaErr
			}
			return errors2.DatatypeInvalidPatch.New(its.L(), err.Error())
		}
	}
	return nil
}

func (its *document) GetByPath(path string) (Document, errors2.OrdaError) {
	path = strings.Trim(path, "/")
	paths := strings.Split(path, "/")
	if len(paths) == 1 && paths[0] == "" {
		return its, nil
	}
	target, err := its.snapshot().getTargetByPaths(paths)
	if err != nil {
		return nil, err
	}
	return its.toDocument(target), nil
}

func (its *document) patchEach(op jsondiff.Operation) errors2.OrdaError {
	// its.L().Infof("%v", op)
	target, key, err := its.snapshot().getTargetFromPatch(op.Path.String())
	if err != nil {
		return err
	}
	// its.L().Infof("target:%#v key:%v", target, key)
	switch op.Type {
	case jsondiff.OperationAdd:
		if op.Value == nil {
			return errors2.DatatypeInvalidPatch.New(its.L(), "invalid JSONPatch:"+op.String())
		}
		if target.getType() == TypeJSONObject {
			if _, err := its.toDocument(target).PutToObject(key, op.Value); err != nil {
				return err
			}
			// its.snapshot().PutCommonInObject(target.getTime(), key, op.Value, its.snapshot().getTime())
		} else if target.getType() == TypeJSONArray {
			var pos int
			if key == "-" {
				pos = target.(*jsonArray).size
			} else {
				var err2 error
				pos, err2 = strconv.Atoi(key)
				if err2 != nil {
					return errors2.DatatypeInvalidPatch.New(its.L(), "invalid array position in JSONPatch:"+op.String())
				}
			}
			if _, err := its.toDocument(target).InsertToArray(pos, op.Value); err != nil {
				return err
			}
		}
		return nil
	case jsondiff.OperationRemove:
		if target.getType() == TypeJSONObject {
			if _, err := its.toDocument(target).DeleteInObject(key); err != nil {
				return err
			}
		} else if target.getType() == TypeJSONArray {
			pos, err := strconv.Atoi(key)
			if err != nil {
				return errors2.DatatypeInvalidPatch.New(its.L(), "invalid array position in JSONPatch:"+op.String())
			}
			if _, err := its.toDocument(target).DeleteInArray(pos); err != nil {
				return err
			}
		}
		return nil
	case jsondiff.OperationReplace:
		if op.Value == nil {
			return errors2.DatatypeInvalidPatch.New(its.L(), "invalid JSONPatch:"+op.String())
		}
		if target.getType() == TypeJSONObject {
			if _, err := its.toDocument(target).PutToObject(key, op.Value); err != nil {
				return err
			}
		} else if target.getType() == TypeJSONArray {
			pos, err := strconv.Atoi(key)
			if err != nil {
				return errors2.DatatypeInvalidPatch.New(its.L(), "invalid array position in JSONPatch:"+op.String())
			}
			if _, err := its.toDocument(target).UpdateManyInArray(pos, op.Value); err != nil {
				return err
			}
		}
	default:
		return errors2.DatatypeInvalidPatch.New(its.L(), "unsupported JSONPatch:"+op.String())
	}
	return nil
}

func (its *document) Transaction(tag string, userFunc func(document DocumentInTx) error) error {
	return its.DoTransaction(tag, its.TxCtx, func(txCtx *datatypes2.TransactionContext) error {
		clone := &document{
			datatype:         its.cloneDatatype(txCtx),
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return userFunc(clone)
	})
}

func (its *document) snapshot() jsonType {
	return its.GetSnapshot().(jsonType)
}

func (its *document) ResetSnapshot() {
	its.Snapshot = newJSONObject(its.BaseDatatype, nil, model.OldestTimestamp())
}

func (its *document) ToJSON() interface{} {
	return its.snapshot().ToJSON()
}

func (its *document) GetValue() interface{} {
	return its.snapshot().ToJSON()
}

func (its *document) ExecuteLocal(op interface{}) (interface{}, errors2.OrdaError) {
	switch cast := op.(type) {
	case *operations2.DocPutInObjOperation:
		return its.snapshot().PutCommonInObject(cast.GetBody().P, cast.GetBody().K, cast.GetBody().V, cast.GetTimestamp())
	case *operations2.DocRemoveInObjOperation:
		return its.snapshot().DeleteCommonInObject(cast.GetBody().P, cast.GetBody().K, cast.GetTimestamp(), true)
	case *operations2.DocInsertToArrayOperation:
		target, parent, err := its.snapshot().InsertLocalInArray(cast.GetBody().P, cast.Pos, cast.ID.GetTimestamp(), cast.GetBody().V...)
		if err != nil {
			return nil, err
		}
		cast.GetBody().T = target
		return parent, nil
	case *operations2.DocDeleteInArrayOperation:
		delTargets, delJSONTypes, err := its.snapshot().DeleteLocalInArray(cast.GetBody().P, cast.Pos, cast.NumOfNodes, cast.ID.GetTimestamp())
		if err != nil {
			return nil, err
		}
		cast.GetBody().T = delTargets
		return delJSONTypes, nil
	case *operations2.DocUpdateInArrayOperation:
		uptTargets, oldOnes, err := its.snapshot().UpdateLocalInArray(cast.GetBody().P, cast.Pos, cast.ID.GetTimestamp(), cast.GetBody().V...)
		if err != nil {
			return nil, err
		}
		cast.GetBody().T = uptTargets
		return oldOnes, nil
	}
	return nil, errors2.DatatypeIllegalOperation.New(its.L(), its.TypeOf.String(), op)
}

func (its *document) ExecuteRemote(op interface{}) (interface{}, errors2.OrdaError) {
	switch cast := op.(type) {
	case *operations2.SnapshotOperation:
		return nil, its.ApplySnapshot(cast.GetBody())
	case *operations2.DocPutInObjOperation:
		return its.snapshot().PutCommonInObject(cast.GetBody().P, cast.GetBody().K, cast.GetBody().V, cast.GetTimestamp())
	case *operations2.DocRemoveInObjOperation:
		return its.snapshot().DeleteCommonInObject(cast.GetBody().P, cast.GetBody().K, cast.GetTimestamp(), false)
	case *operations2.DocInsertToArrayOperation:
		return its.snapshot().InsertRemoteInArray(cast.GetBody().P, cast.GetBody().T, cast.GetTimestamp(), cast.GetBody().V...)
	case *operations2.DocDeleteInArrayOperation:
		return its.snapshot().DeleteRemoteInArray(cast.GetBody().P, cast.GetTimestamp(), cast.GetBody().T)
	case *operations2.DocUpdateInArrayOperation:
		return its.snapshot().UpdateRemoteInArray(cast.GetBody().P, cast.GetTimestamp(), cast.GetBody().T, cast.GetBody().V)
	}
	return nil, errors2.DatatypeIllegalOperation.New(its.L(), its.TypeOf.String(), op)
}

// PutToObject associates a new value with the given key, and returns the old value as a Document
func (its *document) PutToObject(key string, value interface{}) (Document, errors2.OrdaError) {
	if err := its.assertLocalOp("PutToObject", TypeJSONObject, false); err != nil {
		return nil, err
	}
	op := operations2.NewDocPutInObjOperation(its.snapshot().getCreateTime(), key, value)
	removed, err := its.SentenceInTx(its.TxCtx, op, true)
	if err != nil {
		return nil, err
	}
	if removed != nil {
		return its.toDocument(removed.(jsonType)), nil
	}
	return nil, nil
}

// DeleteInObject removes the value associated with the given key, and returns the removed value as a Document.
func (its *document) DeleteInObject(key string) (Document, errors2.OrdaError) {
	if err := its.assertLocalOp("DeleteInObject", TypeJSONObject, false); err != nil {
		return nil, err
	}
	op := operations2.NewDocRemoveInObjOperation(its.snapshot().getCreateTime(), key)
	removed, err := its.SentenceInTx(its.TxCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocument(removed.(jsonType)), nil
}

// GetFromObject returns the child associated with the given key as a Document.
func (its *document) GetFromObject(key string) (Document, errors2.OrdaError) {
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

// GetFromArray returns the element of the JSONArray Document at the given position.
func (its *document) GetFromArray(pos int) (Document, errors2.OrdaError) {
	ret, err := its.GetManyFromArray(pos, 1)
	if err != nil {
		return nil, err
	}
	return ret[0], nil
}

func (its *document) GetManyFromArray(pos int, numOfNodes int) ([]Document, errors2.OrdaError) {
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
func (its *document) InsertToArray(pos int, values ...interface{}) (Document, errors2.OrdaError) {
	if err := its.assertLocalOp("InsertToArray", TypeJSONArray, false); err != nil {
		return its, err
	}
	arr := its.snapshot().(*jsonArray)
	if err := arr.validateInsertPosition(pos); err != nil {
		return its, err
	}
	op := operations2.NewDocInsertToArrayOperation(its.snapshot().getCreateTime(), pos, values)
	if _, err := its.SentenceInTx(its.TxCtx, op, true); err != nil {
		return its, err
	}
	return its, nil
}

// DeleteInArray deletes a value at the given position, and returns the deleted Document.
func (its *document) DeleteInArray(pos int) (Document, errors2.OrdaError) {
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
func (its *document) DeleteManyInArray(pos int, numOfNodes int) ([]Document, errors2.OrdaError) {
	if err := its.assertLocalOp("DeleteManyInArray", TypeJSONArray, false); err != nil {
		return nil, err
	}
	arr := its.snapshot().(*jsonArray)
	if err := arr.validateGetRange(pos, numOfNodes); err != nil {
		return nil, err
	}
	op := operations2.NewDocDeleteInArrayOperation(its.snapshot().getCreateTime(), pos, numOfNodes)
	delJSONTypes, err := its.SentenceInTx(its.TxCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocuments(delJSONTypes.([]jsonType)), nil
}

// UpdateManyInArray updates the child from the given position, and returns the previous child Documents
func (its *document) UpdateManyInArray(pos int, values ...interface{}) ([]Document, errors2.OrdaError) {
	if err := its.assertLocalOp("UpdateManyInArray", TypeJSONArray, false); err != nil {
		return nil, err
	}
	arr := its.snapshot().(*jsonArray)
	if err := arr.validateGetRange(pos, len(values)); err != nil {
		return nil, err
	}
	op := operations2.NewDocUpdateInArrayOperation(its.snapshot().getCreateTime(), pos, values)
	oldOnes, err := its.SentenceInTx(its.TxCtx, op, true)
	if err != nil {
		return nil, err
	}
	return its.toDocuments(oldOnes.([]jsonType)), nil
}

func (its *document) GetTypeOfJSON() TypeOfJSON {
	return its.snapshot().getType()
}

func (its *document) IsGarbage() bool {
	return its.snapshot().isGarbage()
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
		datatype:         its.datatype,
		SnapshotDatatype: datatypes2.NewSnapshotDatatype(its.BaseDatatype, child),
	}
}
func (its *document) assertLocalOp(opName string, ofJSON TypeOfJSON, workOnGarbage bool) errors2.OrdaError {
	if its.GetTypeOfJSON() != ofJSON {
		return errors2.DatatypeInvalidParent.New(its.L(), opName, " is not allowed to ")
	}
	if !workOnGarbage && its.snapshot().isGarbage() {
		return errors2.DatatypeNoOp.New(its.L(), "already deleted from the root Document")
	}
	return nil
}
