package errors

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/log"
	"strings"
)

// MultipleOrdaErrors is used to manage multiple OrdaErrors
type MultipleOrdaErrors struct {
	tError
	errs []*singleOrdaError
}

// Size returns the size of MultipleOrdaErrors
func (its *MultipleOrdaErrors) Size() int {
	return len(its.errs)
}

// ToArray returns MultipleOrdaErrors to the array of OrdaError
func (its *MultipleOrdaErrors) ToArray() []OrdaError {
	var errs []OrdaError
	for _, e := range its.errs {
		errs = append(errs, e)
	}
	return errs
}

// Have returns the number of errors having the specified error code.
func (its *MultipleOrdaErrors) Have(code ErrorCode) int {
	cnt := 0
	for _, e := range its.errs {
		if e.Code == code {
			cnt++
		}
	}
	return cnt
}

// Error returns the string error message
func (its *MultipleOrdaErrors) Error() string {
	var ret []string
	for _, err := range its.errs {
		ret = append(ret, err.Error())
	}
	return fmt.Sprintf("%+q", ret)
}

// GetCode returns the code
func (its *MultipleOrdaErrors) GetCode() ErrorCode {
	return MultipleErrors
}

// Append adds a new OrdaError to MultipleOrdaErrors
func (its *MultipleOrdaErrors) Append(e OrdaError) OrdaError {
	if e == nil {
		return its
	}
	switch cast := e.(type) {
	case *singleOrdaError:
		its.errs = append(its.errs, cast)
	case *MultipleOrdaErrors:
		its.errs = append(its.errs, cast.errs...)
	}
	return its
}

// Return returns itself as OrdaError
func (its *MultipleOrdaErrors) Return() OrdaError {
	if len(its.errs) > 0 {
		return its
	}
	return nil
}

// Print prints out the concatenated errors
func (its *MultipleOrdaErrors) Print(l *log.OrdaLog, skip int) {
	var sb strings.Builder
	sb.WriteString(its.tError.Error())
	for _, frame := range its.StackTrace()[skip:] {
		sb.WriteString("\n\t")
		sb.WriteString(frame.String())
	}
	l.Error(sb.String())
}
