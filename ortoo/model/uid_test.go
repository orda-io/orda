package model

import (
	"encoding/hex"
	"github.com/knowhunger/ortoo/ortoo/log"
	"testing"
)

func TestOperationID(t *testing.T) {
	t.Run("DUID string to DUID", func(t *testing.T) {
		duid, err := NewDUID()
		if err != nil {
			t.Fatal(err)
		}

		str := hex.EncodeToString(duid)
		log.Logger.Infof("%s", str)
		toDUID, err := DUIDFromString(str)
		if err != nil {
			t.Fatal(err)
		}
		if CompareUID(UniqueID(duid), UniqueID(toDUID)) != 0 {
			t.Fatal(err)
		}
	})
}
