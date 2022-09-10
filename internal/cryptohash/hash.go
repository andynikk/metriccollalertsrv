package cryptohash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func HeshSHA256(data string, strKey string) (hash string) {
	var emtyByte string
	if strKey == "" {
		return emtyByte
	}

	key := []byte(strKey)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	hash = fmt.Sprintf("%x", h.Sum(nil))
	return

}
