package service

import (
	"context"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/knowhunger/ortoo/server/mongodb/schema"
)

func (o *OrtooService) ClientCreate(ctx context.Context, in *model.ClientRequest) (*model.ClientReply, error) {
	log.Logger.Infof("%+v", in)
	doc := schema.ClientModelToBson(in.Client)
	clientDoc, err := o.mongo.GetClient(doc.Cuid)
	if err != nil {
		return nil, log.OrtooError(err, "fail to get client")
	}
	if clientDoc == nil {
		o.mongo.CreateClient(doc)
	} else {
	}
	return nil, nil
}
