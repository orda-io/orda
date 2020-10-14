package snapshot

import (
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/pkg/operations"
	"github.com/knowhunger/ortoo/pkg/ortoo"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
)

// Manager is a struct that updates snapshot of a datatype in Ortoo server
type Manager struct {
	ctx           context.OrtooContext
	mongo         *mongodb.RepositoryMongo
	datatypeDoc   *schema.DatatypeDoc
	collectionDoc *schema.CollectionDoc
}

// NewManager returns an instance of Snapshot Manager
func NewManager(
	ctx context.OrtooContext,
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

// UpdateSnapshot updates snapshot for specified datatype
func (its *Manager) UpdateSnapshot() errors.OrtooError {
	var sseq uint64 = 0
	client := ortoo.NewClient(ortoo.NewLocalClientConfig(its.collectionDoc.Name), "server")
	datatype := client.CreateDatatype(its.datatypeDoc.Key, its.datatypeDoc.GetType(), nil).(iface.Datatype)
	datatype.SetLogger(its.ctx.L())
	snapshotDoc, err := its.mongo.GetLatestSnapshot(its.ctx, its.collectionDoc.Num, its.datatypeDoc.DUID)
	if err != nil {
		return err
	}
	if snapshotDoc != nil {
		sseq = snapshotDoc.Sseq
		if err := datatype.SetMetaAndSnapshot(snapshotDoc.Meta, []byte(snapshotDoc.Snapshot)); err != nil {
			return err
		}
	}
	var transaction []*model.Operation
	var remainOfTransaction int32 = 0
	err = its.mongo.GetOperations(its.ctx, its.datatypeDoc.DUID, sseq+1, constants.InfinitySseq,
		func(opDoc *schema.OperationDoc) errors.OrtooError {
			var modelOp = opDoc.GetOperation()
			if modelOp.OpType == model.TypeOfOperation_TRANSACTION {
				trxOp := operations.ModelToOperation(modelOp).(*operations.TransactionOperation)
				remainOfTransaction = trxOp.GetNumOfOps()
			}
			if remainOfTransaction > 0 {
				transaction = append(transaction, modelOp)
				remainOfTransaction--
				if remainOfTransaction == 0 {
					_, err := datatype.ExecuteRemoteTransaction(transaction, false)
					if err != nil {
						return err
					}
					transaction = nil
				}
			} else {
				op := operations.ModelToOperation(modelOp)
				its.ctx.L().Infof("%v", op.String())
				if _, err := op.ExecuteRemote(datatype); err != nil {
					return err
				}
			}
			sseq = opDoc.Sseq
			return nil
		})
	if err != nil {
		return err
	}

	meta, snap, err := datatype.GetMetaAndSnapshot()
	if err != nil {
		return err
	}
	its.ctx.L().Infof("snapB: %v", string(snap))
	if err := its.mongo.InsertSnapshot(its.ctx, its.collectionDoc.Num, its.datatypeDoc.DUID, sseq, meta, string(snap)); err != nil {
		return err
	}

	data := datatype.GetSnapshot().GetAsJSONCompatible()
	if err := its.mongo.InsertRealSnapshot(its.ctx, its.collectionDoc.Name, its.datatypeDoc.Key, data, sseq); err != nil {
		return err
	}
	return nil
}
