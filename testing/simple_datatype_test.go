package testing

import (
	"github.com/knowhunger/ortoo/commons"
	. "github.com/knowhunger/ortoo/commons/utils"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SimpleDatatypeSuite struct {
	suite.Suite
}

func (suite *SimpleDatatypeSuite) SetupTest() {
	suite.T().Log("SetupTest")
}

func (suite *SimpleDatatypeSuite) TestExample() {
	tw := commons.NewTestWire()
	intCounter1 := commons.NewIntCounter(tw)
	intCounter2 := commons.NewIntCounter(tw)

	tw.SetDatatypes(intCounter1, intCounter2)

	intCounter1.Increase()

	Log.Printf("%#v", intCounter1)
	suite.T().Log("TestExample")
}

func TestSimpleDatatypeSuite(t *testing.T) {
	suite.Run(t, new(SimpleDatatypeSuite))
}
