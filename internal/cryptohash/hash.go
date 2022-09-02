package cryptohash

import (
	"crypto/hmac"
	"crypto/sha256"
)

func HeshSHA256(val string, key string) []byte {
	var dst []byte
	if key == "" {
		return dst
	}

	src := []byte(val)
	b := []byte(key)

	h := hmac.New(sha256.New, b)
	h.Write(src)
	dst = h.Sum(nil)

	return dst
}
