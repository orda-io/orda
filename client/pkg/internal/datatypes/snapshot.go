package datatypes

import (
	"encoding/json"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	iface2 "github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/operations"
)

type SnapshotDatatype struct {
	*BaseDatatype
	Snapshot iface2.Snapshot
}

func (its *SnapshotDatatype) ToJSON() interface{} {
	return its.Snapshot.ToJSON()
}

func NewSnapshotDatatype(b *BaseDatatype, snap iface2.Snapshot) *SnapshotDatatype {
	return &SnapshotDatatype{
		BaseDatatype: b,
		Snapshot:     snap,
	}
}

func (its *SnapshotDatatype) ApplySnapshot(snapBody []byte) errors2.OrdaError {
	its.L().Infof("apply SnapshotOperation: %v", string(snapBody))
	if err := json.Unmarshal(snapBody, its.Snapshot); err != nil {
		return errors2.DatatypeSnapshot.New(its.L(), err.Error())
	}
	return nil
}

func (its *SnapshotDatatype) GetSnapshot() iface2.Snapshot {
	return its.Snapshot
}

func (its *SnapshotDatatype) GetMetaAndSnapshot() ([]byte, []byte, errors2.OrdaError) {
	meta, oErr := its.GetMeta()
	if oErr != nil {
		return nil, nil, oErr
	}
	snap, err := json.Marshal(its.GetSnapshot())
	if err != nil {
		return nil, nil, errors2.DatatypeMarshal.New(its.L(), err.Error())
	}
	return meta, snap, nil
}

func (its *SnapshotDatatype) SetMetaAndSnapshot(meta []byte, snap []byte) errors2.OrdaError {
	if err := its.SetMeta(meta); err != nil {
		return err
	}
	if err := json.Unmarshal(snap, its.GetSnapshot()); err != nil {
		return errors2.DatatypeMarshal.New(its.L(), err.Error())
	}
	return nil
}

func (its *SnapshotDatatype) CreateSnapshotOperation() (iface2.Operation, errors2.OrdaError) {
	snap, err := json.Marshal(its.Snapshot)
	if err != nil {
		return nil, errors2.DatatypeMarshal.New(its.L(), err.Error())
	}
	snapOp := operations.NewSnapshotOperation(its.TypeOf, snap)
	return snapOp, nil
}
