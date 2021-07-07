package datatypes

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/operations"
)

type SnapshotDatatype struct {
	*BaseDatatype
	Snapshot iface.Snapshot
}

func (its *SnapshotDatatype) ToJSON() interface{} {
	return its.Snapshot.ToJSON()
}

func NewSnapshotDatatype(b *BaseDatatype, snap iface.Snapshot) *SnapshotDatatype {
	return &SnapshotDatatype{
		BaseDatatype: b,
		Snapshot:     snap,
	}
}

func (its *SnapshotDatatype) ApplySnapshot(snapBody []byte) errors.OrtooError {
	its.L().Infof("apply SnapshotOperation: %v", string(snapBody))
	if err := json.Unmarshal(snapBody, its.Snapshot); err != nil {
		return errors.DatatypeSnapshot.New(its.L(), err.Error())
	}
	return nil
}

func (its *SnapshotDatatype) GetSnapshot() iface.Snapshot {
	return its.Snapshot
}

func (its *SnapshotDatatype) GetMetaAndSnapshot() ([]byte, []byte, errors.OrtooError) {
	meta, oErr := its.GetMeta()
	if oErr != nil {
		return nil, nil, oErr
	}
	snap, err := json.Marshal(its.GetSnapshot())
	if err != nil {
		return nil, nil, errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	return meta, snap, nil
}

func (its *SnapshotDatatype) SetMetaAndSnapshot(meta []byte, snap []byte) errors.OrtooError {
	if err := its.SetMeta(meta); err != nil {
		return err
	}
	if err := json.Unmarshal(snap, its.GetSnapshot()); err != nil {
		return errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	return nil
}

func (its *SnapshotDatatype) CreateSnapshotOperation() (iface.Operation, errors.OrtooError) {
	snap, err := json.Marshal(its.Snapshot)
	if err != nil {
		return nil, errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	snapOp := operations.NewSnapshotOperation(snap)
	return snapOp, nil
}
