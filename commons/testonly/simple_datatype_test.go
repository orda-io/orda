package testonly

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons"
	"github.com/knowhunger/ortoo/commons/internal/datatypes"
	"github.com/knowhunger/ortoo/commons/log"
	"github.com/knowhunger/ortoo/commons/model"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SimpleDatatypeSuite struct {
	suite.Suite
}

func (suite *SimpleDatatypeSuite) SetupTest() {

	log.Logger.Infof("Set up test")
}

func (suite *SimpleDatatypeSuite) TestTransactionFail() {
	tw := NewTestWire()
	cuid1, err := model.NewCUID()
	if err != nil {
		suite.T().Fatal(err)
	}
	intCounter1, err := commons.NewIntCounter("key1", cuid1, tw)
	if err != nil {
		suite.Fail("fail to create intCounter1")
	}
	cuid2, err := model.NewCUID()
	if err != nil {
		suite.T().Fatal(err)
	}
	intCounter2, err := commons.NewIntCounter("key1", cuid2, tw)
	if err != nil {
		suite.Fail("fail to create intCounter2")
	}

	tw.SetDatatypes(intCounter1, intCounter2)
	if err := intCounter1.DoTransaction("transaction1", func(intCounter commons.IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(2)
		suite.Equal(int32(2), intCounter.Get())
		_, _ = intCounter.IncreaseBy(4)
		suite.Equal(int32(6), intCounter.Get())
		return nil
	}); err != nil {
		suite.T().Fatal(err)
	}

	suite.Equal(int32(6), intCounter1.Get())

	if err := intCounter1.DoTransaction("transaction2", func(intCounter commons.IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(3)
		suite.Equal(int32(9), intCounter.Get())
		_, _ = intCounter.IncreaseBy(5)
		suite.Equal(int32(14), intCounter.Get())
		return fmt.Errorf("err")
	}); err == nil {
		suite.T().FailNow()
	}
	suite.Equal(int32(6), intCounter1.Get())
}

func (suite *SimpleDatatypeSuite) TestOneOperationSyncWithTestWire() {
	tw := NewTestWire()
	intCounter1, err := commons.NewIntCounter("key1", model.NewNilCUID(), tw)
	if err != nil {
		suite.T().Fatal(err)
	}
	intCounter2, err := commons.NewIntCounter("key2", model.NewNilCUID(), tw)
	if err != nil {
		suite.T().Fatal(err)
	}

	tw.SetDatatypes(intCounter1, intCounter2)

	i, err := intCounter1.Increase()
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.Equal(i, int32(1))
	suite.Equal(intCounter1.Get(), intCounter2.Get())

	if err := intCounter1.DoTransaction("transaction1", func(intCounter commons.IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(-1)
		suite.Equal(int32(0), intCounter.Get())
		_, _ = intCounter.IncreaseBy(-2)
		_, _ = intCounter.IncreaseBy(-3)
		suite.Equal(int32(-5), intCounter.Get())
		return nil
	}); err != nil {
		suite.T().Fatal(err)
	}

	log.Logger.Printf("%#v vs. %#v", intCounter1.Get(), intCounter2.Get())
	suite.Equal(intCounter1.Get(), intCounter2.Get())

	if err := intCounter1.DoTransaction("transaction2", func(intCounter commons.IntCounterInTxn) error {
		_, _ = intCounter.IncreaseBy(1)
		_, _ = intCounter.IncreaseBy(2)
		suite.Equal(int32(-2), intCounter.Get())
		_, _ = intCounter.IncreaseBy(3)
		_, _ = intCounter.IncreaseBy(4)
		return log.OrtooErrorf(nil, "fail to do the transaction")
	}); err == nil {
		suite.T().Fatal(err)
	}
	log.Logger.Printf("%#v vs. %#v", intCounter1.Get(), intCounter2.Get())
	suite.Equal(intCounter1.Get(), intCounter2.Get())
}

func (suite *SimpleDatatypeSuite) TestPushPullPackSync() {
	intCounter1, err := commons.NewIntCounter("key1", model.NewNilCUID(), datatypes.NewDummyWire())
	if err != nil {
		suite.Fail("fail to create intCounter1")
	}
	intCounter1.Increase()
	intCounter1.Increase()
	// ppp := intCounter1.CreatePushPullPack()
	// log.Logger.Info(proto.MarshalTextString(ppp))

}

func TestSimpleDatatypeSuite(t *testing.T) {
	suite.Run(t, new(SimpleDatatypeSuite))
}
