package commons

import "github.com/knowhunger/ortoo/commons/internal/datatypes"

func getWiredDatatype(datatype interface{}) datatypes.WiredDatatype {
	switch t := datatype.(type) {
	case intCounterImpl:
		return t.WiredDatatypeImpl
	}
	return nil
}
