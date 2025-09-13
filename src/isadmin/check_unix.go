//go:build unix

package isadmin

import "os"

func IsAdmin() bool {
	return os.Geteuid() == 0
}
