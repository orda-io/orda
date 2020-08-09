package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
	"reflect"
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
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
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
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *document) ExecuteLocal(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.DocPutInObjectOperation:
		if _, err := its.snapshot.PutInObjectCommon(cast); err != nil {
			return nil, err
		}
		return its, nil
	case *operations.DocInsertToArrayOperation:
		target, ret, err := its.snapshot.InsertLocal(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V...)
		if err != nil {
			return nil, err
		}
		cast.C.T = target
		return ret, nil
	case *operations.DocDeleteInObjectOperation:
		its.snapshot.DeleteCommonInObject(cast.C.P, cast.C.Key, cast.ID.GetTimestamp())
		return nil, nil
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
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op.(iface.Operation).GetType())
}

func (its *document) ExecuteRemote(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		var newSnap = newJSONObject(nil, model.OldestTimestamp)
		if err := json.Unmarshal([]byte(cast.C.Snapshot), newSnap); err != nil {
			return nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
		}
		its.snapshot = newSnap
		// its.datatype.SetOpID()
		return nil, nil
	case *operations.DocPutInObjectOperation:
		if _, err := its.snapshot.PutInObjectCommon(cast); err != nil {
			return nil, err
		}
		return nil, nil
	case *operations.DocInsertToArrayOperation:
		its.snapshot.InsertRemote(cast.C.P, cast.C.T, cast.ID.GetTimestamp(), cast.C.V...)
		return nil, nil
	case *operations.DocDeleteInObjectOperation:
		its.snapshot.DeleteCommonInObject(cast.C.P, cast.C.Key, cast.ID.GetTimestamp())
		return nil, nil
	case *operations.DocDeleteInArrayOperation:
		its.snapshot.DeleteRemoteInArray(cast.C.P, cast.C.T, cast.ID.GetTimestamp())
		return nil, nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *document) GetFromObject(key string) (Document, error) {
	if its.typeOfDoc == TypeJSONObject {
		if currentRoot, ok := its.snapshot.findJSONObject(its.root); ok {
			child := currentRoot.get(key).(jsonType)
			if child == nil {
				return nil, errors.NewDatatypeError(errors.ErrDatatypeNotExistChildDocument)
			}
			return its.getChildDocument(child), nil
		}

	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *document) getChildDocument(child jsonType) *document {
	return &document{
		datatype:  its.datatype,
		root:      child.getTime(),
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
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
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
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
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
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *document) UpdateInArray(pos int, values ...interface{}) ([]interface{}, error) {
	if its.typeOfDoc == TypeJSONArray {
		if len(values) < 1 {
			return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "at least one value should be inserted")
		}

		op := operations.NewDocUpdateInArrayOperation(its.root, pos, values)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret.([]interface{}), nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
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
		return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return meta, its.snapshot, nil
}

func (its *document) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	log.Logger.Infof("SetMetaAndSnapshot:%v", snapshot)
	if err := its.ManageableDatatype.SetMeta(meta); err != nil {
		return errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	if err := json.Unmarshal([]byte(snapshot), its.snapshot); err != nil {
		return errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return nil
}

// ////////////////////////////////////////////////////////////////
//  jsonSnapshot
// ////////////////////////////////////////////////////////////////

// ////////////////////////////////////
//  jsonElement
// ////////////////////////////////////

type jsonElement struct {
	jsonType
	V types.JSONValue
}

func newJSONElement(parent jsonType, value interface{}, ts *model.Timestamp) *jsonElement {
	return &jsonElement{
		jsonType: &jsonPrimitive{
			parent: parent,
			common: parent.getRoot(),
			K:      ts,
			P:      ts,
		},
		V: value,
	}
}

func (its *jsonElement) getValue() types.JSONValue {
	return its.V
}

func (its *jsonElement) getType() TypeOfJSON {
	return TypeJSONElement
}

func (its *jsonElement) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonElement) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getTime().ToString()
	}
	value := its.V
	if its.isTomb() {
		value = "#!DELETED"
	}
	return fmt.Sprintf("JE(P%v)[T%v|%v]", parentTS, its.getTime().ToString(), value)
}

