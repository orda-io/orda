package operations

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/pkg/iface"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/knowhunger/ortoo/pkg/model"
)

// ModelToOperation changes a model.Operation to an operations.Operation
func ModelToOperation(op *model.Operation) iface.Operation {
	switch op.OpType {
	case model.TypeOfOperation_SNAPSHOT:
		var c snapshotContent
		unmarshalContent(op.Body, &c)
		return &SnapshotOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             &c,
		}
	case model.TypeOfOperation_DELETE:
	case model.TypeOfOperation_ERROR:
		var c errorContent
		unmarshalContent(op.Body, &c)
		return &ErrorOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_TRANSACTION:
		var c transactionContent
		unmarshalContent(op.Body, &c)
		return &TransactionOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_COUNTER_INCREASE:
		var c increaseContent
		unmarshalContent(op.Body, &c)
		return &IncreaseOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_MAP_PUT:
		var c putContent
		unmarshalContent(op.Body, &c)
		return &PutOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_MAP_REMOVE:
		var c removeContent
		unmarshalContent(op.Body, &c)
		return &RemoveOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_LIST_INSERT:
		var c insertContent
		unmarshalContent(op.Body, &c)
		return &InsertOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_LIST_DELETE:
		var c deleteContent
		unmarshalContent(op.Body, &c)
		return &DeleteOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_LIST_UPDATE:
		var c updateContent
		unmarshalContent(op.Body, &c)
		return &UpdateOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOC_PUT_OBJ:
		var c docPutInObjectContent
		unmarshalContent(op.Body, &c)
		return &DocPutInObjectOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOC_INS_ARR:
		var c docInsertToArrayContent
		unmarshalContent(op.Body, &c)
		return &DocInsertToArrayOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOC_DEL_OBJ:
		var c docDeleteInObjectContent
		unmarshalContent(op.Body, &c)
		return &DocDeleteInObjectOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOC_DEL_ARR:
		var c docDeleteInArrayContent
		unmarshalContent(op.Body, &c)
		return &DocDeleteInArrayOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	}
	panic("unsupported type of operation")
}

func unmarshalContent(b []byte, c interface{}) {
	if err := json.Unmarshal(b, c); err != nil {
		log.Logger.Errorf("%v", string(b))
		panic(err) // TODO: this should ne handled
	}
}

func marshalContent(c interface{}) []byte {
	j, err := json.Marshal(c)
	if err != nil {
		panic(err) // TODO: this should ne handled
	}
	return j
}
