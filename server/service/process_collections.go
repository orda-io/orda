package service

import (
	goctx "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/mongodb"
)

// CreateCollection creates a collection
func (its *OrtooService) CreateCollection(goCtx goctx.Context, in *model.CollectionMessage) (*model.CollectionMessage, error) {
	ctx := context.New(goCtx)
	num, err := mongodb.MakeCollection(ctx, its.mongo, in.Collection)
	var msg string
	if err != nil {
		msg = fmt.Sprintf("Fail to create collection '%s'", in.Collection)
		return nil, errors.NewRPCError(err)
	} else {
		msg = fmt.Sprintf("Created collection '%s(%d)'", in.Collection, num)
	}
	ctx.L().Infof("%s", msg)
	return in, nil
}

// ResetCollection resets a collection
func (its *OrtooService) ResetCollection(goCtx goctx.Context, in *model.CollectionMessage) (*model.CollectionMessage, error) {
	ctx := context.New(goCtx)
	if err := its.mongo.PurgeCollection(ctx, in.Collection); err != nil {
		return nil, errors.NewRPCError(err)
	}
	return its.CreateCollection(goCtx, in)
}
