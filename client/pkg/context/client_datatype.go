package context

import (
	"context"
	"github.com/orda-io/orda/client/pkg/constants"
	"github.com/orda-io/orda/client/pkg/iface"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
)

// ClientContext is used to pass over the context of clients
type ClientContext struct {
	iface.OrdaContext
	Client *model.Client
}

// NewClientContext creates a new ClientContext
func NewClientContext(ctx context.Context, client *model.Client) *ClientContext {
	return &ClientContext{
		OrdaContext: NewOrdaContextWithAllTags(ctx, constants.TagSdkClient, client.Collection, "", client.Alias, client.CUID, "", ""),
		Client:      client,
	}
}

// DatatypeContext is used to pass over the context of datatypes
type DatatypeContext struct {
	*ClientContext
	Data iface.BaseDatatype
}

// NewDatatypeContext creates a new DatatypeContext
func NewDatatypeContext(clientContext *ClientContext, baseDatatype iface.BaseDatatype) *DatatypeContext {
	logger := log.NewWithTags(constants.TagSdkDatatype,
		clientContext.Client.Collection, "",
		clientContext.Client.Alias, clientContext.Client.CUID,
		baseDatatype.GetKey(), baseDatatype.GetCUID())
	return &DatatypeContext{
		ClientContext: &ClientContext{
			OrdaContext: &ordaContext{
				Context: clientContext.Ctx(),
				logger:  logger,
			},
			Client: clientContext.Client,
		},
		Data: baseDatatype,
	}
}
