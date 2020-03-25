package operations

import (
	"encoding/json"
	"github.com/knowhunger/ortoo/ortoo/model"
)

func ModelToOperation(op *model.Operation) Operation {
	switch op.OpType {
	case model.TypeOfOperation_SNAPSHOT:
		var c SnapshotContent
		unmarshalContent(op.Json, &c)
		return &SnapshotOperation{
			BaseOperation: &BaseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_DELETE:
	case model.TypeOfOperation_ERROR:
		var c ErrorContent
		unmarshalContent(op.Json, &c)
		return &ErrorOperation{
			BaseOperation: &BaseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_TRANSACTION:
		var c TransactionContent
		unmarshalContent(op.Json, &c)
		return &TransactionOperation{
			BaseOperation: &BaseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_INT_COUNTER_INCREASE:
		var c IncreaseContent
		unmarshalContent(op.Json, &c)
		return &IncreaseOperation{
			BaseOperation: &BaseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_HASH_MAP_PUT:
		var c PutContent
		unmarshalContent(op.Json, &c)
		return &PutOperation{
			BaseOperation: &BaseOperation{ID: op.ID},
			C:             c,
		}
	case model.TypeOfOperation_HASH_MAP_REMOVE:
		var c RemoveContent
		unmarshalContent(op.Json, &c)
		return &RemoveOperation{
			BaseOperation: &BaseOperation{ID: op.ID},
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
