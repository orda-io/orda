package testonly

import (
	"github.com/knowhunger/ortoo/commons"
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

func (suite *SimpleDatatypeSuite) TestExample() {
	tw := NewTestWire()
	intCounter1, err := commons.NewIntCounter(tw)
	if err != nil {
		_ = log.OrtooError(err, "fail to create intCounter1")
	}
	intCounter2, err := commons.NewIntCounter(tw)
	if err != nil {
		_ = log.OrtooError(err, "fail to create intCounter2")
	}

	tw.SetDatatypes(intCounter1, intCounter2)

	intCounter1.Increase()

	log.Logger.Printf("%#v", intCounter1)
	log.Logger.Printf("%#v", intCounter2)
	suite.T().Log("TestExample")
}

func TestSimpleDatatypeSuite(t *testing.T) {
	suite.Run(t, new(SimpleDatatypeSuite))
}
