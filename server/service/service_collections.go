package service

import (
	goctx "context"
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/model"

	"github.com/orda-io/orda/server/constants"
	"github.com/orda-io/orda/server/mongodb"
)

// CreateCollection creates a collection
func (its *OrdaService) CreateCollection(goCtx goctx.Context, in *model.CollectionMessage) (*model.CollectionMessage, error) {
	ctx := context.NewOrdaContext(goCtx, constants.TagCreate).
		UpdateCollectionTags(in.Collection, 0)
	num, err := mongodb.MakeCollection(ctx, its.managers.Mongo, in.Collection)
	var msg string
	if err != nil {
		return nil, errors.NewRPCError(err)
	}

	msg = fmt.Sprintf("create collection '%s(%d)'", in.Collection, num)

	ctx.L().Infof("%s", msg)
	return in, nil
}

// ResetCollection resets a collection
func (its *OrdaService) ResetCollection(goCtx goctx.Context, in *model.CollectionMessage) (*model.CollectionMessage, error) {
	ctx := context.NewOrdaContext(goCtx, constants.TagReset).
		UpdateCollectionTags(in.Collection, 0)
	if err := its.managers.Mongo.PurgeCollection(ctx, in.Collection); err != nil {
		return nil, errors.NewRPCError(err)
	}
	ctx.L().Infof("reset %s collection", in.Collection)
	return its.CreateCollection(goCtx, in)
}
