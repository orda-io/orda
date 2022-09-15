package svrcontext

import (
	gocontext "context"
	"fmt"
	"github.com/orda-io/orda/client/pkg/context"
	"github.com/orda-io/orda/client/pkg/iface"
)

// ServerContext is used to pass over contexts for server
type ServerContext struct {
	iface.OrdaContext
	collection string
	client     string
	datatype   string
}

// NewServerContext creates a new ServerContext
func NewServerContext(ctx gocontext.Context, tag1 string) *ServerContext {
	newCtx := &ServerContext{
		OrdaContext: context.NewOrdaContext(ctx, tag1, ""),
		collection:  "N/A",
		client:      "N/A",
		datatype:    "N/A",
	}
	newCtx.updateLogger()
	return newCtx
}

// CloneWithNewContext clones a new server context which with different tag1
func (its *ServerContext) CloneWithNewContext(tag1 string) *ServerContext {
	return (&ServerContext{
		OrdaContext: context.NewOrdaContext(gocontext.TODO(), tag1, ""),
		collection:  its.collection,
		client:      its.client,
		datatype:    its.datatype,
	}).updateLogger()
}

// UpdateCollection updates the collection tag
func (its *ServerContext) UpdateCollection(collection string) *ServerContext {
	its.collection = collection
	return its.updateLogger()
}

// UpdateClient updates the client tag
func (its *ServerContext) UpdateClient(client string) *ServerContext {
	its.client = client
	return its.updateLogger()
}

// UpdateDatatype updates the datatype tag
func (its *ServerContext) UpdateDatatype(datatype string) *ServerContext {
	its.datatype = datatype
	return its.updateLogger()
}

func (its *ServerContext) updateLogger() *ServerContext {
	its.UpdateTags(its.L().GetTag1(), fmt.Sprintf("%s|%s|%s", its.collection, its.client, its.datatype))
	return its
}
