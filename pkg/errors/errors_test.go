package errors_test

import (
	"github.com/knowhunger/ortoo/pkg/errors"
	"github.com/knowhunger/ortoo/pkg/log"
	"testing"
)

func TestOrtooError(t *testing.T) {
	_ = call1()
}

func call1() errors.OrtooError {
	return call2()
}

func call2() errors.OrtooError {
	return errors.DatatypeNoOp.New(log.Logger, "sample error")
}
