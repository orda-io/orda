package ortoo

import (
	"fmt"
	"github.com/knowhunger/ortoo/ortoo/model"
	"github.com/knowhunger/ortoo/ortoo/types"
)

type timedValue interface {
	getValue() types.JSONValue
	setValue(v types.JSONValue)
	getTime() *model.Timestamp
	setTime(ts *model.Timestamp)
	makeTomb(ts *model.Timestamp)
	String() string
	isTomb() bool
}

type timedValueImpl struct {
	V types.JSONValue
	T *model.Timestamp
}

func (its *timedValueImpl) getValue() types.JSONValue {
	return its.V
}

func (its *timedValueImpl) makeTomb(ts *model.Timestamp) {
	its.V = nil
	its.T = ts
}

func (its)

func (its *timedValueImpl) setValue(v types.JSONValue) {
	its.V = v
}

func (its *timedValueImpl) getTime() *model.Timestamp {
	return its.T
}

func (its *timedValueImpl) setTime(ts *model.Timestamp) {
	its.T = ts
}

func (its *timedValueImpl) String() string {
	if its.V == nil {
		return fmt.Sprintf("Φ|%s", its.T.ToString())
	}
	return fmt.Sprintf("TV[%v|T%s]", its.V, its.T.ToString())
}
