package model

import (
	"fmt"
	"strings"
)

func (its *Operation) ToString() string {
	sb := strings.Builder{}
	_, _ = fmt.Fprintf(&sb, "%s|%s|%s", its.OpType.String(), its.ID.ToString(), string(its.Json))
	return sb.String()
}

func OperationsToString(ops []*Operation) string {
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
