package errors

func NewTransactionError() *TransactionError {
	return &TransactionError{}
}

type TransactionError struct {
}

func (t *TransactionError) Error() string {
	return "transaction error"
}
