package servicestatus

// true -> ok
// false -> fail
var serviceStatus bool = false

// Return true if service is ok and false if fail
func IsOk() bool {
	return serviceStatus
}

// Returns "OK" or "FAIL"
func GetString() string {
	if IsOk() {
		return "OK"
	} else {
		return "FAIL"
	}
}

// Set OK status
func SetOK() {
	serviceStatus = true
}

// Set FAIL status
func SetFAIL() {
	serviceStatus = false
}
