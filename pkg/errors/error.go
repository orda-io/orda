package errors

// ErrorCode is a type for error code of OrtooError
type ErrorCode uint32

// OrtooError defines an OrtooError
type OrtooError interface {
	error
	GetCode() ErrorCode
}

// OrtooErrorImpl implements an error related to Ortoo
type OrtooErrorImpl struct {
	Code ErrorCode
	Msg  string
}

func (d *OrtooErrorImpl) Error() string {
	return d.Msg
}

// GetCode returns ErrorCode
func (d *OrtooErrorImpl) GetCode() ErrorCode {
	return d.Code
}

// ToOrtooError casts an error to OrtooError if it is a OrtooError type
func ToOrtooError(err error) OrtooError {
	if dErr, ok := err.(OrtooError); ok {
		return dErr
	}
	return nil
}
