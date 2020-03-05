package snapshot

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/commons/serverside"
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

// UpdateSnapshot updates snapshot for specified datatype
func (m *Manager) UpdateSnapshot() error {
	var sseq uint64 = 0
	datatype, err := serverside.NewDatatype(m.datatypeDoc.Key, model.TypeOfDatatype_INT_COUNTER)
	if err != nil {
		return log.OrtooError(err)
	}
	snapshotDoc, err := m.mongo.GetLatestSnapshot(m.ctx, m.collectionDoc.Num, m.datatypeDoc.DUID)
	if err != nil {
		return log.OrtooError(err)
	}
	if snapshotDoc != nil {
		sseq = snapshotDoc.Sseq
		serverside.SetSnapshot(datatype, snapshotDoc.Meta, snapshotDoc.Snapshot)
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
				return log.OrtooError(err)
			}
		}
		sseq = opDoc.Sseq
		return nil
	}); err != nil {
		return log.OrtooError(err)
	}

	meta, data, err := datatype.GetMetaAndSnapshot()
	if err != nil {
		return log.OrtooError(err)
	}
	if err := m.mongo.InsertSnapshot(m.ctx, m.collectionDoc.Num, m.datatypeDoc.DUID, sseq, meta, data); err != nil {
		return log.OrtooError(err)
	}

	if err := m.mongo.InsertRealSnapshot(m.ctx, m.collectionDoc.Name, m.datatypeDoc.Key, data, sseq); err != nil {
		return log.OrtooError(err)
	}
	return nil
}
