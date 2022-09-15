package testonly

import (
	ctx "context"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/server/service"
	"github.com/stretchr/testify/require"
	"testing"
)

// RegisterClient is used to test client
func RegisterClient(t *testing.T, service *service.OrdaService, client *model.Client) {

	req := model.NewClientMessage(client)
	res, err := service.ProcessClient(ctx.TODO(), req)
	require.NoError(t, err)
	log.Logger.Infof("ProcessClient result:%v", res)
}
