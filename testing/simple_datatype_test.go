package testing

import (
	"fmt"
	"github.com/knowhunger/ortoo/commons"
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
	intCounter1 := commons.NewIntCounter()
	intCounter1.Increase()
	fmt.Printf("%#v\n", intCounter1)
	suite.T().Log("TestExample")
}

func TestSimpleDatatypeSuite(t *testing.T) {
	suite.Run(t, new(SimpleDatatypeSuite))
}
