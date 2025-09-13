package isadmin

import (
	"golang.org/x/sys/windows"
)

func IsAdmin() bool {
	var sid *windows.SID
	// The administrators group SID is well-known
	sid, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return false
	}
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	return err == nil && member
}
