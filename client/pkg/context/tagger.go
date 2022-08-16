package context

import (
	"fmt"
	"github.com/orda-io/orda/client/pkg/utils"
)

const (
	maxTestName   = 15
	maxServerName = 15
)

// MakeTagInServer creates a tag for server
func MakeTagInServer(host string, port int) string {
	return fmt.Sprintf("%s:%d", utils.MakeShort(host, maxServerName), port)
}

// MakeTagInTest creates a tag for test
func MakeTagInTest(test string) string {
	return utils.MakeShort(test, maxTestName)
}
