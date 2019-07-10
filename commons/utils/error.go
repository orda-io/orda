package utils

import "fmt"

func OrtooError(format string, a ...interface{}) error {
	Log.Errorf(format, a)
	return fmt.Errorf(format, a)
}
