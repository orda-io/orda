package errors_test

import (
	errors2 "github.com/orda-io/orda/client/pkg/errors"
	"github.com/orda-io/orda/client/pkg/log"
	"testing"
)

func TestOrdaError(t *testing.T) {
	_ = call1()
}

func call1() errors2.OrdaError {
	return call2()
}

func call2() errors2.OrdaError {
	return errors2.DatatypeNoOp.New(log.Logger, "This is a generated sample error")
}
