package selenium

import (
	"log"
	"net/url"
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

// filteredURL replaces existing password from the given URL.
func filteredURL(u string) string {
	// Hide password if set in URL
	m, err := url.Parse(u)
	if err != nil {
		return ""
	}
	if m.User != nil {
		if _, ok := m.User.Password(); ok {
			m.User = url.UserPassword(m.User.Username(), "__password__")
		}
	}
	return m.String()
}
