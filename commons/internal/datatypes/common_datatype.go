package datatypes

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
)

//CommonDatatype defines common methods
type CommonDatatype struct {
	*TransactionDatatypeImpl
}

//Initialize is a method for initialization
func (c *CommonDatatype) Initialize(key string, typeOf model.TypeOfDatatype, cuid model.Cuid, w Wire, snapshot Snapshot, finalDatatype model.FinalDatatype) error {
	baseDatatype, err := newBaseDatatype(key, typeOf, cuid)
	if err != nil {
		return log.OrtooError(err, "fail to create baseDatatype")
	}

	wiredDatatype, err := newWiredDataType(baseDatatype, w)
	if err != nil {
		return log.OrtooError(err, "fail to create wiredDatatype")
	}

	transactionDatatype, err := newTransactionDatatype(wiredDatatype, snapshot)
	if err != nil {
		return log.OrtooError(err, "fail to create transaction manager")
	}
	c.TransactionDatatypeImpl = transactionDatatype
	c.SetFinalDatatype(finalDatatype)
	return nil
}
