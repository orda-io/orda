package errors

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/log"
	"strings"

	"github.com/ztrue/tracerr"
)

type tError tracerr.Error

// OrdaError defines an OrdaError
type OrdaError interface {
	tError
	GetCode() ErrorCode
	Append(e OrdaError) OrdaError
	Return() OrdaError
	Have(code ErrorCode) int
	ToArray() []OrdaError
	Size() int
	Print(l *log.OrdaLog, skip int)
}

// singleOrdaError implements an error related to Orda
type singleOrdaError struct {
	tError
	Code ErrorCode
}

func (its *singleOrdaError) Size() int {
	return 1
}

func (its *singleOrdaError) ToArray() []OrdaError {
	return []OrdaError{its}
}

func (its *singleOrdaError) Have(code ErrorCode) int {
	if its.Code == code {
		return 1
	}
	return 0
}

func (its *singleOrdaError) Error() string {
	return its.tError.Error()
}

// GetCode returns ErrorCode
func (its *singleOrdaError) GetCode() ErrorCode {
	return its.Code
}

func (its *singleOrdaError) Return() OrdaError {
	return its
}

func (its *singleOrdaError) Append(e OrdaError) OrdaError {
	var errs []*singleOrdaError
	switch cast := e.(type) {
	case *singleOrdaError:
		errs = append(errs, its, cast)
	case *MultipleOrdaErrors:
		errs = append(errs, its)
		errs = append(errs, cast.errs...)
	}
	return &MultipleOrdaErrors{
		tError: tracerr.New("Multiple OrdaErrors"),
		errs:   errs,
	}
}

func (its *singleOrdaError) Print(l *log.OrdaLog, skip int) {
	if l == nil {
		return
	}
	var sb strings.Builder
	sb.WriteString(its.tError.Error())
	for _, frame := range its.StackTrace()[skip:] {
		sb.WriteString("\n\t")
		sb.WriteString(frame.String())
	}
	l.Error(sb.String())
}

// ToOrdaError casts an error to OrdaError if it is a OrdaError type
func ToOrdaError(err error) OrdaError {
	if dErr, ok := err.(OrdaError); ok {
		return dErr
	}
	return nil
}

func newSingleOrdaError(l *log.OrdaLog, code ErrorCode, name string, msgFormat string, args ...interface{}) OrdaError {
	format := fmt.Sprintf("[%s: %d] %s", name, code, msgFormat)
	err := &singleOrdaError{
		tError: tracerr.New(fmt.Sprintf(format, args...)),
		Code:   code,
	}
	err.Print(l, 2)
	return err
}
