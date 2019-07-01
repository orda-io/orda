package commons

import "github.com/google/uuid"

type Cuid struct {
	uuid *uuid.UUID
}

func NewCuid() *Cuid {
	uuid := uuid.New()
	return &Cuid{
		&uuid,
	}
}

func NewNilCuid() *Cuid {
	uuid := uuid.Nil
	return &Cuid{
		&uuid,
	}
}
