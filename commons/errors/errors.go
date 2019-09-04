package errors

//NewTransactionError creates a new TransactionError
func NewTransactionError() *TransactionError {
	return &TransactionError{}
}

//TransactionError is an error regarding to transaction
type TransactionError struct {
}

func (t *TransactionError) Error() string {
	return "transaction error"
}
