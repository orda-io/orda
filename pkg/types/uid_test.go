package types

import (
	"github.com/knowhunger/ortoo/pkg/log"
	"testing"
)

func TestOperationID(t *testing.T) {
	t.Run("DUID string to DUID", func(t *testing.T) {
		duid := NewDUID()
		str := duid.String()
		log.Logger.Infof("%s", str)
		toDUID, err := DUIDFromString(str)
		if err != nil {
			t.Fatal(err)
		}
		if CompareUID(UID(duid), UID(toDUID)) != 0 {
			t.Fatal(err)
		}
	})
}
