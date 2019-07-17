package commons

import (
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOperationID(t *testing.T) {
	uniqueID, err := newUniqueID()
	if err != nil {
		t.Error("fail")
	}
	log.Logger.Infof("%#v %s", uniqueID, uniqueID.String())
	pb, err := uniqueID.getPb()
	if err != nil {
		t.Error("fail to encode protobuf")
	}

	//log.Logger.Infof("%#v", pb)
	uniqueID2, err := newUniqueIDFromPb(pb)
	if err != nil {
		t.Error("fail to decode protobuf")
	}
	log.Logger.Infof("%#v %s", uniqueID, uniqueID2.String())
	require.EqualValues(t, uniqueID.String(), uniqueID2.String())
}