// ////////////////////////////////////
//  jsonObject
// ////////////////////////////////////

type jsonObject struct {
	jsonType
	*hashMapSnapshot
}

func newJSONObject(parent jsonType, ts *model.Timestamp) *jsonObject {
	var root *jsonCommon
	if parent == nil {
		root = &jsonCommon{
			root:     nil,
			nodeMap:  make(map[string]jsonType),
			cemetery: make(map[string]jsonType),
		}
	} else {
		root = parent.getRoot()
	}
	obj := &jsonObject{
		jsonType: &jsonPrimitive{
			common: root,
			parent: parent,
			K:      ts,
			P:      ts,
		},
		hashMapSnapshot: newHashMapSnapshot(),
	}
	obj.jsonType.setRoot(obj)
	return obj
}

func (its *jsonObject) CloneSnapshot() iface.Snapshot {
	// TODO: implement CloneSnapshot()
	return &jsonObject{}
}

func (its *jsonObject) PutInObjectCommon(op *operations.DocPutInObjectOperation) (jsonType, error) {
	if parentObj, ok := its.findJSONObject(op.C.P); ok {
		return parentObj.put(op.C.K, op.C.V, op.GetTimestamp()), nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) InsertLocal(parent *model.Timestamp, pos int, ts *model.Timestamp, values ...interface{}) (*model.Timestamp, []interface{}, error) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.arrayInsertCommon(pos, nil, ts, values...)
	}
	return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) InsertRemote(parent *model.Timestamp, target, ts *model.Timestamp, values ...interface{}) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		_, _, _ = parentArray.arrayInsertCommon(-1, target, ts, values...)
		return
	}
	_ = errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) DeleteCommonInObject(parent *model.Timestamp, key string, ts *model.Timestamp) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		parentObj.objDeleteCommon(key, ts)
		return
	}
	_ = errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) UpdateLocalInArray(op *operations.DocUpdateInArrayOperation) {
	// if parentArray, ok := its.findJSONArray(op.C.P); ok {
	// 	return parentArray.arrayUpdateLocal(op)
	// }
	// return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) DeleteLocalInArray(
	parent *model.Timestamp, pos, numOfNodes int, ts *model.Timestamp) ([]*model.Timestamp, []interface{}, error) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		return parentArray.arrayDeleteLocal(pos, numOfNodes, ts)
	}
	return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) DeleteRemoteInArray(parent *model.Timestamp, targets []*model.Timestamp, ts *model.Timestamp) {
	if parentArray, ok := its.findJSONArray(parent); ok {
		parentArray.arrayDeleteRemote(targets, ts)
		return
	}
	_ = errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *jsonObject) getValue() types.JSONValue {
	return its.hashMapSnapshot
}

func (its *jsonObject) getType() TypeOfJSON {
	return TypeJSONObject
}

func (its *jsonObject) makeTomb(ts *model.Timestamp) bool {
	// log.Logger.Infof("makeTomb() of jsonObject:%v", its.getValue())
	if its.jsonType.makeTomb(ts) {
		// for _, v := range its.Map {
		//
		// 	switch cast := v.(type) {
		// 	case *jsonElement:
		// 		// if !cast.isTomb() {
		// 		cast.makeTomb(ts)
		// 		// }
		// 	case *jsonObject:
		// 		// if !cast.isTomb() {
		// 		cast.makeTomb(ts)
		// 		// cast.deleteChildren(ts)
		// 		// }
		// 	case *jsonArray:
		// 		// if !cast.isTomb() {
		// 		cast.makeTomb(ts)
		// 		// cast.deleteChildren(ts)
		// 		// }
		// 	}
		// }
		return true
	}
	return false
}

func (its *jsonObject) objDeleteCommon(key string, ts *model.Timestamp) interface{} {
	target := its.get(key)
	if _, ok := target.(jsonType); ok {
		// if !j.isTomb() {
		ret := its.removeCommon(key, ts)
		return ret
		// }
	}
	return nil
}

