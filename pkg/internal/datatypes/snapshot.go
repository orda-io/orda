package datatypes

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
)

type SnapshotDatatype struct {
	iface.Snapshot
}

func (its *SnapshotDatatype) GetAsJSON() interface{} {
	return its.Snapshot.GetAsJSONCompatible()
}

func (its *SnapshotDatatype) ApplySnapshotOperation(
	sc iface.SnapshotContent,
	newSnap iface.Snapshot,
) errors.OrtooError {
	if err := json.Unmarshal([]byte(sc.GetS()), newSnap); err != nil {
		return errors.DatatypeSnapshot.New(its.GetBase().GetLogger(), err.Error())
	}
	its.Snapshot = newSnap
	return nil
}

func (its *SnapshotDatatype) SetSnapshot(snapshot iface.Snapshot) {
	its.Snapshot = snapshot
}

func (its *SnapshotDatatype) GetSnapshot() iface.Snapshot {
	return its.Snapshot
}

func (its *SnapshotDatatype) GetMetaAndSnapshot() ([]byte, []byte, errors.OrtooError) {
	meta, oErr := its.GetBase().GetMeta()
	if oErr != nil {
		return nil, nil, oErr
	}
	snap, err := json.Marshal(its.GetSnapshot())
	if err != nil {
		return nil, nil, errors.DatatypeMarshal.New(its.GetBase().GetLogger(), err.Error())
	}
	return meta, snap, nil
}

func (its *SnapshotDatatype) SetMetaAndSnapshot(meta []byte, snap []byte) errors.OrtooError {
	if err := its.GetBase().SetMeta(meta); err != nil {
		return err
	}
	if err := json.Unmarshal(snap, its.GetSnapshot()); err != nil {
		return errors.DatatypeMarshal.New(its.GetBase().GetLogger(), err.Error())
	}
	return nil
}
