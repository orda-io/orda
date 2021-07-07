package service

import (
	gocontext "context"
	"github.com/orda-io/orda/pkg/iface"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/operations"
	"github.com/orda-io/orda/pkg/orda"
)

func (its *OrdaService) decodeModelOp(in *model.Operation) iface.Operation {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	op := operations.ModelToOperation(in)
	return op
}

func (its *OrdaService) TestEncodingOperation(
	goCtx gocontext.Context,
	in *model.EncodingMessage,
) (ret *model.EncodingMessage, er error) {
	log.Logger.Infof("Receive %v", in)
	defer func() {
		log.Logger.Infof("Returns %v, %v", ret, er)
	}()
	decodedOp := its.decodeModelOp(in.Op)
	switch cast := decodedOp.(type) {
	case *operations.SnapshotOperation:
		return its.testEncodingSnapshotOperation(goCtx, in.Type, cast)
	case *operations.ErrorOperation:
		{
			op := operations.NewErrorOperationWithCodeAndMsg(cast.GetCode(), cast.GetMessage())
			in.Op = op.ToModelOperation()
		}
	case *operations.TransactionOperation:
		{
			op := operations.NewTransactionOperation(cast.GetBody().Tag)
			op.SetNumOfOps(int(cast.GetBody().NumOfOps))
			in.Op = op.ToModelOperation()
		}
	case *operations.IncreaseOperation:
		{
			op := operations.NewIncreaseOperation(cast.GetBody().Delta)
			in.Op = op.ToModelOperation()
		}
	case *operations.PutOperation:
		{
			op := operations.NewPutOperation(cast.GetBody().Key, cast.GetBody().Value)
			in.Op = op.ToModelOperation()
		}
	case *operations.RemoveOperation:
		{
			op := operations.NewRemoveOperation(cast.GetBody().Key)
			in.Op = op.ToModelOperation()
		}
	case *operations.InsertOperation:
		{
			op := operations.NewInsertOperation(cast.Pos, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations.DeleteOperation:
		{
			op := operations.NewDeleteOperation(cast.Pos, cast.NumOfNodes)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations.UpdateOperation:
		{
			op := operations.NewUpdateOperation(cast.Pos, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations.DocPutInObjOperation:
		{
			op := operations.NewDocPutInObjOperation(cast.GetBody().P, cast.GetBody().K, cast.GetBody().V)
			in.Op = op.ToModelOperation()
		}
	case *operations.DocRemoveInObjOperation:
		{
			op := operations.NewDocRemoveInObjOperation(cast.GetBody().P, cast.GetBody().K)
			in.Op = op.ToModelOperation()
		}
	case *operations.DocInsertToArrayOperation:
		{
			op := operations.NewDocInsertToArrayOperation(cast.GetBody().P, 0, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations.DocDeleteInArrayOperation:
		{
			op := operations.NewDocDeleteInArrayOperation(cast.GetBody().P, 0, 0)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations.DocUpdateInArrayOperation:
		{
			op := operations.NewDocUpdateInArrayOperation(cast.GetBody().P, 0, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	default:
		in.Op = operations.NewDeleteOperation(1, 10).ToModelOperation()
	}
	in.Op.ID = decodedOp.GetID()
	return in, nil
}

func (its *OrdaService) testEncodingSnapshotOperation(
	goCtx gocontext.Context,
	typeOf model.TypeOfDatatype,
	sOp *operations.SnapshotOperation,
) (
	*model.EncodingMessage,
	error,
) {
	client := orda.NewClient(orda.NewLocalClientConfig("ENCODING"), "orda-encoding-tester")
	datatype := client.CreateDatatype("Testing", typeOf, nil).(iface.Datatype)

	if _, err := datatype.ExecuteRemote(sOp); err != nil {
		return nil, err
	}

	regenOp, err := operations.NewSnapshotOperationFromDatatype(datatype)
	if err != nil {
		return nil, err
	}
	regenOp.SetID(sOp.ID)
	return &model.EncodingMessage{Type: typeOf, Op: regenOp.ToModelOperation()}, nil
}
