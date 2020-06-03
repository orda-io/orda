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
	Add(key string, value interface{}) (interface{}, error)
	GetByKey(key string) (Document, error)
	GetByIndex(pos int) (Document, error)
	GetDocumentType() TypeOfDocument
}

func newDocument(key string, cuid types.CUID, wire iface.Wire, handlers *Handlers) Document {
	doc := &document{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		root:     model.OldestTimestamp,
		typeof:   TypeJSONObject,
		snapshot: newJSONObject(nil, model.OldestTimestamp),
	}
	doc.Initialize(key, model.TypeOfDatatype_DOCUMENT, cuid, wire, doc.snapshot, doc)
	return doc
}

type document struct {
	*datatype
	root     *model.Timestamp
	typeof   TypeOfDocument
	snapshot *jsonObject
}

func (its *document) Add(key string, value interface{}) (interface{}, error) {
	op := operations.NewAddOperation(its.root, key, value)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (its *document) ExecuteLocal(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.AddOperation:
		its.snapshot.addCommon(cast.C.P, cast.C.K, cast.C.V, cast.ID.GetTimestamp())
		return nil, nil
	case *operations.CutOperation:
	case *operations.SetOperation:
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
	case *operations.AddOperation:
		its.snapshot.addCommon(cast.C.P, cast.C.K, cast.C.V, cast.ID.GetTimestamp())
		return nil, nil
	case *operations.CutOperation:
	case *operations.SetOperation:
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *document) GetByKey(key string) (Document, error) {
	if its.typeof == TypeJSONObject {
		child := its.snapshot.get(key).(jsonPrimitive)
		if child == nil {
			return nil, errors.NewDatatypeError(errors.ErrDatatypeNotExistChildDocument)
		}
		return &document{
			datatype: its.datatype,
			root:     child.getTime(),
			typeof:   child.getType(),
			snapshot: its.snapshot,
		}, nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeInvalidParentDocumentType)
}

func (its *document) GetByIndex(pos int) (Document, error) {
	return nil, nil
}

func (its *document) GetDocumentType() TypeOfDocument {
	return its.typeof
}

func (its *document) GetAsJSON() interface{} {
	panic("implement me")
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
// every jsonXXXX hassh

//  jsonPrimitive

type jsonPrimitive interface {
	timedValue
	getType() TypeOfDocument
	getRoot() *jsonRoot
	setRoot(r *jsonObject)
	getParent() jsonPrimitive
	putInNodeMap(tv timedValue)
	getParentAsJSONObject() *jsonObject
	createJSONObject(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonObject
	createJSONArray(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonArray
}

type jsonRoot struct {
	nodeMap map[string]timedValue
	root    *jsonObject
}

type jsonPrimitiveImpl struct {
	parent jsonPrimitive
	root   *jsonRoot
}

func (its *jsonPrimitiveImpl) putInNodeMap(tv timedValue) {
	its.getRoot().nodeMap[tv.getTime().ToString()] = tv
}

func (its *jsonPrimitiveImpl) getValue() types.JSONValue {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) setValue(v types.JSONValue) {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) getTime() *model.Timestamp {
	panic("should be overridden")
}

func (its *jsonPrimitiveImpl) getRoot() *jsonRoot {
	return its.root
}

func (its *jsonPrimitiveImpl) setRoot(r *jsonObject) {
	its.root.root = r
	its.root.nodeMap[r.T.ToString()] = r
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
			its.putInNodeMap(v)
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

		}
	}
	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		switch field.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(jo, field.Interface(), ts)
			jo.putCommonWithTimedValue(fields.Field(i).Name, ja)
			its.putInNodeMap(ja)
		case reflect.Struct:
			childJO := its.createJSONObject(jo, field.Interface(), ts)
			jo.putCommonWithTimedValue(fields.Field(i).Name, childJO)
			its.putInNodeMap(childJO)
		case reflect.Ptr:
			val := field.Elem()
			its.createJSONObject(parent, val.Interface(), ts)
		default:
			element := newJSONElement(jo, types.ConvertToJSONSupportedValue(field.Interface()), ts.NextDeliminator())
			jo.putCommonWithTimedValue(fields.Field(i).Name, element)
			its.putInNodeMap(element)
		}
	}
	return jo
}

func (its *jsonPrimitiveImpl) addValueToJSONObject(jo jsonObject, key string, field reflect.Value) {
	switch field.Kind() {
	case reflect.Slice, reflect.Array:
		ja := its.createJSONArray(jo, field.Interface(), ts)
		jo.putCommonWithTimedValue(fields.Field(i).Name, ja)
		its.putInNodeMap(ja)
	case reflect.Struct:
		childJO := its.createJSONObject(jo, field.Interface(), ts)
		jo.putCommonWithTimedValue(fields.Field(i).Name, childJO)
		its.putInNodeMap(childJO)
	case reflect.Ptr:
		val := field.Elem()
		its.createJSONObject(parent, val.Interface(), ts)
	default:
		element := newJSONElement(jo, types.ConvertToJSONSupportedValue(field.Interface()), ts.NextDeliminator())
		jo.putCommonWithTimedValue(fields.Field(i).Name, element)
		its.putInNodeMap(element)
	}
}

//  jsonElement

