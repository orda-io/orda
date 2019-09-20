package errors

//ErrTransaction is an error related to transaction
type ErrTransaction struct {
}

func (t *ErrTransaction) Error() string {
	return "transaction error"
}

//ErrLinkDatatype is an error related to linking datatype
type ErrLinkDatatype struct {
}

func (t *ErrLinkDatatype) Error() string {
	return "fail to link datatype to Ortoo server"
}
