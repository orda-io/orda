package service

import (
	gocontext "context"
	iface2 "github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	model2 "github.com/orda-io/orda/client/pkg/model"
	operations2 "github.com/orda-io/orda/client/pkg/operations"
	orda2 "github.com/orda-io/orda/client/pkg/orda"
)

func (its *OrdaService) decodeModelOp(in *model2.Operation) iface2.Operation {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	op := operations2.ModelToOperation(in)
	return op
}

func (its *OrdaService) TestEncodingOperation(
	goCtx gocontext.Context,
	in *model2.EncodingMessage,
) (ret *model2.EncodingMessage, er error) {
	log.Logger.Infof("Receive %v", in)
	defer func() {
		log.Logger.Infof("Returns %v, %v", ret, er)
	}()
	decodedOp := its.decodeModelOp(in.Op)
	switch cast := decodedOp.(type) {
	case *operations2.SnapshotOperation:
		return its.testEncodingSnapshotOperation(goCtx, in.Type, cast)
	case *operations2.ErrorOperation:
		{
			op := operations2.NewErrorOperationWithCodeAndMsg(cast.GetCode(), cast.GetMessage())
			in.Op = op.ToModelOperation()
		}
	case *operations2.TransactionOperation:
		{
			op := operations2.NewTransactionOperation(cast.GetBody().Tag)
			op.SetNumOfOps(int(cast.GetBody().NumOfOps))
			in.Op = op.ToModelOperation()
		}
	case *operations2.IncreaseOperation:
		{
			op := operations2.NewIncreaseOperation(cast.GetBody().Delta)
			in.Op = op.ToModelOperation()
		}
	case *operations2.PutOperation:
		{
			op := operations2.NewPutOperation(cast.GetBody().Key, cast.GetBody().Value)
			in.Op = op.ToModelOperation()
		}
	case *operations2.RemoveOperation:
		{
			op := operations2.NewRemoveOperation(cast.GetBody().Key)
			in.Op = op.ToModelOperation()
		}
	case *operations2.InsertOperation:
		{
			op := operations2.NewInsertOperation(cast.Pos, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations2.DeleteOperation:
		{
			op := operations2.NewDeleteOperation(cast.Pos, cast.NumOfNodes)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations2.UpdateOperation:
		{
			op := operations2.NewUpdateOperation(cast.Pos, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations2.DocPutInObjOperation:
		{
			op := operations2.NewDocPutInObjOperation(cast.GetBody().P, cast.GetBody().K, cast.GetBody().V)
			in.Op = op.ToModelOperation()
		}
	case *operations2.DocRemoveInObjOperation:
		{
			op := operations2.NewDocRemoveInObjOperation(cast.GetBody().P, cast.GetBody().K)
			in.Op = op.ToModelOperation()
		}
	case *operations2.DocInsertToArrayOperation:
		{
			op := operations2.NewDocInsertToArrayOperation(cast.GetBody().P, 0, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations2.DocDeleteInArrayOperation:
		{
			op := operations2.NewDocDeleteInArrayOperation(cast.GetBody().P, 0, 0)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	case *operations2.DocUpdateInArrayOperation:
		{
			op := operations2.NewDocUpdateInArrayOperation(cast.GetBody().P, 0, cast.GetBody().V)
			op.GetBody().T = cast.GetBody().T
			in.Op = op.ToModelOperation()
		}
	default:
		in.Op = operations2.NewDeleteOperation(1, 10).ToModelOperation()
	}
	in.Op.ID = decodedOp.GetID()
	return in, nil
}

func (its *OrdaService) testEncodingSnapshotOperation(
	goCtx gocontext.Context,
	typeOf model2.TypeOfDatatype,
	sOp *operations2.SnapshotOperation,
) (
	*model2.EncodingMessage,
	error,
) {
	client := orda2.NewClient(orda2.NewLocalClientConfig("ENCODING"), "orda-encoding-tester")
	datatype := client.CreateDatatype("Testing", typeOf, nil).(iface2.Datatype)

	if _, err := datatype.ExecuteRemote(sOp); err != nil {
		return nil, err
	}

	regenOp, err := datatype.CreateSnapshotOperation()
	if err != nil {
		return nil, err
	}
	regenOp.SetID(sOp.ID)
	return &model2.EncodingMessage{Type: typeOf, Op: regenOp.ToModelOperation()}, nil
}
