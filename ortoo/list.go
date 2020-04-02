package ortoo

import (
	"encoding/json"
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
	Insert(pos int, value ...interface{}) (interface{}, error)
	Get(pos int) interface{}
	Delete(pos int) (interface{}, error)
	DeleteMany(pos int, numOfNode int) ([]interface{}, error)
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
	return its.snapshot.GetAsJSON()
}

func (its *list) ExecuteRemote(op interface{}) (interface{}, error) {
	switch cast := op.(type) {
	case *operations.SnapshotOperation:

	case *operations.InsertOperation:
		return its.snapshot.insertRemote(cast.C.Target.Hash(), cast.ID.GetTimestamp(), cast.C.Values...)
	case *operations.DeleteOperation:
		return its.snapshot.deleteRemote(cast.C.Targets, cast.ID.GetTimestamp())
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
		target, ret, err := its.snapshot.insertLocal(cast.Pos, cast.ID.GetTimestamp(), cast.C.Values...)
		if err != nil {
			return nil, err
		}
		cast.C.Target = target
		return ret, nil
	case *operations.DeleteOperation:
		deletedTimestamps, deletedValues, err := its.snapshot.deleteLocal(cast.Pos, int(cast.Pos), cast.ID.GetTimestamp())
		if err != nil {
			return nil, err
		}
		cast.C.Targets = deletedTimestamps
		return deletedValues, nil
	}
	return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, op)
}

func (its *list) Insert(pos int, values ...interface{}) (interface{}, error) {
	if len(values) < 1 {
		return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "at least one value should be inserted")
	}
	var jsonValues []interface{}
	for _, val := range values {
		jsonValues = append(jsonValues, types.ConvertToJSONSupportedType(val))
	}

	op := operations.NewInsertOperation(pos, jsonValues...)
	return its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
}

func (its *list) Get(pos int) interface{} {
	return nil
}

// Delete deletes one node at index pos.
func (its *list) Delete(pos int) (interface{}, error) {
	ret, err := its.DeleteMany(pos, 1)
	if err != nil {
		return nil, err
	}
	return ret[0], err
}

// DeleteMany deletes the nodes at index pos in sequence.
func (its *list) DeleteMany(pos int, numOfNode int) ([]interface{}, error) {
	op := operations.NewDeleteOperation(pos, numOfNode)
	ret, err := its.ExecuteOperationWithTransaction(its.TransactionCtx, op, true)
	if err != nil {
		return nil, err
	}
	return ret.([]interface{}), nil
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

func (its *listSnapshot) insertLocal(pos int32, ts *model.Timestamp, values ...interface{}) (*model.Timestamp, interface{}, error) {
	if its.size < pos { // size:0 => possible indexes{0} , s:1 => p{0, 1}
		return nil, nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "out of bound index")
	}
	var inserted []interface{}
	target := its.findNthTarget(pos)
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
		inserted = append(inserted, v)
		its.size++
		currentTs = ts.GetNextDeliminator()
		target = newNode
	}
	return targetTs, inserted, nil
}

func (its *listSnapshot) isTombstone(n *node) bool {
	if n.V == nil && n.P != nil {
		return true
	}
	return false
}

func (its *listSnapshot) updateLocal(pos int32, ts *model.Timestamp, values ...interface{}) ([]*model.Timestamp, interface{}, error) {
	if err := its.validateRange(pos, len(values)); err != nil {
		return nil, nil, err
	}
	var updatedValues []interface{}
	var updatedTargets []*model.Timestamp
	target := its.findNthTarget(pos + 1)
	for _, v := range values {
		target.V = append(updatedValues, target.V)
		updatedTargets = append(updatedTargets, target.T)
		target.V = v
		target.P = ts
		target = target.getNextLiveNode()
	}
	return updatedTargets, updatedValues, nil
}

func (its *listSnapshot) updateRemote(targets []*model.Timestamp, values []interface{}, ts *model.Timestamp) {
	for i, t := range targets {
		if node, ok := its.Map[t.Hash()]; ok {
			if its.isTombstone(node) {
				continue
			}
			if node.T.Compare(ts) < 0 {
				node.V = values[i]
				node.P = ts
			}
		}
	}
}

func (its *listSnapshot) deleteRemote(targets []*model.Timestamp, ts *model.Timestamp) (interface{}, error) {
	for _, t := range targets {
		if node, ok := its.Map[t.Hash()]; ok {
			if !its.isTombstone(node) {
				node.V = nil
				its.size--
				node.P = ts
			} else { // concurrent deletes
				if node.P.Compare(ts) < 0 {
					node.P = ts
				}
			}
		} else {
			log.Logger.Warnf("fail to find delete target: %v", t.ToString())
		}
	}
	return nil, nil
}

func (its *listSnapshot) validateRange(pos int32, numOfNodes int) error {
	// 1st condition: if size==4, pos==3 is ok, but 4 is not ok
	// 2nd condition: if size==4, (pos==3, numOfNodes==1) is ok, (pos==3, numOfNodes=2) is not ok.
	if its.size-1 < pos || pos+int32(numOfNodes) > its.size {
		return errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "out of bound index")
	}
	return nil
}

func (its *listSnapshot) deleteLocal(pos int32, numOfNodes int, ts *model.Timestamp) ([]*model.Timestamp, []interface{}, error) {
	if err := its.validateRange(pos, numOfNodes); err != nil {
		return nil, nil, err
	}
	var deletedTargets []*model.Timestamp
	var deletedValues []interface{}
	target := its.findNthTarget(pos + 1) // no head, but live node
	for i := 0; i < numOfNodes; i++ {
		deletedValues = append(deletedValues, target.V)
		deletedTargets = append(deletedTargets, target.T)
		target.V = nil
		target.P = ts
		its.size--

		target = target.getNextLiveNode()
	}
	return deletedTargets, deletedValues, nil
}

// for example: h t1 n1 n2 t2 t3 n3 t4 (h:head, n:node, t: tombstone) size==3
// pos : 0 => h : when tombstones follows, the node before them is returned.
// pos : 1 => n1
// pos : 2 => n2
// pos : 3 => n3
func (its *listSnapshot) findNthTarget(pos int32) *node {
	ret := its.head
	for i := 1; i <= int(pos); {
		ret = ret.next
		if !its.isTombstone(ret) { // not tombstone
			i++
		} else { // if tombstone
			for ret.next != nil && its.isTombstone(ret.next) { // while next is tombstone
				ret = ret.next
			}
		}
	}
	return ret
}

func (its *listSnapshot) get(pos int32) (interface{}, error) {
	// size == 3, pos can be 0, 1, 2
	if its.size <= pos {
		return nil, errors.NewDatatypeError(errors.ErrDatatypeIllegalOperation, "out of bound index")
	}
	target := its.findNthTarget(pos)
	liveNode := target.getNextLiveNode()
	return liveNode.V, nil
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
	var l []interface{}
	n := its.head.getNextLiveNode()
	for n != nil {
		l = append(l, n.V)
		n = n.getNextLiveNode()
	}
	j, err := json.Marshal(l)
	if err != nil {
		return "", errors.NewDatatypeError(errors.ErrDatatypeSnapshot, err.Error())
	}
	return string(j), nil
}
