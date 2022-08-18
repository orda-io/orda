package model

import (
	"fmt"
	"strings"
)

// ToString returns a string of the operation with its body.
func (its *Operation) ToString() string {
	return fmt.Sprintf("%s|%s|%v", its.OpType.String(), its.ID.ToString(), string(its.Body))
}

// ToShortString returns a string of the operation without its body
func (its *Operation) ToShortString() string {
	return fmt.Sprintf("%s|%s", its.OpType.String(), its.ID.ToString())
}

// OpList is a list of *Operation
type OpList []*Operation

// ToString returns string
func (its OpList) ToString(isFull bool) string {
	sb := strings.Builder{}

	sb.WriteString("[ ")
	for _, op := range its {
		if isFull {
			sb.WriteString(op.ToString())
		} else {
			sb.WriteString(op.ToShortString())
		}
		sb.WriteString(" ")
	}
	sb.WriteString("]")
	return sb.String()
}
