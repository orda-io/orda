package ortoo

import (
	"encoding/json"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/internal/datatypes"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
	"github.com/knowhunger/ortoo/pkg/types"
	"strings"
)

// List is an Ortoo datatype which provides the list interfaces.
type List interface {
	Datatype
	ListInTxn
	DoTransaction(tag string, txnFunc func(listTxn ListInTxn) error) error
}

// ListInTxn is an Ortoo datatype which provides the list interfaces in a transaction.
type ListInTxn interface {
	InsertMany(pos int, value ...interface{}) (interface{}, errors.OrtooError)
	Get(pos int) (interface{}, errors.OrtooError)
	GetMany(pos int, numOfNodes int) ([]interface{}, errors.OrtooError)
	Delete(pos int) (interface{}, errors.OrtooError)
	DeleteMany(pos int, numOfNodes int) ([]interface{}, errors.OrtooError)
	Update(pos int, values ...interface{}) ([]interface{}, errors.OrtooError)
	Size() int
}

func newList(base *datatypes.BaseDatatype, wire iface.Wire, handlers *Handlers) List {
	list := &list{
		datatype: &datatype{
			ManageableDatatype: &datatypes.ManageableDatatype{},
			handlers:           handlers,
		},
		SnapshotDatatype: &datatypes.SnapshotDatatype{
			Snapshot: newListSnapshot(base),
		},
	}
	list.Initialize(base, wire, list.GetSnapshot(), list)
	return list
}

type list struct {
	*datatype
	*datatypes.SnapshotDatatype
}

func (its *list) DoTransaction(tag string, txnFunc func(list ListInTxn) error) error {
	return its.ManageableDatatype.DoTransaction(tag, func(txnCtx *datatypes.TransactionContext) error {
		clone := &list{
			datatype: &datatype{
				ManageableDatatype: &datatypes.ManageableDatatype{
					TransactionDatatype: its.ManageableDatatype.TransactionDatatype,
					TransactionCtx:      txnCtx,
				},
				handlers: its.handlers,
			},
			SnapshotDatatype: its.SnapshotDatatype,
		}
		return txnFunc(clone)
	})
}

func (its *list) GetAsJSON() interface{} {
	return struct {
		List []interface{}
	}{
		List: its.snapshot().GetAsJSONCompatible().([]interface{}),
	}
}

func (its *list) ExecuteLocal(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.InsertOperation:
		target, ret := its.snapshot().insertLocal(cast.Pos, cast.GetTimestamp(), cast.C.V...)
		cast.C.T = target
		return ret, nil
	case *operations.DeleteOperation:
		delTargets, _, delValues := its.snapshot().deleteLocal(cast.Pos, cast.NumOfNodes, cast.GetTimestamp())
		cast.C.T = delTargets
		return delValues, nil
	case *operations.UpdateOperation:
		uptTargets, uptValues, err := its.snapshot().updateLocal(cast.Pos, cast.GetTimestamp(), cast.C.V)
		if err != nil {
			return nil, err
		}
		cast.C.T = uptTargets
		return uptValues, nil
	}
	return nil, errors.DatatypeIllegalParameters.New(its.Logger, op)
}

func (its *list) ExecuteRemote(op interface{}) (interface{}, errors.OrtooError) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:
		err := its.ApplySnapshotOperation(cast.GetContent(), newListSnapshot(its.BaseDatatype))
		return nil, err
	case *operations.InsertOperation:
		return nil, its.snapshot().insertRemote(cast.C.T, cast.ID.GetTimestamp(), cast.C.V...)
	case *operations.DeleteOperation:
		return its.snapshot().deleteRemote(cast.C.T, cast.ID.GetTimestamp())
	case *operations.UpdateOperation:
		ret, _ := its.snapshot().updateRemote(cast.C.T, cast.C.V, cast.ID.GetTimestamp())
		return ret, nil
	}
	return nil, errors.DatatypeIllegalParameters.New(its.Logger, op)
}

func (its *list) Size() int {
	return its.snapshot().Size()
}

func (its *list) snapshot() *listSnapshot {
	return its.GetSnapshot().(*listSnapshot)
}

