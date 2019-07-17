package commons

type DatatypeType uint8

const (
	TypeIntCounter DatatypeType = 1 + iota
	TypeJson
)

func getWiredDatatypeT(d interface{}) *WiredDatatypeT {
	switch v := d.(type) {
	case *IntCounter:
		return v.WiredDatatypeT
	}
	return nil
}
