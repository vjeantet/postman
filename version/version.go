package version

import (
	"fmt"
	"os"
	"path"
	"runtime"
)

const (
	//VERSION const
	VERSION = "0.4.0"
)

var app, version, versionShort string

//App function
func App() string {
	return app
}

//Version function
func Version() string {
	return version
}

//VersionShort function
func VersionShort() string {
	return versionShort
}

func init() {
	app = path.Base(os.Args[0])
	versionShort = fmt.Sprintf("%s/%s", app, VERSION)
	version = fmt.Sprintf("%s/%s %s/%s %s", app, VERSION, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
