package testonly

import (
	"fmt"
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

func (suite *SimpleDatatypeSuite) TestTransactionFail() {
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
	intCounter1.DoTransaction("transaction1", func(intCounter commons.IntCounterInTransaction) error {
		intCounter.IncreaseBy(2)
		suite.Equal(int32(2), intCounter.Get())
		intCounter.IncreaseBy(4)
		suite.Equal(int32(6), intCounter.Get())
		return nil
	})
	suite.Equal(int32(6), intCounter1.Get())
	intCounter1.DoTransaction("transaction2", func(intCounter commons.IntCounterInTransaction) error {
		intCounter.IncreaseBy(3)
		suite.Equal(int32(9), intCounter.Get())
		intCounter.IncreaseBy(5)
		suite.Equal(int32(14), intCounter.Get())
		return fmt.Errorf("err")
	})
	suite.Equal(int32(6), intCounter1.Get())
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
	//var wg = new(sync.WaitGroup)
	//wg.Add(2)
	intCounter1.DoTransaction("transaction1", func(intCounter commons.IntCounterInTransaction) error {
		//defer wg.Done()
		intCounter.IncreaseBy(-1)
		suite.Equal(int32(0), intCounter.Get())
		intCounter.IncreaseBy(-2)
		intCounter.IncreaseBy(-3)
		suite.Equal(int32(-5), intCounter.Get())
		return nil
	})

	log.Logger.Printf("%#v", intCounter1.Get())
	log.Logger.Printf("%#v", intCounter2.Get())
	suite.Equal(intCounter1.Get(), intCounter2.Get())

	intCounter1.DoTransaction("transaction2", func(intCounter commons.IntCounterInTransaction) error {
		//defer wg.Done()
		intCounter.IncreaseBy(1)
		intCounter.IncreaseBy(2)
		suite.Equal(int32(-2), intCounter.Get())
		intCounter.IncreaseBy(3)
		intCounter.IncreaseBy(4)
		return log.OrtooError(nil, "fail to transaction")
	})
	//
	////wg.Wait()
	log.Logger.Printf("%#v", intCounter1.Get())
	log.Logger.Printf("%#v", intCounter2.Get())
	suite.Equal(intCounter1.Get(), intCounter2.Get())
}

func (suite *SimpleDatatypeSuite) TestPushPullPackSync() {
	intCounter1, err := commons.NewIntCounter(nil, datatypes.NewDummyWire())
	if err != nil {
		suite.Fail("fail to create intCounter1")
	}
	intCounter1.Increase()
	intCounter1.Increase()
	//ppp := intCounter1.CreatePushPullPack()
	//log.Logger.Info(proto.MarshalTextString(ppp))

}

func TestSimpleDatatypeSuite(t *testing.T) {
	suite.Run(t, new(SimpleDatatypeSuite))
}
