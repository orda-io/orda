package datatypes

import (
	"github.com/knowhunger/ortoo/commons/model"
)

type Wire interface {
	DeliverOperation(wired WiredDatatype, op model.Operation)
	DeliverTransaction(wired WiredDatatype, transaction []model.Operation)
}
