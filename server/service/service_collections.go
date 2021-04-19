package service

import (
	goctx "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/constants"
	"github.com/knowhunger/ortoo/server/mongodb"
	"github.com/knowhunger/ortoo/server/svrcontext"
)

// CreateCollection creates a collection
func (its *OrtooService) CreateCollection(goCtx goctx.Context, in *model.CollectionMessage) (*model.CollectionMessage, error) {
	ctx := svrcontext.NewServerContext(goCtx, constants.TagCreate).UpdateCollection(in.Collection)
	num, err := mongodb.MakeCollection(ctx, its.mongo, in.Collection)
	var msg string
	if err != nil {
		return nil, errors.NewRPCError(err)
	} else {
		msg = fmt.Sprintf("create collection '%s(%d)'", in.Collection, num)
	}
	ctx.L().Infof("%s", msg)
	return in, nil
}

// ResetCollection resets a collection
func (its *OrtooService) ResetCollection(goCtx goctx.Context, in *model.CollectionMessage) (*model.CollectionMessage, error) {
	ctx := svrcontext.NewServerContext(goCtx, constants.TagReset).UpdateCollection(in.Collection)
	if err := its.mongo.PurgeCollection(ctx, in.Collection); err != nil {
		return nil, errors.NewRPCError(err)
	}
	ctx.L().Infof("reset %s collection", in.Collection)
	return its.CreateCollection(goCtx, in)
}
