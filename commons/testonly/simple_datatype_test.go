package testonly

import (
	"github.com/golang/protobuf/proto"
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/stretchr/testify/suite"

	"testing"
)

type SimpleDatatypeSuite struct {
	suite.Suite
}

func (suite *SimpleDatatypeSuite) SetupTest() {
	suite.T().Log("SetupTest")
}

func (suite *SimpleDatatypeSuite) TestOneOperationSyncWithTestWire() {
	tw := NewTestWire()
	intCounter1, err := commons.NewIntCounter(nil, tw)
	if err != nil {
		suite.Fail("fail to create intCounter1")
	}
	intCounter2, err := commons.NewIntCounter(nil, tw)
	if err != nil {
		suite.Fail("fail to create intCounter2")
	}

	tw.SetDatatypes(intCounter1, intCounter2)

	i, err := intCounter1.Increase()
	if err != nil {
		suite.Fail("fail to increase")
	}
	suite.Equal(i, int32(1))
	intCounter1.DoTransaction(func(datatype interface{}) bool {

		return true
	})
	log.Logger.Printf("%#v", intCounter1)
	log.Logger.Printf("%#v", intCounter2)
	suite.Equal(intCounter1.Get(), intCounter2.Get())
}

func (suite *SimpleDatatypeSuite) TestPushPullPackSync() {
	intCounter1, err := commons.NewIntCounter(nil, datatypes.NewDummyWire())
	if err != nil {
		suite.Fail("fail to create intCounter1")
	}
	intCounter1.Increase()
	intCounter1.Increase()
	ppp := intCounter1.CreatePushPullPack()
	log.Logger.Info(proto.MarshalTextString(ppp))

}

func TestSimpleDatatypeSuite(t *testing.T) {
	suite.Run(t, new(SimpleDatatypeSuite))
}
