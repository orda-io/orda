package datatypes

import (
	"encoding/json"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/operations"
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

func (its *SnapshotDatatype) ApplySnapshot(snapBody []byte) errors.OrdaError {
	its.L().Infof("apply SnapshotOperation: %v", string(snapBody))
	if err := json.Unmarshal(snapBody, its.Snapshot); err != nil {
		return errors.DatatypeSnapshot.New(its.L(), err.Error())
	}
	return nil
}

func (its *SnapshotDatatype) GetSnapshot() iface.Snapshot {
	return its.Snapshot
}

func (its *SnapshotDatatype) GetMetaAndSnapshot() ([]byte, []byte, errors.OrdaError) {
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

func (its *SnapshotDatatype) SetMetaAndSnapshot(meta []byte, snap []byte) errors.OrdaError {
	if err := its.SetMeta(meta); err != nil {
		return err
	}
	if err := json.Unmarshal(snap, its.GetSnapshot()); err != nil {
		return errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	return nil
}

func (its *SnapshotDatatype) CreateSnapshotOperation() (iface.Operation, errors.OrdaError) {
	snap, err := json.Marshal(its.Snapshot)
	if err != nil {
		return nil, errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	snapOp := operations.NewSnapshotOperation(its.TypeOf, snap)
	return snapOp, nil
}
