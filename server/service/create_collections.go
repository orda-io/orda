package service

import (
	goctx "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/model"
	"github.com/knowhunger/ortoo/server/mongodb"
)

func (its *OrtooService) CreateCollections(goCtx goctx.Context, in *model.CollectionMessage) (*model.CollectionMessage, error) {
	ctx := context.New(goCtx)
	// collectionName := strings.TrimPrefix(req.URL.Path, apiCollections)
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
