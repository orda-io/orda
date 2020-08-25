package errors

type ErrorCode uint32

type OrtooError interface {
	error
	GetCode() ErrorCode
}

// OrtooError is an error related to Datatype
type OrtooErrorImpl struct {
	Code ErrorCode
	Msg  string
}

func (d *OrtooErrorImpl) Error() string {
	return d.Msg
}

func (d *OrtooErrorImpl) GetCode() ErrorCode {
	return d.Code
}

func ToOrtooError(err error) OrtooError {
	if dErr, ok := err.(OrtooError); ok {
		return dErr
	}
	return nil
}
