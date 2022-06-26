package constants

import (
	"io/ioutil"
	"strings"
)

var (
	// GitCommit is a git commit hash which is injected by Makefile
	GitCommit = "unspecified"

	// Version is a version string which is injected by Makefile
	Version = "unknown"
)

func init() {
	if Version == "unknown" {
		if read, err := ioutil.ReadFile("version"); err == nil {
			Version = strings.TrimSpace(string(read))
		}
	}

}
