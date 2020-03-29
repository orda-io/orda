package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/internal/datatypes"
	"github.com/knowhunger/ortoo/ortoo/internal/types"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"strings"
)

type List interface {
	Datatype
	ListInTxn
}

type ListInTxn interface {
	Insert(pos int32, value ...interface{}) (interface{}, error)
	Get(pos int32) interface{}
	Delete(pos int32) interface{}
}

func newList(key string, cuid model.CUID, wire datatypes.Wire, handlers *Handlers) List {
	list := &list{
		datatype: &datatype{
			FinalDatatype: &datatypes.FinalDatatype{},
			handlers:      handlers,
		},
		snapshot: newListSnapshot(),
	}
	list.Initialize(key, model.TypeOfDatatype_LIST, cuid, wire, list.snapshot, list)
	return list
}

type list struct {
	*datatype
	snapshot *listSnapshot
}

func (its *list) GetAsJSON() (string, error) {
	panic("implement me")
}

func (its *list) ExecuteRemote(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:

	case *operations.InsertOperation:
		return its.snapshot.insertRemote(cast.C.Target.Hash(), cast.ID.GetTimestamp(), cast.C.Values...)
	case *operations.DeleteOperation:
	}
	panic("implement me")
}

func (its *list) GetSnapshot() model.Snapshot {
	panic("implement me")
}

func (its *list) GetMetaAndSnapshot() ([]byte, model.Snapshot, error) {
	panic("implement me")
}

func (its *list) SetMetaAndSnapshot(meta []byte, snapshot string) error {
	panic("implement me")
}

func (its *list) ExecuteLocal(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.InsertOperation:
		// cast.C.Target
		target, ret, err := its.snapshot.insertLocal(cast.Pos, cast.ID.GetTimestamp(), cast.C.Values...)
		if err != nil {

		}
		cast.C.Target = target
		return ret, err
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *list) Insert(pos int32, values ...interface{}) (interface{}, error) {
	var jsonValues []interface{}
	for _, val := range values {
		jsonValues = append(jsonValues, types.ConvertToJSONSupportedType(val))
	}

	op := operations.NewInsertOperation(pos, jsonValues...)
	return its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
}

func (its *list) Get(pos int32) interface{} {
	return nil
}

func (its *list) Delete(pos int32) interface{} {
	return nil
}

// ////////////////////////////////////////////////////////////////
//  listSnapshot
// ////////////////////////////////////////////////////////////////

type node struct {
	V    types.JSONType
	T    *model.Timestamp
	P    *model.Timestamp
	next *node
	prev *node
}

func (its *node) hash() string {
	return its.T.Hash()
}

func (its *node) String() string {
	var sb strings.Builder
	sb.WriteString(its.T.ToString())
	if its.P != nil {
		sb.WriteString(its.P.ToString())
	}
	if its.V == nil {
		sb.WriteString(":DELETED")
	} else {
		_, _ = fmt.Fprintf(&sb, ":%v", its.V)
	}

	return sb.String()
}

func (its *node) getNextLiveNode() *node {
	ret := its.next
	for ret != nil {
		if ret.V != nil {
			return ret
		}
		ret = ret.next
	}
	return nil
}

type listSnapshot struct {
	head *node
	size int32
	Map  map[string]*node
}

func (its *listSnapshot) CloneSnapshot() model.Snapshot {
	var cloneMap = make(map[string]*node)
	for k, v := range its.Map {
		cloneMap[k] = v
	}
	return &listSnapshot{
		head: its.head,
		size: its.size,
		Map:  cloneMap,
	}
}

func newListSnapshot() *listSnapshot {
	head := &node{
		V:    nil,
		T:    model.OldestTimestamp,
		P:    nil,
		prev: nil,
		next: nil,
	}
	m := make(map[string]*node)
	m[head.hash()] = head
	return &listSnapshot{
		head: head,
		Map:  m,
		size: 0,
	}
}

func (its *listSnapshot) insertRemote(pos string, ts *model.Timestamp, values ...interface{}) (interface{}, error) {
	if target, ok := its.Map[pos]; ok {
		currentTs := ts
		for _, val := range values {
			nextTarget := target.next
			for nextTarget != nil && nextTarget.T.Compare(ts) < 0 {
				target = target.next
				nextTarget = nextTarget.next
			}
			newNode := &node{
				V:    val,
				T:    currentTs,
				P:    nil,
				next: target.next,
				prev: target,
			}
			target.next = newNode
			its.Map[newNode.hash()] = newNode
			its.size++
			target = newNode
			currentTs = ts.GetNextDeliminator()
		}
		return nil, nil
	}
	log.Logger.Warnf("no target exists for insertRemote")
	return nil, nil
}

func (its *listSnapshot) deleteRemote(pos string, ts *model.Timestamp) (interface{}, error) {
	return nil, nil
}

func (its *listSnapshot) insertLocal(pos int32, ts *model.Timestamp, values ...interface{}) (*model.Timestamp, interface{}, error) {
	if its.size < pos { // size:0 => possible indexes{0} , s:1 => p{0, 1}
		return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "out of bound index")
	}
	target := its.findNthNode(pos)
	targetTs := target.T
	currentTs := ts
	for _, v := range values {
		newNode := &node{
			V:    v,
			T:    currentTs,
			next: target.next,
			prev: target,
		}
		target.next = newNode
		its.Map[newNode.hash()] = newNode
		its.size++
		currentTs = ts.GetNextDeliminator()
		target = newNode
	}
	return targetTs, nil, nil
}

func (its *listSnapshot) deleteLocal(pos int32, ts *model.Timestamp) (interface{}, error) {
	if its.size-1 < pos { // if size==4, 3 is ok, but 4 is not ok
		return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "out of bound index")
	}
	target := its.findNthNode(pos + 1)
	if target.V != nil {
		target.V = nil
		target.P = ts
		its.size--
	}
	return nil, nil
}

// for example: h t1 n1 n2 t2 t3 n3 t4 (h:head, n:node, t: tombstone)
// pos : 0 => h : when tombstones follows, the node before them is returned.
// pos : 1 => n1
// pos : 2 => n2
// pos : 3 => n3
func (its *listSnapshot) findNthNode(pos int32) *node {
	ret := its.head
	for i := 1; i <= int(pos); {
		ret = ret.next
		if ret.V != nil { // not tombstone
			i++
		} else { // if head or tombstone
			for ret.next != nil && ret.next.V == nil { // while next is tombstone
				ret = ret.next
			}
		}
	}
	return ret
}

func (its *listSnapshot) String() string {
	sb := strings.Builder{}
	_, _ = fmt.Fprintf(&sb, "(SIZE"+
		":%d) ", its.size)
	sb.WriteString("HEAD =>")
	n := its.head.next
	for n != nil {
		sb.WriteString(n.String())
		n = n.next
		if n != nil {
			sb.WriteString(" => ")
		}
	}
	return sb.String()
}

func (its *listSnapshot) GetAsJSON() (string, error) {
	return "", nil
}