func (its *jsonObject) put(key string, value interface{}, ts *model.Timestamp) jsonType {
	rt := reflect.ValueOf(value)
	var primitive jsonType
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		primitive = its.createJSONArray(its, value, ts)
	case reflect.Struct, reflect.Map:
		primitive = its.createJSONObject(its, value, ts)
	case reflect.Ptr:
		val := rt.Elem()
		primitive = its.put(key, val.Interface(), ts) // recursively
	default:
		primitive = newJSONElement(its, types.ConvertToJSONSupportedValue(value), ts.NextDeliminator())
	}
	removed, _ := its.putCommonWithTimedValue(key, primitive) // in hash map
	if removed != nil {
		its.addToCemetery(removed.(jsonType))
	}
	its.addToNodeMap(primitive)
	return primitive
}

func (its *jsonObject) getChildAsJSONElement(key string) *jsonElement {
	value := its.get(key)
	if value == nil {
		return nil
	}
	return value.(*jsonElement)
}

func (its *jsonObject) getChildAsJSONObject(key string) *jsonObject {
	value := its.get(key)
	return value.(*jsonObject)
}

func (its *jsonObject) getChildAsJSONArray(key string) *jsonArray {
	value := its.get(key)
	return value.(*jsonArray)
}

func (its *jsonObject) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JO(%v)[T%v|V%v]", parentTS, its.getTime().ToString(), its.hashMapSnapshot.String())
}

func (its *jsonObject) GetAsJSONCompatible() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v != nil {
			switch cast := v.(type) {
			case *jsonObject:
				if !cast.isTomb() {
					m[k] = cast.GetAsJSONCompatible()
				}

			case *jsonElement:
				if !cast.isTomb() {
					m[k] = v.getValue()
				}
			case *jsonArray:
				if !cast.isTomb() {
					m[k] = cast.GetAsJSONCompatible()
				}
			}
		}
	}
	return m
}

// ////////////////////////////////////
//  jsonArray
// ////////////////////////////////////

type jsonArray struct {
	jsonType
	*listSnapshot
}

func newJSONArray(parent jsonType, ts *model.Timestamp) *jsonArray {
	return &jsonArray{
		jsonType: &jsonPrimitive{
			parent: parent,
			common: parent.getRoot(),
			K:      ts,
			P:      ts,
		},
		listSnapshot: newListSnapshot(),
	}
}

func (its *jsonArray) makeTomb(ts *model.Timestamp) bool {
	log.Logger.Infof("makeTomb() of jsonArray")
	if its.jsonType.makeTomb(ts) {
		return true
	}
	return false
}

func (its *jsonArray) arrayDeleteRemote(targets []*model.Timestamp, ts *model.Timestamp) {
	for _, t := range targets {
		if j, ok := its.findJSONPrimitive(t); ok {
			if !j.isTomb() {
				j.makeTomb(ts)
				its.size--
			} else { // concurrent deletes
				if j.getPrecedence().Compare(ts) < 0 {
					j.setTime(ts)
				}
			}
		} else {
			log.Logger.Warnf("fail to find delete target: %v", t.ToString())
		}
	}
	// return nil, nil
}

func (its *jsonArray) arrayUpdateLocal(op *operations.DocUpdateInArrayOperation) error {
	if err := its.validateRange(op.Pos, len(op.C.V)); err != nil {
		return err
	}
	// orderedType := its.findNthTarget(op.Pos + 1)
	// for i := 0; i < len(op.C.V); i++ {
	//
	// }
	return nil
}

