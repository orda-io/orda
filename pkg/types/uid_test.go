package types

import (
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestUID(t *testing.T) {
	t.Run("Validate UID", func(t *testing.T) {
		duid := NewDUID()
		log.Logger.Infof("%s", duid)
		assert.True(t, ValidateUID(duid))

		invalidUID1 := "X"
		assert.False(t, ValidateUID(invalidUID1))

		invalidUID2 := "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
		assert.False(t, ValidateUID(invalidUID2))

		cuid1 := "abc"
		cuid2 := "def"
		assert.True(t, strings.Compare(cuid1, cuid2) < 0)
	})
}
