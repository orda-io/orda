package commons

import (
	"gotest.tools/assert"
	"testing"
)

func TestClientId(t *testing.T) {
	opID1 := NewOperationId()
	opID2 := NewOperationId()
	assert.Assert(t, Compare(opID1, opID2) == 0)
	opID1.Next()
	assert.Assert(t, Compare(opID1, opID2) > 0)
}
