package types

// CUID is a uniqueID for a client.
type CUID UID

// NewCUID creates a new CUID
func NewCUID() CUID {
	return CUID(newUniqueID())
}

// NewNilCUID creates an instance of Nil CUID.
func NewNilCUID() CUID {
	bin := make([]byte, 16)
	return bin
}

func (its *CUID) String() string {
	return UIDtoString(*its)
}

// ShortString returns a short string.
func (its *CUID) ShortString() string {
	return UIDtoShortString(*its)
}
