package selenium

import (
	"log"
)

var debugFlag = false

// SetDebug sets debug mode
func SetDebug(debug bool) {
	debugFlag = debug
}

func debugLog(format string, args ...interface{}) {
	if !debugFlag {
		return
	}
	log.Printf(format+"\n", args...)
}
