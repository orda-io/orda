package operations

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/iface"
	"github.com/knowhunger/ortoo/ortoo/model"
)

// ModelToOperation changes a model.Operation to an operations.Operation
func ModelToOperation(op *model.Operation) iface.Operation {
	switch op.OpType {
	case model.TypeOfOperation_SNAPSHOT:
		var c snapshotContent
		unmarshalContent(op.Json, &c)
		return &SnapshotOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DELETE:
	case model.TypeOfOperation_ERROR:
		var c errorContent
		unmarshalContent(op.Json, &c)
		return &ErrorOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_TRANSACTION:
		var c transactionContent
		unmarshalContent(op.Json, &c)
		return &TransactionOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_INT_COUNTER_INCREASE:
		var c increaseContent
		unmarshalContent(op.Json, &c)
		return &IncreaseOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_HASH_MAP_PUT:
		var c putContent
		unmarshalContent(op.Json, &c)
		return &PutOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_HASH_MAP_REMOVE:
		var c removeContent
		unmarshalContent(op.Json, &c)
		return &RemoveOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_LIST_INSERT:
		var c insertContent
		unmarshalContent(op.Json, &c)
		return &InsertOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_LIST_DELETE:
		var c deleteContent
		unmarshalContent(op.Json, &c)
		return &DeleteOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_LIST_UPDATE:
		var c updateContent
		unmarshalContent(op.Json, &c)
		return &UpdateOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOCUMENT_PUT_OBJ:
		var c docPutInObjectContent
		unmarshalContent(op.Json, &c)
		return &DocPutInObjectOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOCUMENT_INS_ARR:
		var c docInsertToArrayContent
		unmarshalContent(op.Json, &c)
		return &DocInsertToArrayOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOCUMENT_DEL_OBJ:
		var c docDeleteInObjectContent
		unmarshalContent(op.Json, &c)
		return &DocDeleteInObjectOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DOCUMENT_DEL_ARR:
		var c docDeleteInArrayContent
		unmarshalContent(op.Json, &c)
		return &DocDeleteInArrayOperation{
			baseOperation: &baseOperation{ID: op.ID},
			C:             c,
		}
	}
	panic("unsupported type of operation")
}

func unmarshalContent(b []byte, c interface{}) {
	if err := json.Unmarshal(b, c); err != nil {
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
