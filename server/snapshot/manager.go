package snapshot

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/orda"

	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/managers"
	"github.com/orda-io/orda/server/schema"
)

// Manager is a struct that updates snapshot of a datatype in Orda server
type Manager struct {
	ctx           context.OrdaContext
	managers      *managers.Managers
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc
}

// NewManager returns an instance of Snapshot Manager
func NewManager(
	ctx context.OrdaContext,
	managers *managers.Managers,
	datatypeDoc *schema.DatatypeDoc,
	collectionDoc *schema.CollectionDoc,
) *Manager {
	return &Manager{
		ctx:           ctx,
		managers:      managers,
		datatypeDoc:   datatypeDoc,
		collectionDoc: collectionDoc,
	}
}

func (its *Manager) GetLatestDatatype() (iface.Datatype, uint64, errors.OrdaError) {
	var lastSseq uint64 = 0
	client := orda.NewClient(orda.NewLocalClientConfig(its.collectionDoc.Name), "orda-server")
	datatype := client.CreateDatatype(its.datatypeDoc.Key, its.datatypeDoc.GetType(), nil).(iface.Datatype)
	datatype.SetLogger(its.ctx.L())
	if its.datatypeDoc.DUID == "" {
		return datatype, lastSseq, nil
	} else {
		datatype.SetDUID(its.datatypeDoc.DUID)
	}

	snapshotDoc, err := its.managers.Mongo.GetLatestSnapshot(its.ctx, its.datatypeDoc.CollectionNum, its.datatypeDoc.DUID)
	if err != nil {
		return nil, 0, err
	}
	if snapshotDoc != nil {
		lastSseq = snapshotDoc.Sseq
		if err = datatype.SetMetaAndSnapshot([]byte(snapshotDoc.Meta), snapshotDoc.Snapshot); err != nil {
			return nil, 0, err
		}
		datatype.ResetWired()
	}
	opList, sseqList, err := its.managers.Mongo.GetOperations(its.ctx, its.datatypeDoc.DUID, lastSseq+1, constants.InfinitySseq)
	if err != nil {
		return nil, 0, err
	}

	if len(sseqList) <= 0 {
		return datatype, lastSseq, nil
	}

	its.ctx.L().Infof("apply %d operations: %+v", len(opList), opList.ToString())
	if _, err = datatype.ReceiveRemoteModelOperations(opList, false); err != nil {
		// TODO: should fix corruption
		return nil, 0, err
	}
	lastSseq = sseqList[len(sseqList)-1]
	return datatype, lastSseq, nil
}

func (its *Manager) getLockKey() string {
	return fmt.Sprintf("US:%d:%s", its.collectionDoc.Num, its.datatypeDoc.Key)
}

// UpdateSnapshot updates snapshot for specified datatype
func (its *Manager) UpdateSnapshot() errors.OrdaError {
	lock := its.managers.GetLock(its.ctx, its.getLockKey())
	if !lock.TryLock() {
		return errors.ServerUpdateSnapshot.New(its.ctx.L(), "try lock failure")
	}
	defer lock.Unlock()
	its.ctx.L().Infof("BEGIN UPD_SNAP: '%v'", its.datatypeDoc.Key)
	datatype, lastSseq, err := its.GetLatestDatatype()
	if err != nil {
		return err
	}

	meta, snap, err := datatype.GetMetaAndSnapshot()
	if err != nil {
		return err
	}

	if err := its.managers.Mongo.InsertSnapshot(its.ctx, its.collectionDoc.Num, its.datatypeDoc.DUID, lastSseq, meta, snap); err != nil {
		return err
	}

	data := datatype.ToJSON()

	if err := its.managers.Mongo.InsertRealSnapshot(its.ctx, its.collectionDoc.Name, its.datatypeDoc.Key, data, lastSseq); err != nil {
		return err
	}
	its.ctx.L().Infof("FINISH UPD_SNAP: '%v': %d", its.datatypeDoc.Key, lastSseq)
	return nil
}
