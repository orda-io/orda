package model

import (
	"gotest.tools/assert"
	"testing"
)

func TestClientId(t *testing.T) {
	opID1 := NewOperationID()
	opID2 := NewOperationID()
	assert.Assert(t, opID1.Compare(opID2) == 0)
	opID1.Next()
	assert.Assert(t, opID1.Compare(opID2) > 0)
}
