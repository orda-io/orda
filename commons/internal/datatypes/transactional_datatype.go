package datatypes

type transactionalDatatypeImpl struct {
	*baseDatatypeImpl
}

type TransactionalDatatype interface {
}

type PublicTransactionalInterface interface {
	DoTransaction(func(datatype interface{}) error)
}

func (t *transactionalDatatypeImpl) DoTransaction(transFunc func(datatype interface{}) error) {
	// lock
	err := transFunc(t.GetOpExecuter())
	if err != nil {
		// rollback
	}

	// unlocak
}
