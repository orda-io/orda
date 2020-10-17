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
