package schema

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

const ordaPatchAPI string = "!@#$OrdaPatchAPI"

var administrators = map[string][16]byte{
	ordaPatchAPI: md5.Sum([]byte(ordaPatchAPI)),
}

// NewPatchClient creates a new patch client for each collection
func NewPatchClient(collectionDoc *CollectionDoc) *ClientDoc {

	alias := administrators[ordaPatchAPI]
	return &ClientDoc{
		CUID:          ordaPatchAPI,
		Alias:         hex.EncodeToString(alias[:]),
		CollectionNum: collectionDoc.Num,
		Type:          0,
		SyncType:      0,
		CheckPoints:   nil,
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
	}
}

// IsAdmin returns true if the client is admin
func (its *ClientDoc) IsAdmin() bool {
	if alias, ok := administrators[its.CUID]; ok {
		return its.Alias == hex.EncodeToString(alias[:])
	}
	return false
}
