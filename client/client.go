package client

import "github.com/knowhunger/ortoo/commons"

type ClientT struct {
	db             string
	collectionName string
	clientId       *commons.Cuid
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
