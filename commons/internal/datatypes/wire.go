package datatypes

import (
	"github.com/knowhunger/ortoo/commons/model"
)

//Wire defines the interfaces related to delivering operations.
type Wire interface {
	DeliverOperation(wired WiredDatatype, op model.Operation)
	DeliverTransaction(wired WiredDatatype, transaction []model.Operation)
}
