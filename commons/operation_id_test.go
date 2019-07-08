package commons

import (
	"gotest.tools/assert"
	"testing"
)

func TestClientId(t *testing.T) {
	opID1 := newOperationID()
	opID2 := newOperationID()
	assert.Assert(t, Compare(opID1, opID2) == 0)
	opID1.Next()
	assert.Assert(t, Compare(opID1, opID2) > 0)
}