func (its *list) InsertMany(pos int, values ...interface{}) (interface{}, errors.OrtooError) {
	if err := its.snapshot().validateInsertPosition(pos); err != nil {
		return nil, err
	}
	jsonValues, err2 := types.ConvertValueList(values)
	if err2 != nil {
		return nil, errors.DatatypeIllegalParameters.New(its.Logger, err2.Error())
	}
	op := operations.NewInsertOperation(pos, jsonValues)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (its *list) Update(pos int, values ...interface{}) ([]interface{}, errors.OrtooError) {
	if err := its.snapshot().validateGetRange(pos, len(values)); err != nil {
		return nil, err
	}
	jsonValues, err2 := types.ConvertValueList(values)
	if err2 != nil {
		return nil, errors.DatatypeIllegalParameters.New(its.Logger, err2.Error())
	}
	op := operations.NewUpdateOperation(pos, jsonValues)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return ret.([]interface{}), nil
}

// Delete deletes one orderedType at index pos.
func (its *list) Delete(pos int) (interface{}, errors.OrtooError) {
	ret, err := its.DeleteMany(pos, 1)
	return ret[0], err
}

// DeleteMany deletes the nodes at index pos in sequence.
func (its *list) DeleteMany(pos int, numOfNode int) ([]interface{}, errors.OrtooError) {
	if err := its.snapshot().validateGetRange(pos, numOfNode); err != nil {
		return nil, err
	}
	op := operations.NewDeleteOperation(pos, numOfNode)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return nil, nil
	}

	return types.ToInterfaceArray(ret.([]types.JSONValue)), nil
}

func (its *list) Get(pos int) (interface{}, errors.OrtooError) {
	if err := its.snapshot().validateGetPosition(pos); err != nil {
		return nil, err
	}
	return its.snapshot().findValue(pos), nil
}

func (its *list) GetMany(pos int, numOfNodes int) ([]interface{}, errors.OrtooError) {
	if err := its.snapshot().validateGetRange(pos, numOfNodes); err != nil {
		return nil, err
	}
	return its.snapshot().findManyValues(pos, numOfNodes), nil
}

// ////////////////////////////////////////////////////////////////
//  listSnapshot
// ////////////////////////////////////////////////////////////////

type listSnapshot struct {
	base
	head orderedType
	size int
	Map  map[string]orderedType
}

func (its *listSnapshot) CloneSnapshot() iface.Snapshot {
	var cloneMap = make(map[string]orderedType)
	for k, v := range its.Map {
		cloneMap[k] = v
	}
	return &listSnapshot{
		base: its.base,
		head: its.head,
		size: its.size,
		Map:  cloneMap,
	}
}

func newListSnapshot(base iface.BaseDatatype) *listSnapshot {
	head := newHead()
	m := make(map[string]orderedType)
	m[head.hash()] = head
	return &listSnapshot{
		base: base,
		head: head,
		Map:  m,
		size: 0,
	}
}

func (its *listSnapshot) insertRemote(
	pos *model.Timestamp,
	ts *model.Timestamp,
	values ...interface{},
) errors.OrtooError {
	var tts []timedType
	for _, v := range values {
		tts = append(tts, newTimedNode(v, ts.GetAndNextDelimiter()))
	}
	return its.insertRemoteWithTimedTypes(pos, ts, tts...)
}

func (its *listSnapshot) insertRemoteWithTimedTypes(
	pos *model.Timestamp,
	ts *model.Timestamp,
	tts ...timedType,
) errors.OrtooError {
	if target, ok := its.Map[pos.Hash()]; ok {
		// A -> T -> B, target: T, N: new one
		for _, tt := range tts {
			nextTarget := target.getNext()                                       // nextTarget: B
			for nextTarget != nil && nextTarget.getOrderTime().Compare(ts) > 0 { // B is newer, go to next.
				target = target.getNext()
				nextTarget = nextTarget.getNext()
			}
			newNode := &orderedNode{ // N
				timedType: tt,
				O:         tt.getTime(),
			}
			target.insertNext(newNode) // T <--> N <--> B
			its.Map[newNode.hash()] = newNode
			its.size++
			target = newNode // N => T
		}
		return nil
	}
	return errors.DatatypeNoTarget.New(its.GetLogger(), pos.Hash())
}

func (its *listSnapshot) insertLocal(
	pos int,
	ts *model.Timestamp,
	values ...interface{},
) (*model.Timestamp, []interface{}) {
	var tts []timedType
	for _, v := range values {
		tts = append(tts, newTimedNode(v, ts.GetAndNextDelimiter()))
	}
	return its.insertLocalWithTimedTypes(pos, tts...)
}

func (its *listSnapshot) insertLocalWithTimedTypes(
	pos int,
	tts ...timedType,
) (*model.Timestamp, []interface{}) {
	target := its.findOrderedTypeForInsert(pos)
	var inserted []interface{}
	targetTs := target.getOrderTime()
	for _, tt := range tts {
		newNode := &orderedNode{
			timedType: tt,
			O:         tt.getTime(),
		}
		target.insertNext(newNode)
		its.Map[newNode.hash()] = newNode
		inserted = append(inserted, tt.getValue())
		its.size++
		target = newNode
	}
	return targetTs, inserted
}

