package integration

import (
	"github.com/knowhunger/ortoo/commons"
	"testing"
)

func TestStruct(t *testing.T) {
	d := commons.NewOrtooData()
	d.GetCounter().Increase()
}
