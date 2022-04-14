package service

import (
	gocontext "context"
	"github.com/orda-io/orda/pkg/errors"
	"github.com/orda-io/orda/pkg/iface"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/orda"
	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/schema"
	"github.com/orda-io/orda/server/snapshot"
	"github.com/orda-io/orda/server/svrcontext"
)

func (its *OrdaService) PatchDocument(goCtx gocontext.Context, req *model.PatchMessage) (*model.PatchMessage, error) {
	ctx := svrcontext.NewServerContext(goCtx, constants.TagPatch).UpdateCollection(req.Collection)
	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, req.Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}

	clientDoc := schema.NewPatchClient(collectionDoc)
	ctx.UpdateCollection(collectionDoc.GetSummary()).UpdateClient(clientDoc.ToString())
	ctx.L().Infof("BEGIN PatchDocument: %#v", req.GetJson())
	defer ctx.L().Infof("END PatchDocument")

	datatypeDoc, rpcErr := its.mongo.GetDatatypeByKey(ctx, collectionDoc.Num, req.Key)
	if rpcErr != nil {
		return nil, rpcErr
	}

	if datatypeDoc == nil {
		datatypeDoc = &schema.DatatypeDoc{
			Key:           req.Key,
			CollectionNum: collectionDoc.Num,
			Type:          model.TypeOfDatatype_DOCUMENT.String(),
		}
	}
	if datatypeDoc.Type != model.TypeOfDatatype_DOCUMENT.String() {
		return nil, errors.NewRPCError(errors.ServerBadRequest.New(ctx.L(), "not document type: "+datatypeDoc.Type))
	}
	snapshotManager := snapshot.NewManager(ctx, its.mongo, datatypeDoc, collectionDoc)
	data, lastSseq, err := snapshotManager.GetLatestDatatype()
	ctx.UpdateDatatype(data.GetSummary())
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	if lastSseq > 0 {
		data.SetState(model.StateOfDatatype_SUBSCRIBED)
		data.(iface.Datatype).SetCheckPoint(lastSseq, 0)
	}
	doc := data.(orda.Document)
	n, err := doc.(orda.Document).PatchByJSON(req.Json)
	if err != nil {
		return nil, errors.NewRPCError(errors.ServerBadRequest.New(ctx.L(), err.Error()))
	}

	if n > 0 {
		ppp := doc.(iface.Datatype).CreatePushPullPack()
		ctx.L().Infof("%v", ppp.ToString())

		pushPullHandler := newPushPullHandler(ctx, ppp, clientDoc, collectionDoc, its)
		pppCh := pushPullHandler.Start()
		_ = <-pppCh
	}

	return &model.PatchMessage{
		Key:        req.Key,
		Collection: req.Collection,
		Json:       string(doc.ToJSONBytes()),
	}, nil

}
