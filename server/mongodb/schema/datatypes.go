package schema

import "time"

type DatatypeDoc struct {
	Duid string `bson:"_id"`
	Key  string `bson:"_id"`

	Type      string    `bson:"type"`
	Sseq      uint64    `bson:"sseq"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}
