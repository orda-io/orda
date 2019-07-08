package commons

import "fmt"

type operationer interface {
	executeLocal()
	executeRemote()
}

type operation struct {
	id        operationID
	typ       OpType
	timestamp timestamp
}

func (c *operation) executeLocal() {
	fmt.Println("operation executeLocal")
}

func (c *operation) executeRemote() {

}
