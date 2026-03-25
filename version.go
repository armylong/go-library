package library

import (
	"regexp"
	"runtime"
)

var version string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	version = `unknown`
	if ok {
		compile, _ := regexp.Compile(`/library@v([\d\\.]+)/`)
		match := compile.FindStringSubmatch(filename)
		if len(match) > 1 {
			version = match[1]
		}
	}
}

func Version() string {
	return version
}
