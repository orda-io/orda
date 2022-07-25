package service

import (
	gocontext "context"
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	model2 "github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"

	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/schema"
	"github.com/orda-io/orda/server/snapshot"
	"github.com/orda-io/orda/server/svrcontext"
	"github.com/orda-io/orda/server/utils"
)

func (its *OrdaService) PatchDocument(goCtx gocontext.Context, req *model2.PatchMessage) (*model2.PatchMessage, error) {
	ctx := svrcontext.NewServerContext(goCtx, constants.TagPatch).UpdateCollection(req.Collection)
	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, req.Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}
	ctx.L().Infof("BEGIN PatchDocument: '%v' %#v", req.Key, req.GetJson())
	defer ctx.L().Infof("END PatchDocument: '%v'", req.Key)

	clientDoc := schema.NewPatchClient(collectionDoc)
	ctx.UpdateCollection(collectionDoc.GetSummary()).UpdateClient(clientDoc.ToString())

	lock := its.managers.GetLock(ctx, utils.GetLockName("PD", collectionDoc.Num, req.Key))
	if !lock.TryLock() {
		return nil, errors2.NewRPCError(errors2.ServerInit.New(ctx.L(), "fail to lock"))
	}
	defer lock.Unlock()

	datatypeDoc, rpcErr := its.managers.Mongo.GetDatatypeByKey(ctx, collectionDoc.Num, req.Key)
	if rpcErr != nil {
		return nil, rpcErr
	}

	if datatypeDoc == nil {
		datatypeDoc = &schema.DatatypeDoc{
			Key:           req.Key,
			CollectionNum: collectionDoc.Num,
			Type:          model2.TypeOfDatatype_DOCUMENT.String(),
		}
	}
	if datatypeDoc.Type != model2.TypeOfDatatype_DOCUMENT.String() {
		return nil, errors2.NewRPCError(errors2.ServerBadRequest.New(ctx.L(), "not document type: "+datatypeDoc.Type))
	}

	snapshotManager := snapshot.NewManager(ctx, its.managers, datatypeDoc, collectionDoc)
	data, lastSseq, err := snapshotManager.GetLatestDatatype()
	if err != nil {
		return nil, errors2.NewRPCError(err)
	}
	ctx.UpdateDatatype(data.GetSummary())

	if lastSseq > 0 {
		data.SetState(model2.StateOfDatatype_SUBSCRIBED)
		data.SetCheckPoint(lastSseq, 0)
	}
	doc := data.(orda.Document)
	patches, err := doc.(orda.Document).PatchByJSON(req.Json)
	if err != nil {
		return nil, errors2.NewRPCError(errors2.ServerBadRequest.New(ctx.L(), err.Error()))
	}

	if len(patches) > 0 {
		ppp := doc.(iface.Datatype).CreatePushPullPack()
		ctx.L().Infof("%v", ppp.ToString())

		pushPullHandler := newPushPullHandler(ctx, ppp, clientDoc, collectionDoc, its.managers)
		pppCh := pushPullHandler.Start()
		_ = <-pppCh
	}

	return &model2.PatchMessage{
		Key:        req.Key,
		Collection: req.Collection,
		Json:       string(doc.ToJSONBytes()),
	}, nil

}
