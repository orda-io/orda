package errors

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/log"
	"github.com/ztrue/tracerr"
	"strings"
)

type tError tracerr.Error

// OrtooError defines an OrtooError
type OrtooError interface {
	tError
	GetCode() ErrorCode
	Append(e OrtooError) OrtooError
	Return() OrtooError
	Have(code ErrorCode) int
	ToArray() []OrtooError
	Size() int
	Print(l *log.OrtooLog)
}

// singleOrtooError implements an error related to Ortoo
type singleOrtooError struct {
	tError
	Code ErrorCode
}

func (its *singleOrtooError) Size() int {
	return 1
}

func (its *singleOrtooError) ToArray() []OrtooError {
	return []OrtooError{its}
}

func (its *singleOrtooError) Have(code ErrorCode) int {
	if its.Code == code {
		return 1
	}
	return 0
}

func (its *singleOrtooError) Error() string {
	return its.tError.Error()
}

// GetCode returns ErrorCode
func (its *singleOrtooError) GetCode() ErrorCode {
	return its.Code
}

func (its *singleOrtooError) Return() OrtooError {
	return its
}

func (its *singleOrtooError) Append(e OrtooError) OrtooError {
	var errs []*singleOrtooError
	switch cast := e.(type) {
	case *singleOrtooError:
		errs = append(errs, its, cast)
	case *MultipleOrtooErrors:
		errs = append(errs, its)
		errs = append(errs, cast.errs...)
	}
	return &MultipleOrtooErrors{
		tError: tracerr.New("Multiple OrtooErrors"),
		errs:   errs,
	}
}

func (its *singleOrtooError) Print(l *log.OrtooLog) {
	var sb strings.Builder
	sb.WriteString(its.tError.Error())
	for _, frame := range its.StackTrace()[1:] {
		sb.WriteString("\n\t")
		sb.WriteString(frame.String())
	}
	l.Error(sb.String())
}

// ToOrtooError casts an error to OrtooError if it is a OrtooError type
func ToOrtooError(err error) OrtooError {
	if dErr, ok := err.(OrtooError); ok {
		return dErr
	}
	return nil
}

type MultipleOrtooErrors struct {
	tError
	errs []*singleOrtooError
}

func (its *MultipleOrtooErrors) Size() int {
	return len(its.errs)
}

func (its *MultipleOrtooErrors) ToArray() []OrtooError {
	var errs []OrtooError
	for _, e := range its.errs {
		errs = append(errs, e)
	}
	return errs
}

func (its *MultipleOrtooErrors) Have(code ErrorCode) int {
	cnt := 0
	for _, e := range its.errs {
		if e.Code == code {
			cnt++
		}
	}
	return cnt
}

func (its *MultipleOrtooErrors) Error() string {
	var ret []string
	for _, err := range its.errs {
		ret = append(ret, err.Error())
	}
	return fmt.Sprintf("%+q", ret)
}

func (its *MultipleOrtooErrors) GetCode() ErrorCode {
	return MultipleErrors
}

func (its *MultipleOrtooErrors) Append(e OrtooError) OrtooError {
	switch cast := e.(type) {
	case *singleOrtooError:
		its.errs = append(its.errs, cast)
	case *MultipleOrtooErrors:
		its.errs = append(its.errs, cast.errs...)
	}
	return its
}

func (its *MultipleOrtooErrors) Return() OrtooError {
	if len(its.errs) > 0 {
		return its
	}
	return nil
}

func (its *MultipleOrtooErrors) Print(l *log.OrtooLog) {
	var sb strings.Builder
	sb.WriteString(its.tError.Error())
	for _, frame := range its.StackTrace()[1:] {
		sb.WriteString("\n\t")
		sb.WriteString(frame.String())
	}
	l.Error(sb.String())
}

func newSingleOrtooError(l *log.OrtooLog, code ErrorCode, name string, msgFormat string, args ...interface{}) OrtooError {
	format := fmt.Sprintf("[%s: %d] %s", name, code, msgFormat)
	err := &singleOrtooError{
		tError: tracerr.New(fmt.Sprintf(format, args...)),
		Code:   code,
	}
	err.Print(l)
	return err
}
