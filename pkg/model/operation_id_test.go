package model

import (
	"testing"

	"gotest.tools/assert"
)

func TestClientId(t *testing.T) {
	t.Run("Can compare OperationIDs", func(t *testing.T) {
		opID1 := NewOperationID()
		opID2 := NewOperationID()
		assert.Assert(t, opID1.Compare(opID2) == 0)
		opID1.Next()
		assert.Assert(t, opID1.Compare(opID2) > 0)
	})
}
