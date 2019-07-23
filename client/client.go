package client

import (
	"github.com/knowhunger/ortoo/commons/model"
)

type ClientT struct {
	db             string
	collectionName string
	clientId       *model.Cuid
}

func (c *ClientT) connect() {

}

func (c *ClientT) createDatatype() {

}

type Client interface {
	connect()
	createDatatype()
}

func NewClient() {

}
