package utils

import (
	"encoding/json"
	"github.com/orda-io/orda/client/pkg/log"

	"github.com/TylerBrock/colorjson"
)

// PrintMarshalDoc logs out doc without error
func PrintMarshalDoc(l *log.OrdaLog, doc interface{}) {
	f := colorjson.NewFormatter()
	f.Indent = 2
	f.DisabledColor = true
	m, _ := json.Marshal(doc)
	var obj map[string]interface{}
	_ = json.Unmarshal(m, &obj)
	s, _ := f.Marshal(obj)
	l.Infof("%v", string(s))
}

// ToStringMarshalDoc returns marshaled string without error
func ToStringMarshalDoc(doc interface{}) string {
	m, _ := json.Marshal(doc)
	return string(m)
}