func (its *listSnapshot) updateLocal(
	pos int,
	ts *model.Timestamp,
	values []interface{},
) ([]*model.Timestamp, []interface{}, errors.OrtooError) {
	target := its.findOrderedType(pos)
	var updatedValues []interface{}
	var updatedTargets []*model.Timestamp
	for _, v := range values {
		updatedTargets = append(updatedTargets, target.getOrderTime())
		updatedValues = append(updatedValues, target.getValue())
		target.setValue(v)
		target.setTime(ts.GetAndNextDelimiter())
		target = target.getNextLive()
	}
	return updatedTargets, updatedValues, nil
}

func (its *listSnapshot) updateRemote(
	targets []*model.Timestamp,
	values []interface{},
	ts *model.Timestamp,
) ([]interface{}, errors.OrtooError) {
	var updated []interface{}
	errs := &errors.MultipleOrtooErrors{}
	for i, t := range targets {
		thisTS := ts.GetAndNextDelimiter()
		if node, ok := its.Map[t.Hash()]; ok {
			// tombstone is not recovered.
			if node.isTomb() {
				continue
			}
			if node.getTime() == nil || node.getTime().Compare(thisTS) < 0 {
				updated = append(updated, node.getValue())
				node.setValue(values[i])
				node.setTime(thisTS)
			}
		} else {
			_ = errs.Append(errors.DatatypeNoTarget.New(its.GetLogger(), t.ToString()))
		}
	}
	return updated, errs.Return()
}

func (its *listSnapshot) deleteLocal(
	pos int,
	numOfNodes int,
	ts *model.Timestamp,
) ([]*model.Timestamp, []timedType, []types.JSONValue) {
	target := its.findOrderedType(pos)
	var delTargets []*model.Timestamp
	var delTimedTypes []timedType
	var delValues []types.JSONValue
	for i := 0; i < numOfNodes; i++ {
		delTimedTypes = append(delTimedTypes, target.getTimedType())
		delTargets = append(delTargets, target.getOrderTime())
		delValues = append(delValues, target.getValue())
		// targets should be deleted with different timestamps because they can be inserted into Cemetery in Document
		target.makeTomb(ts.GetAndNextDelimiter())
		its.size--
		target = target.getNextLive()
	}
	return delTargets, delTimedTypes, delValues
}

func (its *listSnapshot) deleteRemote(
	targets []*model.Timestamp,
	ts *model.Timestamp,
) ([]timedType, errors.OrtooError) {
	errs := &errors.MultipleOrtooErrors{}
	var deleted []timedType
	for _, t := range targets {
		thisTS := ts.GetAndNextDelimiter()
		if node, ok := its.Map[t.Hash()]; ok {
			if !node.isTomb() { // if not tombstone
				// A node should be deleted even if it has been updated by any update operation(s).
				node.makeTomb(thisTS)
				deleted = append(deleted, node.getTimedType())
				its.size--
			} else { // if tombstone,
				if node.getTime().Compare(thisTS) < 0 {
					node.makeTomb(thisTS)
				}
			}
		} else {
			_ = errs.Append(errors.DatatypeNoTarget.New(its.GetLogger(), t.ToString()))
		}
	}
	return deleted, errs.Return()
}

// //////////////////////////////////////////////////////////////////////
// For getting / finding / retrieving
// //////////////////////////////////////////////////////////////////////

// retrieve does not validate pos
// it is assumed that always valid pos is passed.
// for example: h t1 n1 n2 t2 t3 n3 t4 (h:head, n:orderedType, t: tombstone) size==3
// pos : 0 => h : when tombstones follows, the orderedType before them is returned.
// pos : 1 => n1
// pos : 2 => n2
// pos : 3 => n3
func (its *listSnapshot) retrieve(pos int) orderedType {
	ret := its.head
	for i := 1; i <= pos; {
		ret = ret.getNext()
		if ret == nil {
			return nil
		}
		if !ret.isTomb() { // not tombstone
			i++
		} else { // if tombstone
			for ret.getNext() != nil && ret.getNext().isTomb() { // while next is tombstone
				ret = ret.getNext()
			}
		}
	}
	return ret
}

func (its *listSnapshot) validateGetRange(pos int, numOfNodes int) errors.OrtooError {
	// 1st condition: if size==4, pos==3 is ok, but 4 is not ok
	// 2nd condition: if size==4, (pos==3, numOfNodes==1) is ok, (pos==3, numOfNodes=2) is not ok.
	if pos < 0 {
		return errors.DatatypeIllegalParameters.New(its.GetLogger(), "negative position")
	}
	if numOfNodes < 1 {
		return errors.DatatypeIllegalParameters.New(its.GetLogger(), "numOfNodes should be more than 0")
	}
	if its.size-1 < pos || pos+numOfNodes > its.size {
		return errors.DatatypeIllegalParameters.New(its.GetLogger(), "out of bound index")
	}
	return nil
}

