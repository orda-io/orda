package snapshot

import (
	"context"
	"encoding/json"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
)

// Manager is a struct that updates snapshot of a datatype in Ortoo server
type Manager struct {
	ctx           context.Context
	mongo         *mongodb.RepositoryMongo
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc
}

// NewManager returns an instance of Snapshot Manager
func NewManager(
	ctx context.Context,
	mongo *mongodb.RepositoryMongo,
	datatypeDoc *schema.DatatypeDoc,
	collectionDoc *schema.CollectionDoc) *Manager {
	return &Manager{
		ctx:           ctx,
		mongo:         mongo,
		datatypeDoc:   datatypeDoc,
		collectionDoc: collectionDoc,
	}
}

func (m *Manager) getPushPullTag() model.PushPullTag {
	return model.PushPullTag{
		CollectionName: m.collectionDoc.Name,
		Key:            m.datatypeDoc.Key,
		DUID:           m.datatypeDoc.DUID,
	}
}

// UpdateSnapshot updates snapshot for specified datatype
func (m *Manager) UpdateSnapshot() error {
	var sseq uint64 = 0
	client := ortoo.NewClient(ortoo.NewLocalClientConfig(m.collectionDoc.Name), "server")
	datatype := client.CreateDatatype(m.datatypeDoc.Key, m.datatypeDoc.GetType())
	snapshotDoc, err := m.mongo.GetLatestSnapshot(m.ctx, m.collectionDoc.Num, m.datatypeDoc.DUID)
	if err != nil {
		return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	if snapshotDoc != nil {
		sseq = snapshotDoc.Sseq
		if err := datatype.SetMetaAndSnapshot(snapshotDoc.Meta, snapshotDoc.Snapshot.(model.Snapshot)); err != nil {
			return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
		}
	}
	var transaction []model.Operation
	var remainOfTransaction uint32 = 0
	if err := m.mongo.GetOperations(m.ctx, m.datatypeDoc.DUID, sseq+1, constants.InfinitySseq, func(opDoc *schema.OperationDoc) error {
		// opOnWire := opDoc.Operation
		var opOnWire model.OperationOnWire
		if err := proto.Unmarshal(opDoc.Operation, &opOnWire); err != nil {
			return err
		}
		op := model.ToOperation(&opOnWire)
		if op.GetBase().OpType == model.TypeOfOperation_TRANSACTION {
			trxOp := op.(*model.TransactionOperation)
			remainOfTransaction = trxOp.NumOfOps
		}
		if remainOfTransaction > 0 {
			transaction = append(transaction, op)
			remainOfTransaction--
			if remainOfTransaction == 0 {
				if err := datatype.ExecuteTransactionRemote(transaction); err != nil {
					return log.OrtooError(err)
				}
				transaction = nil
			}
		} else {
			_, err := op.ExecuteRemote(datatype)
			if err != nil {
				return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
			}
		}
		sseq = opDoc.Sseq
		return nil
	}); err != nil {
		return err
	}

	meta, snap, err := datatype.GetMetaAndSnapshot()
	if err != nil {
		return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	snapStr, err := json.Marshal(snap)
	if err != nil {
		return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	if err := m.mongo.InsertSnapshot(m.ctx, m.collectionDoc.Num, m.datatypeDoc.DUID, sseq, meta, string(snapStr)); err != nil {
		return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}

	data, err := snap.GetAsJSON()
	if err != nil {
		return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	if err := m.mongo.InsertRealSnapshot(m.ctx, m.collectionDoc.Name, m.datatypeDoc.Key, data, sseq); err != nil {
		return model.NewPushPullError(model.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	return nil
}
