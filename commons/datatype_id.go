package commons

import "github.com/google/uuid"

type datatypeID struct {
	*uuid.UUID
}

func newDatatypeID() *datatypeID {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil
	}
	return &datatypeID{&u}
}
