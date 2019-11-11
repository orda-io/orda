package datatypes

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// CommonDatatype defines common methods
type CommonDatatype struct {
	*TransactionDatatypeImpl
	TransactionCtx *TransactionContext
}

// Initialize is a method for initialization
func (c *CommonDatatype) Initialize(
	key string,
	typeOf model.TypeOfDatatype,
	cuid model.CUID, w Wire,
	snapshot model.Snapshot,
	finalDatatype model.FinalDatatype) error {
	baseDatatype, err := newBaseDatatype(key, typeOf, cuid)
	if err != nil {
		return log.OrtooErrorf(err, "fail to create baseDatatype")
	}

	wiredDatatype, err := newWiredDataType(baseDatatype, w)
	if err != nil {
		return log.OrtooErrorf(err, "fail to create wiredDatatype")
	}

	transactionDatatype, err := newTransactionDatatype(wiredDatatype, snapshot)
	if err != nil {
		return log.OrtooErrorf(err, "fail to create transaction manager")
	}
	c.TransactionDatatypeImpl = transactionDatatype
	c.TransactionCtx = nil
	c.SetFinalDatatype(finalDatatype)

	return nil
}

func (c *CommonDatatype) SubscribeOrCreate(state model.StateOfDatatype) error {
	if state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		c.state = state
		return nil
	}
	subscribeOp, err := model.NewSnapshotOperation(c.TypeOf, state, c.finalDatatype.GetSnapshot())
	if err != nil {
		return log.OrtooErrorf(err, "fail to subscribe")
	}
	_, err = c.ExecuteTransaction(c.TransactionCtx, subscribeOp, true)
	if err != nil {
		return log.OrtooErrorf(err, "fail to execute SubscribeOperation")
	}
	return nil
}

// ExecuteTransactionRemote is a method to execute a transaction of remote operations
func (c *CommonDatatype) ExecuteTransactionRemote(transaction []model.Operation) error {
	var transactionCtx *TransactionContext
	var err error
	if len(transaction) > 1 {
		if err := validateTransaction(transaction); err != nil {
			return log.OrtooErrorf(err, "fail to validate transaction")
		}
		transactionOp := transaction[0].(*model.TransactionOperation)
		transactionCtx, err = c.BeginTransaction(transactionOp.Tag, c.TransactionCtx, false)
		if err != nil {
			return log.OrtooError(err)
		}
		defer func() {
			if err := c.EndTransaction(transactionCtx, false); err != nil {
				_ = log.OrtooError(err)
			}
		}()
		transaction = transaction[1:]
	}
	for _, op := range transaction {
		c.ExecuteTransaction(transactionCtx, op, false)
	}
	return nil
}
