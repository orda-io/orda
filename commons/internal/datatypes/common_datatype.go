package datatypes

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

//CommonDatatype defines common methods
type CommonDatatype struct {
	*TransactionDatatypeImpl
	TransactionCtx *TransactionContext
}

//Initialize is a method for initialization
func (c *CommonDatatype) Initialize(key string, typeOf model.TypeOfDatatype, cuid model.Cuid, w Wire, snapshot model.Snapshot, finalDatatype model.FinalDatatype) error {
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

func (c *CommonDatatype) Subscribe() error {
	subscOp, err := model.NewSubscribeOperation(c.TypeOf, c.finalDatatype.GetSnapshot())
	if err != nil {
		return log.OrtooErrorf(err, "fail to subscribe")
	}
	_, err = c.ExecuteTransaction(c.TransactionCtx, subscOp, true)
	if err != nil {
		return log.OrtooErrorf(err, "fail to execute SubscribeOperation")
	}
	return nil
}
