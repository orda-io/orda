package operations

import (
	"encoding/json"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
)

// ModelToOperation changes a model.Operation to an operations.Operation
func ModelToOperation(op *model.Operation) iface.Operation {
	switch op.OpType {
	case model.TypeOfOperation_COUNTER_SNAPSHOT,
		model.TypeOfOperation_MAP_SNAPSHOT,
		model.TypeOfOperation_LIST_SNAPSHOT,
		model.TypeOfOperation_DOC_SNAPSHOT:
		return &SnapshotOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, op.Body),
		}
	case model.TypeOfOperation_ERROR:
		return &ErrorOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &errorBody{})),
		}
	case model.TypeOfOperation_TRANSACTION:
		return &TransactionOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &transactionBody{})),
		}
	case model.TypeOfOperation_COUNTER_INCREASE:
		return &IncreaseOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &increaseBody{})),
		}
	case model.TypeOfOperation_MAP_PUT:
		return &PutOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &putBody{})),
		}
	case model.TypeOfOperation_MAP_REMOVE:
		return &RemoveOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &removeBody{})),
		}
	case model.TypeOfOperation_LIST_INSERT:
		return &InsertOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &insertBody{})),
		}
	case model.TypeOfOperation_LIST_DELETE:
		return &DeleteOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &deleteBody{})),
		}
	case model.TypeOfOperation_LIST_UPDATE:
		return &UpdateOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &updateBody{})),
		}
	case model.TypeOfOperation_DOC_OBJ_PUT:
		return &DocPutInObjOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &docPutInObjBody{})),
		}
	case model.TypeOfOperation_DOC_OBJ_RMV:
		return &DocRemoveInObjOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &DocRemoveInObjectBody{})),
		}
	case model.TypeOfOperation_DOC_ARR_INS:
		return &DocInsertToArrayOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &DocInsertToArrayBody{})),
		}
	case model.TypeOfOperation_DOC_ARR_DEL:
		return &DocDeleteInArrayOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &DocDeleteInArrayBody{})),
		}
	case model.TypeOfOperation_DOC_ARR_UPD:
		return &DocUpdateInArrayOperation{
			baseOperation: newBaseOperation(op.OpType, op.ID, unmarshalBody(op.Body, &DocUpdateInArrayBody{})),
		}
	}
	panic("unsupported type of operation")
}

func unmarshalBody(b []byte, c interface{}) interface{} {
	switch c.(type) {
	case string:
		return string(b)
	case []byte:
		return b
	}
	if err := json.Unmarshal(b, c); err != nil {
		log.Logger.Errorf("%v", string(b))
		panic(err) // TODO: this should ne handled
	}
	return c
}

func marshalBody(c interface{}) []byte {
	switch cast := c.(type) {
	case string:
		return []byte(cast)
	case []byte:
		return cast
	}
	j, err := json.Marshal(c)
	if err != nil {
		panic(err) // TODO: this should ne handled
	}
	return j
}
