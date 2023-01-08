package cli

import "errors"

// AttachToParentConsole is a stub for a function available only on the Windows platform
func AttachToParentConsole() (attached bool, err error) {
	return false, errors.New("AttachToParentConsole is only available on windows")
}
