//go:build !windows && !unix

package isadmin

func IsAdmin() bool {
	return false
}
