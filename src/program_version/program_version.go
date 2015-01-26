package program_version

import (
	"expvar"
)

var version = "UNKNOWN"

func GetVersion() string {
	return version
}

func publishVersion() {
	strVar := expvar.NewString("Version")
	strVar.Set(version)
}
