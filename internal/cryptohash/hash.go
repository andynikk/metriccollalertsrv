package cryptohash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func HeshSHA256(data string, strKey string) string {
	if strKey == "" {
		return ""
	}

	key := []byte(strKey)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))

}
