package errors

import (
	"fmt"
)

// ErrorCode is a type for error code of OrtooError
type ErrorCode uint32

const (
	ErrSingle = iota
	ErrMultiple
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

// SingleOrtooError implements an error related to Ortoo
type SingleOrtooError struct {
	Code ErrorCode
	Msg  string
}

func (its *SingleOrtooError) Size() int {
	return 1
}

func (its *SingleOrtooError) ToArray() []OrtooError {
	return []OrtooError{its}
}

func (its *SingleOrtooError) Have(code ErrorCode) int {
	if its.Code == code {
		return 1
	}
	return 0
}

func (its *SingleOrtooError) Error() string {
	return its.Msg
}

// GetCode returns ErrorCode
func (its *SingleOrtooError) GetCode() ErrorCode {
	return its.Code
}

func (its *SingleOrtooError) Return() OrtooError {
	return its
}

func (its *SingleOrtooError) Append(e OrtooError) OrtooError {
	var codes []ErrorCode
	var msgs []string
	switch cast := e.(type) {
	case *SingleOrtooError:
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
		errs = append(errs, &SingleOrtooError{
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
	case *SingleOrtooError:
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
