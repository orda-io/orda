package service

import (
	gocontext "context"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
	"github.com/orda-io/orda/server/admin"
	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/schema"
	"github.com/orda-io/orda/server/snapshot"
	"github.com/orda-io/orda/server/utils"
)

// PatchDocument patches document datatype
func (its *OrdaService) PatchDocument(goCtx gocontext.Context, req *model.PatchMessage) (*model.PatchMessage, error) {
	ctx := context.NewOrdaContext(goCtx, constants.TagPatch).
		UpdateCollectionTags(req.Collection, 0)
	collectionDoc, rpcErr := its.getCollectionDocWithRPCError(ctx, req.Collection)
	if rpcErr != nil {
		return nil, rpcErr
	}
	clientDoc := admin.NewPatchClient(collectionDoc)
	ctx.UpdateCollectionTags(collectionDoc.Name, collectionDoc.Num).
		UpdateClientTags(clientDoc.Alias, clientDoc.CUID).
		UpdateDatatypeTags(req.Key, "")

	ctx.L().Infof("BEGIN PatchDocument: '%v' %#v", req.Key, req.GetJson())
	defer ctx.L().Infof("END PatchDocument: '%v'", req.Key)

	lock := its.managers.GetLock(ctx, utils.GetLockName("PD", collectionDoc.Num, req.Key))
	if !lock.TryLock() {
		return nil, errors.NewRPCError(errors.ServerInit.New(ctx.L(), "fail to lock"))
	}
	defer lock.Unlock()

	datatypeDoc, rpcErr := its.managers.Mongo.GetDatatypeByKey(ctx, collectionDoc.Num, req.Key)
	if rpcErr != nil {
		return nil, rpcErr
	}

	if datatypeDoc == nil {
		datatypeDoc = schema.NewDatatypeDoc("", req.Key, collectionDoc.Num, model.TypeOfDatatype_DOCUMENT.String())
	}
	if datatypeDoc.Type != model.TypeOfDatatype_DOCUMENT.String() {
		return nil, errors.NewRPCError(errors.ServerBadRequest.New(ctx.L(), "not document type: "+datatypeDoc.Type))
	}

	snapshotManager := snapshot.NewManager(ctx, its.managers, datatypeDoc, collectionDoc)
	data, lastSseq, err := snapshotManager.GetLatestDatatype()
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	ctx.UpdateDatatypeTags(data.GetKey(), data.GetDUID())

	if lastSseq > 0 {
		data.SetState(model.StateOfDatatype_SUBSCRIBED)
		data.SetCheckPoint(lastSseq, 0)
	}
	doc := data.(orda.Document)
	patches, err := doc.(orda.Document).PatchByJSON(req.Json)
	if err != nil {
		return nil, errors.NewRPCError(errors.ServerBadRequest.New(ctx.L(), err.Error()))
	}

	if len(patches) > 0 {
		ppp := doc.(iface.Datatype).CreatePushPullPack()
		ctx.L().Infof("%v", ppp.ToString(true))

		pushPullHandler := newPushPullHandler(ctx, ppp, clientDoc, collectionDoc, its.managers)
		pppCh := pushPullHandler.Start()
		_ = <-pppCh
	}

	return &model.PatchMessage{
		Key:        req.Key,
		Collection: req.Collection,
		Json:       string(doc.ToJSONBytes()),
	}, nil

}
