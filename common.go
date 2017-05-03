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

//logURL replaces existing password from the given URL
func logURL(URL string) (string, error) {
	// Hide password if set in URL
	u, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	if u.User != nil {
		_, exists := u.User.Password()
		if exists {
			u.User = url.UserPassword(u.User.Username(),
				"__password__")
		}
	}
	return u.String(), nil
}
