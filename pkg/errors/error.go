package errors

import (
	"fmt"
)

// ErrorCode is a type for error code of OrtooError
type ErrorCode uint32

const (
	// ErrMultiple is an error code that includes many OrtooErrors
	ErrMultiple = iota + 1
)

// OrtooError defines an OrtooError
type OrtooError interface {
	error
	GetCode() ErrorCode
	Append(e OrtooError) OrtooError
	Return() OrtooError
	Have(code ErrorCode) int
	ToArray() []OrtooError
	Size() int
}

// singleOrtooError implements an error related to Ortoo
type singleOrtooError struct {
	Code ErrorCode
	Msg  string
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
	return its.Msg
}

// GetCode returns ErrorCode
func (its *singleOrtooError) GetCode() ErrorCode {
	return its.Code
}

func (its *singleOrtooError) Return() OrtooError {
	return its
}

func (its *singleOrtooError) Append(e OrtooError) OrtooError {
	var codes []ErrorCode
	var msgs []string
	switch cast := e.(type) {
	case *singleOrtooError:
		codes = append(codes, its.Code, cast.Code)
		msgs = append(msgs, its.Msg, cast.Msg)
	case *MultipleOrtooErrors:
		codes = append(codes, its.Code)
		codes = append(codes, cast.Codes...)
		msgs = append(msgs, its.Msg)
		msgs = append(msgs, cast.Msgs...)
	}
	return &MultipleOrtooErrors{
		Codes: codes,
		Msgs:  msgs,
	}
}

// ToOrtooError casts an error to OrtooError if it is a OrtooError type
func ToOrtooError(err error) OrtooError {
	if dErr, ok := err.(OrtooError); ok {
		return dErr
	}
	return nil
}

type MultipleOrtooErrors struct {
	Codes []ErrorCode
	Msgs  []string
}

func (its *MultipleOrtooErrors) Size() int {
	return len(its.Codes)
}

func (its *MultipleOrtooErrors) ToArray() []OrtooError {
	var errs []OrtooError
	for i, code := range its.Codes {
		errs = append(errs, &singleOrtooError{
			Code: code,
			Msg:  its.Msgs[i],
		})
	}
	return errs
}

func (its *MultipleOrtooErrors) Have(code ErrorCode) int {
	cnt := 0
	for _, e := range its.Codes {
		if e == code {
			cnt++
		}
	}
	return cnt
}

func (its *MultipleOrtooErrors) Error() string {
	return fmt.Sprintf("%+q", its.Msgs)
}

func (its *MultipleOrtooErrors) GetCode() ErrorCode {
	return ErrMultiple
}

func (its *MultipleOrtooErrors) Append(e OrtooError) OrtooError {
	switch cast := e.(type) {
	case *singleOrtooError:
		its.Codes = append(its.Codes, cast.Code)
		its.Msgs = append(its.Msgs, cast.Msg)
	case *MultipleOrtooErrors:
		its.Codes = append(its.Codes, cast.Codes...)
		its.Msgs = append(its.Msgs, cast.Msgs...)
	}
	return its
}

func (its *MultipleOrtooErrors) Return() OrtooError {
	if len(its.Codes) > 0 {
		return its
	}
	return nil
}
