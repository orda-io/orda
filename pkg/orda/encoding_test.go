package orda

import (
	"github.com/orda-io/orda/pkg/iface"
	"github.com/orda-io/orda/pkg/log"
	"github.com/orda-io/orda/pkg/model"
	"github.com/orda-io/orda/pkg/operations"
	"github.com/orda-io/orda/pkg/testonly"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOperationEncoding(t *testing.T) {
	t.Run("Test encoding counter operations", func(t *testing.T) {
		base := testonly.NewBase("key1", model.TypeOfDatatype_COUNTER)
		counter1, _ := newCounter(base, nil, nil)
		counter1.IncreaseBy(1024)

		gOp1, err := operations.NewSnapshotOperationFromDatatype(counter1.(iface.Datatype))

		require.NoError(t, err)
		log.Logger.Infof("%v", testonly.Marshal(t, gOp1))

		mOp1 := gOp1.ToModelOperation()
		log.Logger.Infof("%v", mOp1)
		gmOp1 := operations.ModelToOperation(mOp1)

		require.Equal(t, testonly.Marshal(t, gOp1), testonly.Marshal(t, gmOp1))

		counter2, _ := newCounter(base, nil, nil)
		counter2.(iface.Datatype).ExecuteRemote(gOp1)

		gOp2, err := operations.NewSnapshotOperationFromDatatype(counter2.(iface.Datatype))
		require.NoError(t, err)
		log.Logger.Infof("%v", testonly.Marshal(t, gOp2))
		require.Equal(t, testonly.Marshal(t, gOp1), testonly.Marshal(t, gOp2))
	})
}
