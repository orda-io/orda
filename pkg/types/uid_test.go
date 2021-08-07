package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/orda-io/orda/pkg/log"
)

func TestUID(t *testing.T) {
	t.Run("Can validate UIDs", func(t *testing.T) {
		uid := NewUID()
		log.Logger.Infof("%s", uid)
		assert.True(t, ValidateUID(uid))

		invalidUID1 := "X"
		assert.False(t, ValidateUID(invalidUID1))

		invalidUID2 := "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
		assert.False(t, ValidateUID(invalidUID2))

		var cnt int32 = 0
		for i := 0; i < 100; i++ {
			uid1 := NewUID()
			uid2 := NewUID()
			require.NotEqual(t, uid1, uid2)
			log.Logger.Infof("%v vs %v : %v", uid1, uid2, uid1 < uid2)
			if uid1 < uid2 {
				cnt++
			}
		}

		log.Logger.Info(cnt)
	})

	t.Run("Can compare strings consistently", func(t *testing.T) {
		// In the all SDKs, these comparison should be consistent
		uid1 := "123abc"
		uid2 := "abc123"
		uid3 := "ABC123"
		uid4 := "abc1234"
		require.True(t, strings.Compare(uid1, uid1) == 0)
		require.True(t, strings.Compare(uid1, uid2) == -1)
		require.True(t, strings.Compare(uid2, uid3) == 1)
		require.True(t, strings.Compare(uid2, uid4) == -1)
	})
}
