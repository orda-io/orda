package admin

import (
	"github.com/orda-io/orda/client/pkg/model"
	"github.com/orda-io/orda/server/schema"
	"time"
)

const ordaPatchAPICUID string = "!@#$OrdaPatchAPI"

var administrators = map[string]string{
	ordaPatchAPICUID: "ordaPatchAPI",
}

// NewPatchClient creates a new patch client for each collection
func NewPatchClient(collectionDoc *schema.CollectionDoc) *schema.ClientDoc {

	alias := administrators[ordaPatchAPICUID]
	return &schema.ClientDoc{
		CUID:          ordaPatchAPICUID,
		Alias:         alias,
		CollectionNum: collectionDoc.Num,
		Type:          int8(model.ClientType_VOLATILE),
		SyncType:      0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// IsAdminCUID returns true if the client is admin
func IsAdminCUID(cuid string) bool {
	if _, ok := administrators[cuid]; ok {
		return true
	}
	return false
}
