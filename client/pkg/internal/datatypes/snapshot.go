package datatypes

import (
	"encoding/json"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/operations"
)

// SnapshotDatatype is responsible for managing the snapshot
type SnapshotDatatype struct {
	*BaseDatatype
	Snapshot iface.Snapshot
}

// ToJSON returns the snapshot in the format of JSON
func (its *SnapshotDatatype) ToJSON() interface{} {
	return its.Snapshot.ToJSON()
}

// NewSnapshotDatatype creates a new SnapshotDatatype
func NewSnapshotDatatype(b *BaseDatatype, snap iface.Snapshot) *SnapshotDatatype {
	return &SnapshotDatatype{
		BaseDatatype: b,
		Snapshot:     snap,
	}
}

// ApplySnapshot applies for a snapshot
func (its *SnapshotDatatype) ApplySnapshot(snapBody []byte) errors.OrdaError {
	its.L().Infof("apply SnapshotOperation: %v", string(snapBody))
	if err := json.Unmarshal(snapBody, its.Snapshot); err != nil {
		return errors.DatatypeSnapshot.New(its.L(), err.Error())
	}
	return nil
}

// GetSnapshot returns the snapshot
func (its *SnapshotDatatype) GetSnapshot() iface.Snapshot {
	return its.Snapshot
}

// GetMetaAndSnapshot returns the meta and snapshot
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

// SetMetaAndSnapshot sets the meta and snapshot
func (its *SnapshotDatatype) SetMetaAndSnapshot(meta, snap []byte) errors.OrdaError {
	if err := its.SetMeta(meta); err != nil {
		return err
	}
	if err := json.Unmarshal(snap, its.GetSnapshot()); err != nil {
		return errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	return nil
}

// CreateSnapshotOperation returns the SnapshotOperation from the snapshot and meta
func (its *SnapshotDatatype) CreateSnapshotOperation() (iface.Operation, errors.OrdaError) {
	snap, err := json.Marshal(its.Snapshot)
	if err != nil {
		return nil, errors.DatatypeMarshal.New(its.L(), err.Error())
	}
	snapOp := operations.NewSnapshotOperation(its.TypeOf, snap)
	return snapOp, nil
}
