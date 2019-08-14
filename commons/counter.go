package commons

import "github.com/knowhunger/ortoo/commons/model"

type Counter struct {
	parent *OrtooData
}

func (c *Counter) SetOrtooData(o *OrtooData) {
	c.parent = o
}

func (c *Counter) Increase() {
	io := model.NewIncreaseOperation(1)
	c.parent.execute(io)
}
