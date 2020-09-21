package testonly

import (
	"github.com/knowhunger/ortoo/pkg/model"
	"strings"
)

// OperationsToString returns a string of an array of operations
func OperationsToString(ops []*model.Operation) string {
	sb := strings.Builder{}
	sb.WriteString("[ ")
	for i, op := range ops {
		sb.WriteString(op.ToString())
		if len(ops)-1 != i {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(" ]")
	return sb.String()
}
