package snapshot

import (
	"github.com/orda-io/orda/pkg/context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/iface"
	"github.com/orda-io/orda/pkg/orda"
	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/mongodb"
	"github.com/orda-io/orda/server/schema"
)

// Manager is a struct that updates snapshot of a datatype in Orda server
type Manager struct {
	ctx           context.OrdaContext
	mongo         *mongodb.RepositoryMongo
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc
}

// NewManager returns an instance of Snapshot Manager
func NewManager(
	ctx context.OrdaContext,
	mongo *mongodb.RepositoryMongo,
	datatypeDoc *schema.DatatypeDoc,
	collectionDoc *schema.CollectionDoc,
) *Manager {
	return &Manager{
		ctx:           ctx,
		mongo:         mongo,
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

	snapshotDoc, err := its.mongo.GetLatestSnapshot(its.ctx, its.datatypeDoc.CollectionNum, its.datatypeDoc.DUID)
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
	opList, sseqList, err := its.mongo.GetOperations(its.ctx, its.datatypeDoc.DUID, lastSseq+1, constants.InfinitySseq)

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

// UpdateSnapshot updates snapshot for specified datatype
func (its *Manager) UpdateSnapshot() errors.OrdaError {
	datatype, lastSseq, err := its.GetLatestDatatype()
	if err != nil {
		return err
	}

	meta, snap, err := datatype.GetMetaAndSnapshot()
	if err != nil {
		return err
	}

	if err := its.mongo.InsertSnapshot(its.ctx, its.collectionDoc.Num, its.datatypeDoc.DUID, lastSseq, meta, snap); err != nil {
		return err
	}

	data := datatype.ToJSON()
	//its.ctx.L().Infof("final snapshot: %+v", data)
	if err := its.mongo.InsertRealSnapshot(its.ctx, its.collectionDoc.Name, its.datatypeDoc.Key, data, lastSseq); err != nil {
		return err
	}
	//its.ctx.L().Infof("update snapshot and real snapshot %+v", data)
	return nil
}
