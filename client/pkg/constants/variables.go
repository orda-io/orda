package constants

import (
	"fmt"
	"os"
	"runtime/debug"
)

// BuildInfo is a git commit hash which is injected by Makefile
var BuildInfo = func() string {
	var goVer, hash, os, arch = "", "", "", ""
	if info, ok := debug.ReadBuildInfo(); ok {
		goVer = info.GoVersion
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				hash = setting.Value[:7]
			} else if setting.Key == "GOOS" {
				os = setting.Value
			} else if setting.Key == "GOARCH" {
				arch = setting.Value
			}
		}
		return fmt.Sprintf("%s-%s-%s-%s", goVer, os, arch, hash)
	}
	return ""
}()

// SDKType is a version string which is injected by Makefile
var SDKType = func() string {
	if env, ok := os.LookupEnv("ORDA_SDK_TYPE"); ok {
		return env
	}
	return "go"
}()
