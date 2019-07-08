package commons

import "github.com/google/uuid"

type Cuid struct {
	uuid *uuid.UUID
}

func newCuid() *Cuid {
	uuid := uuid.New()
	return &Cuid{
		&uuid,
	}
}

func newNilCuid() *Cuid {
	uuid := uuid.Nil
	return &Cuid{
		&uuid,
	}
}
