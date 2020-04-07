package snapshot

import (
	"context"
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo"
	"github.com/knowhunger/ortoo/ortoo/errors"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/operations"
	"github.com/knowhunger/ortoo/ortoo/types"
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

func (m *Manager) getPushPullTag() errors.PushPullTag {
	return errors.PushPullTag{
		CollectionName: m.collectionDoc.Name,
		Key:            m.datatypeDoc.Key,
		DUID:           m.datatypeDoc.DUID,
	}
}

// UpdateSnapshot updates snapshot for specified datatype
func (m *Manager) UpdateSnapshot() error {
	var sseq uint64 = 0
	client := ortoo.NewClient(ortoo.NewLocalClientConfig(m.collectionDoc.Name), "server")
	datatype := client.CreateDatatype(m.datatypeDoc.Key, m.datatypeDoc.GetType(), nil).(types.Datatype)
	snapshotDoc, err := m.mongo.GetLatestSnapshot(m.ctx, m.collectionDoc.Num, m.datatypeDoc.DUID)
	if err != nil {
		return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	if snapshotDoc != nil {
		sseq = snapshotDoc.Sseq
		if err := datatype.SetMetaAndSnapshot(snapshotDoc.Meta, snapshotDoc.Snapshot); err != nil {
			return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
		}
	}
	var transaction []*model.Operation
	var remainOfTransaction int32 = 0
	if err := m.mongo.GetOperations(m.ctx, m.datatypeDoc.DUID, sseq+1, constants.InfinitySseq,
		func(opDoc *schema.OperationDoc) error {
			var modelOp = opDoc.GetOperation()
			if modelOp.OpType == model.TypeOfOperation_TRANSACTION {
				trxOp := operations.ModelToOperation(modelOp).(*operations.TransactionOperation)
				remainOfTransaction = trxOp.GetNumOfOps()
			}
			if remainOfTransaction > 0 {
				transaction = append(transaction, modelOp)
				remainOfTransaction--
				if remainOfTransaction == 0 {
					_, err := datatype.ExecuteTransactionRemote(transaction, false)
					if err != nil {
						return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
					}
					transaction = nil
				}
			} else {
				op := operations.ModelToOperation(modelOp)
				_, err := op.ExecuteRemote(datatype)
				if err != nil {
					return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
				}
			}
			sseq = opDoc.Sseq
			return nil
		}); err != nil {
		return err
	}

	meta, snap, err := datatype.GetMetaAndSnapshot()
	if err != nil {
		return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	snapb, err := json.Marshal(snap)
	if err != nil {
		return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	if err := m.mongo.InsertSnapshot(m.ctx, m.collectionDoc.Num, m.datatypeDoc.DUID, sseq, meta, string(snapb)); err != nil {
		return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}

	data := snap.GetAsJSON()
	if err := m.mongo.InsertRealSnapshot(m.ctx, m.collectionDoc.Name, m.datatypeDoc.Key, data, sseq); err != nil {
		return errors.NewPushPullError(errors.PushPullErrUpdateSnapshot, m.getPushPullTag(), err.Error())
	}
	return nil
}
