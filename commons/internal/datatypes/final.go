package datatypes

import (
	"github.com/gogo/protobuf/proto"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

// FinalDatatype implements the datatype features finally used.
type FinalDatatype struct {
	*TransactionDatatype
	TransactionCtx *TransactionContext
}

// FinalDatatypeInterface defines the interface to obtain FinalDatatype which provide final interfaces.
type FinalDatatypeInterface interface {
	GetFinal() *FinalDatatype
}

// Initialize is a method for initialization
func (c *FinalDatatype) Initialize(
	key string,
	typeOf model.TypeOfDatatype,
	cuid model.CUID,
	w Wire,
	snapshot model.Snapshot,
	finalDatatype model.CommonDatatype) error {

	baseDatatype, err := newBaseDatatype(key, typeOf, cuid)
	if err != nil {
		return log.OrtooErrorf(err, "fail to create baseDatatype")
	}

	wiredDatatype, err := newWiredDatatype(baseDatatype, w)
	if err != nil {
		return log.OrtooErrorf(err, "fail to create wiredDatatype")
	}

	transactionDatatype, err := newTransactionDatatype(wiredDatatype, snapshot)
	if err != nil {
		return log.OrtooErrorf(err, "fail to create transaction manager")
	}
	c.TransactionDatatype = transactionDatatype
	c.TransactionCtx = nil
	c.SetFinalDatatype(finalDatatype)

	return nil
}

// GetMeta returns the binary of metadata of the datatype.
func (c *FinalDatatype) GetMeta() ([]byte, error) {
	meta := model.DatatypeMeta{
		Key:    c.Key,
		DUID:   c.id,
		OpID:   c.opID,
		TypeOf: c.TypeOf,
		State:  c.state,
	}
	metab, err := proto.Marshal(&meta)
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return metab, nil
}

// SetMeta sets the metadata with binary metadata.
func (c *FinalDatatype) SetMeta(meta []byte) error {
	m := model.DatatypeMeta{}
	if err := proto.Unmarshal(meta, &m); err != nil {
		return log.OrtooError(err)
	}
	c.Key = m.Key
	c.id = m.DUID
	c.opID = m.OpID
	c.TypeOf = m.TypeOf
	c.state = m.State
	return nil
}

// SubscribeOrCreate enables a datatype to subscribe and create itself.
func (c *FinalDatatype) SubscribeOrCreate(state model.StateOfDatatype) error {
	if state == model.StateOfDatatype_DUE_TO_SUBSCRIBE {
		c.state = state
		return nil
	}
	subscribeOp, err := model.NewSnapshotOperation(c.TypeOf, state, c.finalDatatype.GetSnapshot())
	if err != nil {
		return log.OrtooErrorf(err, "fail to subscribe")
	}
	_, err = c.ExecuteOperationWithTransaction(c.TransactionCtx, subscribeOp, true)
	if err != nil {
		return log.OrtooErrorf(err, "fail to execute SubscribeOperation")
	}
	return nil
}

// ExecuteTransactionRemote is a method to execute a transaction of remote operations
func (c FinalDatatype) ExecuteTransactionRemote(transaction []model.Operation) error {
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
			if err := c.EndTransaction(transactionCtx, false, false); err != nil {
				_ = log.OrtooError(err)
			}
		}()
		transaction = transaction[1:]
	}
	for _, op := range transaction {
		c.ExecuteOperationWithTransaction(transactionCtx, op, false)
	}
	return nil
}
