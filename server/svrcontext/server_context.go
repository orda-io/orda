package svrcontext

import (
	gocontext "context"
	"fmt"
	"github.com/knowhunger/ortoo/pkg/context"
)

type ServerContext struct {
	context.OrtooContext
	collection string
	client     string
	datatype   string
}

// NewServerContext creates a new ServerContext
func NewServerContext(ctx gocontext.Context, tag1 string) *ServerContext {
	newCtx := &ServerContext{
		OrtooContext: context.NewOrtooContext(ctx, tag1, ""),
		collection:   "N/A",
		client:       "N/A",
		datatype:     "N/A",
	}
	newCtx.updateLogger()
	return newCtx
}

func (its *ServerContext) CloneWithNewContext(tag1 string) *ServerContext {
	return (&ServerContext{
		OrtooContext: context.NewOrtooContext(gocontext.TODO(), tag1, ""),
		collection:   its.collection,
		client:       its.client,
		datatype:     its.datatype,
	}).updateLogger()
}

func (its *ServerContext) UpdateCollection(collection string) *ServerContext {
	its.collection = collection
	return its.updateLogger()
}

func (its *ServerContext) UpdateClient(client string) *ServerContext {
	its.client = client
	return its.updateLogger()
}

func (its *ServerContext) UpdateDatatype(datatype string) *ServerContext {
	its.datatype = datatype
	return its.updateLogger()
}

func (its *ServerContext) updateLogger() *ServerContext {
	its.UpdateTags(its.L().GetTag1(), fmt.Sprintf("%s|%s|%s", its.collection, its.client, its.datatype))
	return its
}
