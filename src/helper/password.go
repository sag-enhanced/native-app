package helper

import "golang.org/x/crypto/argon2"

func DeriveKey(password string, salt []byte) []byte {
	memory := uint32(64 * 1024)
	time := uint32(3)
	threads := uint8(4)
	bytes := uint32(32)

	return argon2.IDKey([]byte(password), salt, time, memory, threads, bytes)
}
