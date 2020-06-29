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
	GetDocumentType() TypeOfDocument
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
	typeOfDoc TypeOfDocument
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
		op := operations.NewPutObjectOperation(its.root, key, value)
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
		op := operations.NewInsArrayOperation(its.root, pos, values)
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
	case *operations.PutObjectOperation:
		its.snapshot.AddCommon(cast.C.P, cast.C.K, cast.C.V, cast.ID.GetTimestamp())
		return nil, nil
	case *operations.InsArrayOperation:
		target, ret, err := its.snapshot.InsertLocal(cast.C.P, cast.Pos, cast.ID.GetTimestamp(), cast.C.V...)
		if err != nil {
			return nil, err
		}
		cast.C.T = target
		return ret, nil
	case *operations.DelInObjectOperation:
		its.snapshot.DeleteCommonInObject(cast.C.P, cast.C.Key, cast.ID.GetTimestamp())
		return nil, nil
	case *operations.DelInArrayOperation:
		deletedTargets, deletedValues, err := its.snapshot.DeleteLocalInArray(cast.C.P, cast.Pos, cast.NumOfNodes, cast.ID.GetTimestamp())
		if err != nil {
			return nil, err
		}
		cast.C.T = deletedTargets
		return deletedValues, nil
	case *operations.UpdInArrayOperation:
		its.snapshot.UpdateLocaInArray()
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
		return nil, nil
	case *operations.PutObjectOperation:
		its.snapshot.AddCommon(cast.C.P, cast.C.K, cast.C.V, cast.ID.GetTimestamp())
		return nil, nil
	case *operations.InsArrayOperation:
		its.snapshot.InsertRemote(cast.C.P, cast.C.T, cast.ID.GetTimestamp(), cast.C.V...)
		return nil, nil
	case *operations.DelInObjectOperation:
		its.snapshot.DeleteCommonInObject(cast.C.P, cast.C.Key, cast.ID.GetTimestamp())
		return nil, nil
	case *operations.DelInArrayOperation:
		its.snapshot.DeleteRemoteInArray(cast.C.P, cast.C.T, cast.ID.GetTimestamp())
		return nil, nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *document) GetFromObject(key string) (Document, error) {
	if its.typeOfDoc == TypeJSONObject {
		if currentRoot, ok := its.snapshot.findJSONObject(its.root); ok {
			child := currentRoot.get(key).(jsonPrimitive)
			if child == nil {
				return nil, errors.NewDatatypeError(errors.ErrDatatypeNotExistChildDocument)
			}
			return &document{
				datatype:  its.datatype,
				root:      child.getKey(),
				typeOfDoc: child.getType(),
				snapshot:  its.snapshot,
			}, nil
		}

	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *document) GetFromArray(pos int) (Document, error) {
	if its.typeOfDoc == TypeJSONArray {
		if currentRoot, ok := its.snapshot.findJSONArray(its.root); ok {
			c, err := currentRoot.getTimedValue(pos)
			if err != nil {
				return nil, err
			}
			child := c.(jsonPrimitive)
			return &document{
				datatype:  its.datatype,
				root:      child.getKey(),
				typeOfDoc: child.getType(),
				snapshot:  its.snapshot,
			}, nil
		}

	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *document) DeleteInObject(key string) (interface{}, error) {
	if its.typeOfDoc == TypeJSONObject {
		op := operations.NewDelInObjectOperation(its.root, key)
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
		op := operations.NewDelInArrayOperation(its.root, pos, numOfNodes)
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

		op := operations.NewUpdInArrayOperation(its.root, pos, values)
		ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
		if err != nil {
			return nil, err
		}
		return ret.([]interface{}), nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
}

func (its *document) GetDocumentType() TypeOfDocument {
	return its.typeOfDoc
}

func (its *document) GetAsJSON() interface{} {
	r, _ := its.snapshot.findJSONPrimitive(its.root)
	switch cast := r.(type) {
	case *jsonObject:
		return cast.GetAsJSON()
	case *jsonArray:
		return cast.GetAsJSON()
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
	panic("implement me")
}

// ////////////////////////////////////////////////////////////////
//  jsonSnapshot
// ////////////////////////////////////////////////////////////////

type TypeOfDocument int

const (
	typeJSONPrimitive TypeOfDocument = iota
	TypeJSONElement
	TypeJSONObject
	TypeJSONArray
)

// jsonPrimitive extends timedValue
// jsonElement extends jsonPrimitive
// jsonObject extends jsonPrimitive
// jsonArray extends jsonPrimitive

// ////////////////////////////////////
//  jsonPrimitive
// ////////////////////////////////////

type jsonPrimitive interface {
	timedValue
	getType() TypeOfDocument
	getRoot() *jsonRoot
	setRoot(r *jsonObject)
	findJSONPrimitive(ts *model.Timestamp) (j jsonPrimitive, ok bool)
	findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool)
	findJSONObject(ts *model.Timestamp) (j *jsonObject, ok bool)
	findJSONArray(ts *model.Timestamp) (j *jsonArray, ok bool)
	getParent() jsonPrimitive
	getKey() *model.Timestamp
	putInNodeMap(tv jsonPrimitive)
	getParentAsJSONObject() *jsonObject
	createJSONObject(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonObject
	createJSONArray(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonArray
}

type jsonRoot struct {
	nodeMap map[string]jsonPrimitive
	root    *jsonObject
}

type jsonPrimitiveImpl struct {
	parent  jsonPrimitive
	root    *jsonRoot
	K       *model.Timestamp // this is used for key that is immutable and used in the root
	P       *model.Timestamp // this is used for precedence
	deleted bool
}

func (its *jsonPrimitiveImpl) isTomb() bool {
	return its.deleted
}

// 어떤 애가 지워졌는지는 parent까지 조사해야 하나?
func (its *jsonPrimitiveImpl) makeTomb(ts *model.Timestamp) bool {
	if its.deleted {
		if its.P.Compare(ts) > 0 { // This condition makes newer timestamps remain in nodes.
			log.Logger.Infof("fail to makeTomb() of jsonPrimitiveImpl:%v", its.K.ToString())
			return false
		}
	}
	its.P = ts
	its.deleted = true
	log.Logger.Infof("makeTomb() of jsonPrimitiveImpl:%v", its.K.ToString())
	return true
}

func (its *jsonPrimitiveImpl) getKey() *model.Timestamp {
	return its.K
}

func (its *jsonPrimitiveImpl) getTime() *model.Timestamp {
	return its.P
}

func (its *jsonPrimitiveImpl) setTime(ts *model.Timestamp) {
	its.P = ts
}

func (its *jsonPrimitiveImpl) findJSONPrimitive(ts *model.Timestamp) (j jsonPrimitive, ok bool) {
	node, ok := its.getRoot().nodeMap[ts.Hash()]
	return node, ok
}

func (its *jsonPrimitiveImpl) findJSONElement(ts *model.Timestamp) (j *jsonElement, ok bool) {
	if node, ok := its.getRoot().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonElement); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitiveImpl) findJSONObject(ts *model.Timestamp) (json *jsonObject, ok bool) {
	if node, ok := its.getRoot().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonObject); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitiveImpl) findJSONArray(ts *model.Timestamp) (json *jsonArray, ok bool) {
	if node, ok := its.getRoot().nodeMap[ts.Hash()]; ok {
		if j, ok2 := node.(*jsonArray); ok2 {
			return j, ok2
		}
	}
	return nil, false
}

func (its *jsonPrimitiveImpl) putInNodeMap(primitive jsonPrimitive) {
	its.getRoot().nodeMap[primitive.getKey().Hash()] = primitive
}

func (its *jsonPrimitiveImpl) getValue() types.JSONValue {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) setValue(v types.JSONValue) {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) getRoot() *jsonRoot {
	return its.root
}

func (its *jsonPrimitiveImpl) setRoot(r *jsonObject) {
	its.root.root = r
	its.root.nodeMap[r.getKey().Hash()] = r
}

func (its *jsonPrimitiveImpl) getType() TypeOfDocument {
	return typeJSONPrimitive
}

func (its *jsonPrimitiveImpl) getParent() jsonPrimitive {
	return its.parent
}

func (its *jsonPrimitiveImpl) getParentAsJSONObject() *jsonObject {
	return its.parent.(*jsonObject)
}

func (its *jsonPrimitiveImpl) String() string {
	return fmt.Sprintf("%x", &its.parent)
}

func (its *jsonPrimitiveImpl) createJSONArray(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonArray {
	ja := newJSONArray(parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	var appendValues []timedValue
	for i := 0; i < target.Len(); i++ {
		field := target.Index(i)
		switch field.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(ja, field.Interface(), ts)
			appendValues = append(appendValues, ja)
		case reflect.Struct:
			childJO := its.createJSONObject(ja, field.Interface(), ts)
			appendValues = append(appendValues, childJO)
		case reflect.Ptr:
			val := field.Elem()
			its.createJSONArray(parent, val.Interface(), ts)
		default:
			element := newJSONElement(ja, types.ConvertToJSONSupportedValue(field.Interface()), ts.NextDeliminator())
			appendValues = append(appendValues, element)
		}
	}
	if appendValues != nil {
		ja.insertLocalWithTimedValue(0, appendValues...)
		for _, v := range appendValues {
			its.putInNodeMap(v.(jsonPrimitive))
		}
	}
	// log.Logger.Infof("%v", ja.String())
	return ja
}

func (its *jsonPrimitiveImpl) createJSONObject(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonObject {
	jo := newJSONObject(parent, ts.NextDeliminator())
	target := reflect.ValueOf(value)
	fields := reflect.TypeOf(value)

	if target.Kind() == reflect.Map {
		mapValue := value.(map[string]interface{})
		for k, v := range mapValue {
			val := reflect.ValueOf(v)
			its.addValueToJSONObject(jo, k, val, ts)
		}
	} else {
		for i := 0; i < target.NumField(); i++ {
			value := target.Field(i)
			its.addValueToJSONObject(jo, fields.Field(i).Name, value, ts)
		}
	}

	return jo
}

func (its *jsonPrimitiveImpl) addValueToJSONObject(jo *jsonObject, key string, value reflect.Value, ts *model.Timestamp) {
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		ja := its.createJSONArray(jo, value.Interface(), ts)
		jo.putCommonWithTimedValue(key, ja)
		its.putInNodeMap(ja)
	case reflect.Struct:
		childJO := its.createJSONObject(jo, value.Interface(), ts)
		jo.putCommonWithTimedValue(key, childJO)
		its.putInNodeMap(childJO)
	case reflect.Ptr:
		val := value.Elem()
		its.createJSONObject(jo, val.Interface(), ts)
	default:
		element := newJSONElement(jo, types.ConvertToJSONSupportedValue(value.Interface()), ts.NextDeliminator())
		jo.putCommonWithTimedValue(key, element)
		its.putInNodeMap(element)
	}
}

// ////////////////////////////////////
//  jsonElement
// ////////////////////////////////////

type jsonElement struct {
	jsonPrimitive
	V types.JSONValue
}

func newJSONElement(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonElement {
	return &jsonElement{
		jsonPrimitive: &jsonPrimitiveImpl{
			parent: parent,
			root:   parent.getRoot(),
			K:      ts,
			P:      ts,
		},
		V: value,
	}
}

func (its *jsonElement) getValue() types.JSONValue {
	return its.V
}

func (its *jsonElement) getType() TypeOfDocument {
	return TypeJSONElement
}

func (its *jsonElement) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonElement) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getKey().ToString()
	}
	return fmt.Sprintf("JE(P%v)[T%v|%v]", parentTS, its.getKey().ToString(), its.V)
}

// ////////////////////////////////////
//  jsonObject
// ////////////////////////////////////

type jsonObject struct {
	jsonPrimitive
	*hashMapSnapshot
}

func newJSONObject(parent jsonPrimitive, ts *model.Timestamp) *jsonObject {
	var root *jsonRoot
	if parent == nil {
		root = &jsonRoot{
			nodeMap: make(map[string]jsonPrimitive),
			root:    nil,
		}
	} else {
		root = parent.getRoot()
	}
	obj := &jsonObject{
		jsonPrimitive: &jsonPrimitiveImpl{
			root:   root,
			parent: parent,
			K:      ts,
			P:      ts,
		},
		hashMapSnapshot: newHashMapSnapshot(),
	}
	obj.jsonPrimitive.setRoot(obj)
	return obj
}

func (its *jsonObject) CloneSnapshot() iface.Snapshot {
	// TODO: implement CloneSnapshot()
	return &jsonObject{}
}

func (its *jsonObject) AddCommon(parent *model.Timestamp, key string, value interface{}, ts *model.Timestamp) {
	if parentObj, ok := its.findJSONObject(parent); ok {
		parentObj.objAdd(key, value, ts)
		return
	}
	_ = errors.NewDatatypeError(errors.ErrDatatypeInvalidParent)
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

func (its *jsonObject) UpdateLocalInArray(op *operations.UpdInArrayOperation) {
	if parentArray, ok := its.findJSONArray(op.C.P); ok {
		return parentArray.arrayUpdateLocal(op)
	}
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

func (its *jsonObject) getType() TypeOfDocument {
	return TypeJSONObject
}

func (its *jsonObject) makeTomb(ts *model.Timestamp) bool {
	// log.Logger.Infof("makeTomb() of jsonObject:%v", its.getValue())
	if its.jsonPrimitive.makeTomb(ts) {
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
	if _, ok := target.(jsonPrimitive); ok {
		// if !j.isTomb() {
		ret := its.removeCommon(key, ts)
		return ret
		// }
	}
	return nil
}

func (its *jsonObject) objAdd(key string, value interface{}, ts *model.Timestamp) {
	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		ja := its.createJSONArray(its, value, ts) // in jsonObject
		its.putCommonWithTimedValue(key, ja)      // in hash map
		its.putInNodeMap(ja)
	case reflect.Struct, reflect.Map:
		jo := its.createJSONObject(its, value, ts)
		its.putCommonWithTimedValue(key, jo) // in hash map
		its.putInNodeMap(jo)
	case reflect.Ptr:
		val := rt.Elem()
		log.Logger.Infof("%+v", val.Interface())
		its.objAdd(key, val.Interface(), ts) // recursively
	default:
		je := newJSONElement(its, types.ConvertToJSONSupportedValue(value), ts.NextDeliminator())
		its.putCommonWithTimedValue(key, je) // in hash map
		its.putInNodeMap(je)
	}
}

func (its *jsonObject) getChildAsJSONElement(key string) *jsonElement {
	value := its.get(key)
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
		parentTS = parent.getKey().ToString()
	}
	return fmt.Sprintf("JO(%v)[T%v|V%v]", parentTS, its.getKey().ToString(), its.hashMapSnapshot.String())
}

func (its *jsonObject) GetAsJSON() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v != nil {
			switch cast := v.(type) {
			case *jsonObject:
				if !cast.isTomb() {
					m[k] = cast.GetAsJSON()
				}

			case *jsonElement:
				if !cast.isTomb() {
					m[k] = v.getValue()
				}
			case *jsonArray:
				if !cast.isTomb() {
					m[k] = cast.GetAsJSON()
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
	jsonPrimitive
	*listSnapshot
}

func newJSONArray(parent jsonPrimitive, ts *model.Timestamp) *jsonArray {
	return &jsonArray{
		jsonPrimitive: &jsonPrimitiveImpl{
			parent: parent,
			root:   parent.getRoot(),
			K:      ts,
			P:      ts,
		},
		listSnapshot: newListSnapshot(),
	}
}

func (its *jsonArray) makeTomb(ts *model.Timestamp) bool {
	log.Logger.Infof("makeTomb() of jsonArray")
	if its.jsonPrimitive.makeTomb(ts) {
		return true
	}
	return false
	// node := its.head
	// for node != nil {
	// 	switch cast := node.timedValue.(type) {
	// 	case *jsonElement:
	// 		if !cast.isTomb() {
	// 			cast.makeTomb(ts)
	// 		}
	// 	case *jsonObject:
	// 		if !cast.isTomb() {
	// 			cast.makeTomb(ts)
	// 			// cast.deleteChildren(ts)
	// 		}
	// 	case *jsonArray:
	// 		if !cast.isTomb() {
	// 			cast.makeTomb(ts)
	// 			// cast.deleteChildren(ts)
	// 		}
	// 	}
	// 	node = node.next
	// }

}

func (its *jsonArray) arrayDeleteRemote(targets []*model.Timestamp, ts *model.Timestamp) {
	for _, t := range targets {
		if j, ok := its.findJSONPrimitive(t); ok {
			if !j.isTomb() {
				j.makeTomb(ts)
				its.size--
			} else { // concurrent deletes
				if j.getTime().Compare(ts) < 0 {
					j.setTime(ts)
				}
			}
		} else {
			log.Logger.Warnf("fail to find delete target: %v", t.ToString())
		}
	}
	// return nil, nil
}

func (its *jsonArray) arrayUpdateLocal(op *operations.UpdInArrayOperation) error {
	if err := its.validateRange(op.Pos, len(op.C.V)); err != nil {
		return err
	}
	node := its.findNthTarget(op.Pos + 1)
	for i := 0; i < len(op.C.V); i++ {

	}
}

func (its *jsonArray) arrayDeleteLocal(pos, numOfNodes int, ts *model.Timestamp) ([]*model.Timestamp, []interface{}, error) {
	if err := its.validateRange(pos, numOfNodes); err != nil {
		return nil, nil, err
	}
	var deletedTargets []*model.Timestamp
	var deletedValues []interface{}
	node := its.findNthTarget(pos + 1)
	for i := 0; i < numOfNodes; i++ {
		cast := node.timedValue.(jsonPrimitive)
		if cast.makeTomb(ts) {
			deletedTargets = append(deletedTargets, cast.getKey())
			deletedValues = append(deletedValues, cast)
		}
		// switch cast := node.timedValue.(type) {
		// case *jsonElement:
		// 	if !cast.isTomb() {
		// 		cast.makeTomb(ts)
		// 		deletedTargets = append(deletedTargets, cast.getKey())
		// 		deletedValues = append(deletedValues, cast.getValue())
		// 	}
		// case *jsonObject:
		// 	if !cast.isTomb() {
		// 		cast.makeTomb(ts)
		// 		deletedTargets = append(deletedTargets, cast.getKey())
		// 		deletedValues = append(deletedValues, cast.getValue())
		// 	}
		// case *jsonArray:
		// 	if !cast.isTomb() {
		// 		cast.makeTomb(ts)
		// 		deletedTargets = append(deletedTargets, cast.getKey())
		// 		deletedValues = append(deletedValues, cast.getValue())
		// 	}
		// }
		node = node.getNextLiveNode()
	}
	return deletedTargets, deletedValues, nil

}

func (its *jsonArray) arrayInsertCommon(pos int, target, ts *model.Timestamp, values ...interface{}) (*model.Timestamp, []interface{}, error) {
	var tvs []timedValue
	for _, v := range values {
		rt := reflect.ValueOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(its, v, ts)
			tvs = append(tvs, ja)
			its.putInNodeMap(ja)
		case reflect.Struct, reflect.Map:
			jo := its.createJSONObject(its, v, ts)
			tvs = append(tvs, jo)
			its.putInNodeMap(jo)
		case reflect.Ptr:
			ptrVal := rt.Elem()
			its.arrayInsertCommon(pos, target, ts, ptrVal)
		default:
			je := newJSONElement(its, types.ConvertToJSONSupportedValue(v), ts.NextDeliminator())
			tvs = append(tvs, je)
			its.putInNodeMap(je)
		}
	}
	if target == nil {
		return its.listSnapshot.insertLocalWithTimedValue(pos, tvs...)
	} else {
		its.listSnapshot.insertRemoteWithTimedValue(target, ts, tvs...)
		return nil, nil, nil
	}
}

func (its *jsonArray) getAsJSONElement(pos int) (*jsonElement, error) {
	val, err := its.listSnapshot.getTimedValue(pos)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return val.(*jsonElement), nil
}

func (its *jsonArray) getValue() types.JSONValue {
	return its.listSnapshot
}

func (its *jsonArray) getType() TypeOfDocument {
	return TypeJSONArray
}

func (its *jsonArray) setValue(v types.JSONValue) {
	panic("not used")
}

func (its *jsonArray) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getKey().ToString()
	}
	return fmt.Sprintf("JA(%v)[T%v|V%v", parentTS, its.getKey().ToString(), its.listSnapshot.String())
}

func (its *jsonArray) GetAsJSON() interface{} {
	var list []interface{}
	n := its.listSnapshot.head.getNextLiveNode()
	for n != nil {
		switch cast := n.timedValue.(type) {
		case *jsonObject:
			list = append(list, cast.GetAsJSON())
		case *jsonElement:
			list = append(list, cast.getValue())
		case *jsonArray:
			list = append(list, cast.GetAsJSON())
		}
		n = n.getNextLiveNode()
	}
	return list
}
