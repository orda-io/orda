package model

import (
	"fmt"
	"strings"
)

// ToString returns a summary of the operation.
func (its *Operation) ToString() string {
	sb := strings.Builder{}
	_, _ = fmt.Fprintf(&sb, "%s|%s|%s", its.OpType.String(), its.ID.ToString(), string(its.Body))
	return sb.String()
}

// OpList is a list of *Operation
type OpList []*Operation

// ToString returns string
func (its OpList) ToString() string {
	sb := strings.Builder{}

	sb.WriteString("[ ")
	for _, op := range its {
		sb.WriteString(op.ToString())
		sb.WriteString(" ")
	}
	sb.WriteString("]")
	return sb.String()
}