func (its *listSnapshot) validateInsertPosition(pos int) errors.OrtooError {
	if pos < 0 {
		return errors.DatatypeIllegalParameters.New(its.GetLogger(), "negative position")
	}
	if pos > its.size { // size:0 => pos{0} , size:1 => pos{0, 1}
		return errors.DatatypeIllegalParameters.New(its.GetLogger(), "out of bound index")
	}
	return nil
}

func (its *listSnapshot) validateGetPosition(pos int) errors.OrtooError {
	if pos < 0 {
		return errors.DatatypeIllegalParameters.New(its.GetLogger(), "negative position")
	}
	if pos >= its.size { // size: 1 => pos {0}, size: 2 => pos {0, 1}
		return errors.DatatypeIllegalParameters.New(its.GetLogger(), "out of bound index")
	}
	return nil
}

func (its *listSnapshot) findOrderedType(pos int) orderedType {
	return its.retrieve(pos + 1)
}

// findOrderedTypeForInsert finds a place related to Insert
func (its *listSnapshot) findOrderedTypeForInsert(pos int) orderedType {
	return its.retrieve(pos)
}

func (its *listSnapshot) findTimedType(pos int) timedType {
	o := its.findOrderedType(pos)
	return o.getTimedType()
}

func (its *listSnapshot) findManyTimedTypes(pos int, numOfNodes int) []timedType {
	target := its.findOrderedType(pos)
	var ret []timedType
	for i := 1; i <= numOfNodes; i++ {
		ret = append(ret, target.getTimedType())
		target = target.getNextLive()
	}
	return ret
}

func (its *listSnapshot) findValue(pos int) interface{} {
	pt := its.findTimedType(pos)
	return pt.getValue()
}

func (its *listSnapshot) findManyValues(pos int, numOfNodes int) []interface{} {
	target := its.findOrderedType(pos)
	var ret []interface{}
	for i := 1; i <= numOfNodes; i++ {
		ret = append(ret, target.getValue())
		target = target.getNextLive()
	}
	return ret
}

func (its *listSnapshot) String() string {
	sb := strings.Builder{}
	_, _ = fmt.Fprintf(&sb, "(SIZE:%d) HEAD =>", its.size)
	n := its.head.getNext()
	for n != nil {
		sb.WriteString(n.String())
		n = n.getNext()
		if n != nil {
			sb.WriteString(" => ")
		}
	}
	return sb.String()
}

func (its *listSnapshot) GetBase() iface.BaseDatatype {
	return its.base
}

func (its *listSnapshot) SetBase(base iface.BaseDatatype) {
	its.base = base
}

func (its *listSnapshot) GetAsJSONCompatible() interface{} {
	var l []interface{}
	n := its.head.getNextLive()
	for n != nil {
		l = append(l, n.getValue())
		n = n.getNextLive()
	}
	return l
}

func (its *listSnapshot) Size() int {
	return its.size
}

// ////////////////////////////////////////////////////
// For marshaling
// ////////////////////////////////////////////////////

type marshaledNode struct {
	V types.JSONValue
	T *model.Timestamp
	O *model.Timestamp
}

type marshaledList struct {
	Nodes []*marshaledNode
	Size  int
}

func (its *listSnapshot) MarshalJSON() ([]byte, error) {
	forMarshal := marshaledList{
		Size: its.size,
	}
	n := its.head.getNext()
	for n != nil {
		forMarshal.Nodes = append(forMarshal.Nodes, n.marshal())
		n = n.getNext()
	}
	return json.Marshal(forMarshal)
}

func (its *listSnapshot) UnmarshalJSON(bytes []byte) error {
	forUnmarshal := marshaledList{}
	err := json.Unmarshal(bytes, &forUnmarshal)
	if err != nil {
		return err
	}
	its.head = newHead()
	its.size = forUnmarshal.Size
	its.Map = make(map[string]orderedType)
	its.Map[its.head.hash()] = its.head

	prev := its.head
	for _, n := range forUnmarshal.Nodes {
		node := n.unmarshalAsNode()
		prev.insertNext(node)
		prev = node
		its.Map[node.getOrderTime().Hash()] = node
	}
	return nil
}

func (its *marshaledNode) unmarshalAsNode() orderedType {
	return &orderedNode{
		timedType: newTimedNode(its.V, its.T),
		O:         its.O,
		next:      nil,
		prev:      nil,
	}
}

func (its *orderedNode) marshal() *marshaledNode {
	return &marshaledNode{
		V: its.getValue(),
		T: its.getTime(),
		O: its.getOrderTime(),
	}
}

func (its *orderedNode) UnmarshalJSON(bytes []byte) error {
	panic("not supported")
}

func (its *orderedNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(its.marshal())
}
