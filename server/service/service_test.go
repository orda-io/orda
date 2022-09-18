package service_test

import (
	gocontext "context"
	"github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/client/pkg/orda"
	"github.com/orda-io/orda/server/service"
	"github.com/orda-io/orda/server/testonly"
	"github.com/orda-io/orda/server/wrapper"
	integration "github.com/orda-io/orda/test"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestOrdaService(t *testing.T) {

	_, ctx, _ := integration.InitTestDBCollection(t, testonly.TestDBName)
	managers, err := integration.NewTestManagers(ctx, testonly.TestDBName)
	svc := service.NewOrdaService(managers)
	require.NoError(t, err)
	collectionName := t.Name()

	conf := &orda.ClientConfig{
		CollectionName: collectionName,
		SyncType:       model.SyncType_MANUALLY,
	}

	t.Run("Can avoid to create admin client", func(t *testing.T) {
		cm := &model.Client{
			CUID:       "!@#$OrdaPatchAPI",
			Alias:      "",
			Collection: "test",
			Type:       0,
			SyncType:   0,
		}
		req := model.NewClientMessage(cm)
		res, err := svc.ProcessClient(ctx, req)
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), errors.ServerNoPermission.New(nil, "").Error()))
		require.Nil(t, res)
	})

	t.Run("Can do ", func(t *testing.T) {

		client1 := orda.NewClient(conf, t.Name()+"1")
		counter1 := client1.SubscribeOrCreateCounter(t.Name(), nil)

		wrapper1 := wrapper.NewDatatypeWrapper(counter1)
		_, _ = counter1.Increase()
		testonly.RegisterClient(t, svc, wrapper1.GetClientModel())

		res, err := svc.ProcessPushPull(gocontext.TODO(), wrapper1.CreatePushPullMessage())
		require.NoError(t, err)
		ppp1 := res.GetPushPullPacks()[0]
		require.True(t, ppp1.CheckPoint.Compare(model.NewSetCheckPoint(2, 2)))
		require.Equal(t, len(ppp1.Operations), 0)
		require.Equal(t, ppp1.GetOption(), uint32(model.PushPullBitCreate))
		log.Logger.Infof("%v", ppp1.ToString(false))
		wrapper1.ApplyPushPullPack(ppp1)

		client2 := orda.NewClient(conf, t.Name()+"2")
		counter2 := client2.SubscribeCounter(t.Name(), nil)
		wrapper2 := wrapper.NewDatatypeWrapper(counter2)
		testonly.RegisterClient(t, svc, wrapper2.GetClientModel())

		res, err = svc.ProcessPushPull(gocontext.TODO(), wrapper2.CreatePushPullMessage())
		require.NoError(t, err)
		ppp2 := res.GetPushPullPacks()[0]
		log.Logger.Infof("%v", ppp2.ToString(false))
		require.True(t, ppp2.CheckPoint.Compare(model.NewSetCheckPoint(2, 0)))
		require.Equal(t, len(ppp2.Operations), 2)
		require.Equal(t, ppp2.GetOption(), uint32(model.PushPullBitSubscribe))
		wrapper2.ApplyPushPullPack(ppp2)

		_, _ = counter2.Increase()
		_, _ = counter2.Increase()

		res, err = svc.ProcessPushPull(gocontext.TODO(), wrapper2.CreatePushPullMessage())
		require.NoError(t, err)
		ppp3 := res.GetPushPullPacks()[0]
		log.Logger.Infof("%v", ppp3.ToString(false))
		require.True(t, ppp3.CheckPoint.Compare(model.NewSetCheckPoint(4, 2)))
		require.Equal(t, len(ppp3.Operations), 0)
		require.Equal(t, ppp3.GetOption(), uint32(model.PushPullBitNormal))
	})
}
