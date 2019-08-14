package commons

import (
	"github.com/knowhunger/ortoo/commons/internal/data"
	"github.com/knowhunger/ortoo/commons/model"
)

type OrtooData struct {
	tnx  *data.Trans
	data OrtooDataInterface
}

func NewOrtooData() *OrtooData {
	c := &Counter{}
	ortoo := &OrtooData{
		tnx:  &data.Trans{},
		data: c,
	}
	ortoo.data.SetOrtooData(ortoo)
	return ortoo
}

type OrtooDataInterface interface {
	SetOrtooData(o *OrtooData)
}

func (o *OrtooData) GetCounter() *Counter {
	c, ok := o.data.(*Counter)
	if !ok {
		return nil
	}
	return c
}

func (o *OrtooData) execute(m model.Operation) {

}
