package errors

//ErrTransaction is an error related to transaction
type ErrTransaction struct {
}

func (t *ErrTransaction) Error() string {
	return "transaction error"
}

//ErrSubscribeDatatype is an error related to linking datatype
type ErrSubscribeDatatype struct {
}

func (t *ErrSubscribeDatatype) Error() string {
	return "fail to subscribe datatype to Ortoo server"
}