type jsonElement struct {
	jsonPrimitive
	V types.JSONValue
	T *model.Timestamp
}

func newJSONElement(parent jsonPrimitive, value interface{}, ts *model.Timestamp) *jsonElement {
	return &jsonElement{
		jsonPrimitive: &jsonPrimitiveImpl{
			parent: parent,
		},
		V: value,
		T: ts,
	}
}

func (its *jsonElement) getValue() types.JSONValue {
	return its.V
}

func (its *jsonElement) getTime() *model.Timestamp {
	return its.T
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
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JE(P%v)[T%v|%v]", parentTS, its.getTime().ToString(), its.V)
}

//  jsonObject

type jsonObject struct {
	jsonPrimitive
	T *model.Timestamp
	*hashMapSnapshot
}

func newJSONObject(parent jsonPrimitive, ts *model.Timestamp) *jsonObject {
	var root *jsonRoot
	if parent == nil {
		root = &jsonRoot{
			nodeMap: make(map[string]timedValue),
			root:    nil,
		}
	} else {
		root = parent.getRoot()
	}
	obj := &jsonObject{
		T: ts,
		jsonPrimitive: &jsonPrimitiveImpl{
			root:   root,
			parent: parent,
		},
		hashMapSnapshot: newHashMapSnapshot(),
	}
	obj.jsonPrimitive.setRoot(obj)
	return obj
}

// func (its *jsonObject) GetJsonByTimestamp()

func (its *jsonObject) CloneSnapshot() iface.Snapshot {
	// TODO: implement CloneSnapshot()
	return &jsonObject{}
}

func (its *jsonObject) addCommon(parent *model.Timestamp, key string, value interface{}, ts *model.Timestamp) {
	parentObj := its.getRoot().nodeMap[parent.ToString()].(*jsonObject)
	parentObj.put(key, value, ts)
}

func (its *jsonObject) getTime() *model.Timestamp {
	return its.T
}

func (its *jsonObject) getValue() types.JSONValue {
	return its.hashMapSnapshot
}

func (its *jsonObject) getType() TypeOfDocument {
	return TypeJSONObject
}

func (its *jsonObject) put(key string, value interface{}, ts *model.Timestamp) {
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
		its.put(key, val.Interface(), ts) // recursively
	default:
		je := newJSONElement(its, types.ConvertToJSONSupportedValue(value), ts.NextDeliminator())
		its.putCommonWithTimedValue(key, je) // in hash map
		its.putInNodeMap(je)
	}
}

func (its *jsonObject) getAsJSONElement(key string) *jsonElement {
	value := its.get(key)
	return value.(*jsonElement)
}

func (its *jsonObject) getAsJSONObject(key string) *jsonObject {
	value := its.get(key)
	return value.(*jsonObject)
}

func (its *jsonObject) getAsJSONArray(key string) *jsonArray {
	value := its.get(key)
	return value.(*jsonArray)
}

func (its *jsonObject) String() string {
	parent := its.getParent()
	parentTS := "nil"
	if parent != nil {
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JO(%v)[T%v|V%v]", parentTS, its.T.ToString(), its.hashMapSnapshot.String())
}

func (its *jsonObject) GetAsJSON() interface{} {
	m := make(map[string]interface{})
	for k, v := range its.Map {
		if v != nil {
			switch cast := v.(type) {
			case *jsonObject:
				log.Logger.Infof("%v:%v", k, cast)
				m[k] = cast.GetAsJSON()
			case *jsonElement:
				m[k] = v.getValue()
			case *jsonArray:
				m[k] = cast.GetAsJSON()
			}
		}
	}
	return m
}

//  jsonArray

type jsonArray struct {
	jsonPrimitive
	T *model.Timestamp
	*listSnapshot
}

func newJSONArray(parent jsonPrimitive, ts *model.Timestamp) *jsonArray {
	return &jsonArray{
		T: ts,
		jsonPrimitive: &jsonPrimitiveImpl{
			parent: parent,
		},
		listSnapshot: newListSnapshot(),
	}
}

func (its *jsonArray) insertLocal(pos int32, ts *model.Timestamp, values ...interface{}) {
	var tvs []timedValue
	for _, v := range values {
		rt := reflect.ValueOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Array:
			ja := its.createJSONArray(its, v, ts)
			tvs = append(tvs, ja)
		case reflect.Struct:
			jo := its.createJSONObject(its, v, ts)
			tvs = append(tvs, jo)
		case reflect.Ptr:
			ptrVal := rt.Elem()
			its.insertLocal(pos, ts, ptrVal)
		default:
			je := newJSONElement(its, types.ConvertToJSONSupportedValue(v), ts.NextDeliminator())
			tvs = append(tvs, je)
		}
	}
	its.listSnapshot.insertLocalWithTimedValue(pos, tvs...)

}

func (its *jsonArray) getAsJSONElement(pos int32) (*jsonElement, error) {
	val, err := its.listSnapshot.getTimedValue(pos)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return val.(*jsonElement), nil
}

func (its *jsonArray) getTime() *model.Timestamp {
	return its.T
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
		parentTS = parent.getTime().ToString()
	}
	return fmt.Sprintf("JA(%v)[T%v|V%v", parentTS, its.T.ToString(), its.listSnapshot.String())
}