func (its *jsonArray) arrayDeleteLocal(pos, numOfNodes int, ts *model.Timestamp) ([]*model.Timestamp, []interface{}, error) {
	targets, values, err := its.listSnapshot.deleteLocal(pos, numOfNodes, ts)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range targets {
		if jt, ok := its.findJSONPrimitive(v); ok {
			its.addToCemetery(jt)
		}
	}
	return targets, values, err
	// if err := its.validateRange(pos, numOfNodes); err != nil {
	// 	return nil, nil, err
	// }
	// var deletedTargets []*model.Timestamp
	// var deletedValues []interface{}
	// node := its.findNthTarget(pos + 1)
	// for i := 0; i < numOfNodes; i++ {
	// 	cast := node.getPrecededType().(jsonType)
	// 	if cast.makeTomb(ts) {
	// 		deletedTargets = append(deletedTargets, cast.getTime())
	// 		deletedValues = append(deletedValues, cast)
	// 	}
	// 	// switch cast := orderedType.timedType.(type) {
	// 	// case *jsonElement:
	// 	// 	if !cast.isTomb() {
	// 	// 		cast.makeTomb(ts)
	// 	// 		deletedTargets = append(deletedTargets, cast.getKey())
	// 	// 		deletedValues = append(deletedValues, cast.getValue())
	// 	// 	}
	// 	// case *jsonObject:
	// 	// 	if !cast.isTomb() {
	// 	// 		cast.makeTomb(ts)
	// 	// 		deletedTargets = append(deletedTargets, cast.getKey())
	// 	// 		deletedValues = append(deletedValues, cast.getValue())
	// 	// 	}
	// 	// case *jsonArray:
	// 	// 	if !cast.isTomb() {
	// 	// 		cast.makeTomb(ts)
	// 	// 		deletedTargets = append(deletedTargets, cast.getKey())
	// 	// 		deletedValues = append(deletedValues, cast.getValue())
	// 	// 	}
	// 	// }
	// 	node = node.getNextLive()
	// }
	// return deletedTargets, deletedValues, nil

}

func (its *jsonArray) arrayInsertCommon(
	pos int, // in the case of the local insert
	target *model.Timestamp, // in the case of the remote insert
	ts *model.Timestamp,
	values ...interface{},
) (*model.Timestamp, []interface{}, error) {
	var pts []precededType
	for _, v := range values {
		rt := reflect.ValueOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(its, v, ts)
			pts = append(pts, ja)
			its.addToNodeMap(ja)
		case reflect.Struct, reflect.Map:
			jo := its.createJSONObject(its, v, ts)
			pts = append(pts, jo)
			its.addToNodeMap(jo)
		case reflect.Ptr:
			ptrVal := rt.Elem()
			its.arrayInsertCommon(pos, target, ts, ptrVal)
		default:
			je := newJSONElement(its, types.ConvertToJSONSupportedValue(v), ts.NextDeliminator())
			pts = append(pts, je)
			its.addToNodeMap(je)
		}
	}
	if target == nil { // InsertLocal
		return its.listSnapshot.insertLocalWithPrecededTypes(pos, pts...)
	} else { // InsertRemote
		its.listSnapshot.insertRemoteWithPrecededTypes(target, ts, pts...)
		return nil, nil, nil
	}
}

func (its *jsonArray) getAsJSONElement(pos int) (*jsonElement, error) {
	val, err := its.listSnapshot.getPrecededType(pos)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return val.(*jsonElement), nil
}

func (its *jsonArray) getValue() types.JSONValue {
	return its.listSnapshot
}

func (its *jsonArray) getType() TypeOfJSON {
	return TypeJSONArray
}

func (its *jsonArray) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonArray) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JA(%v)[T%v|V%v", parentTS, its.getTime().ToString(), its.listSnapshot.String())
}

func (its *jsonArray) GetAsJSONCompatible() interface{} {
	var list []interface{}
	n := its.listSnapshot.head.getNextLive()
	for n != nil {
		if !n.isTomb() {
			switch cast := n.getPrecededType().(type) {
			case *jsonObject:
				list = append(list, cast.GetAsJSONCompatible())
			case *jsonElement:
				list = append(list, cast.getValue())
			case *jsonArray:
				list = append(list, cast.GetAsJSONCompatible())
			}
		}
		n = n.getNextLive()
	}
	return list
}
