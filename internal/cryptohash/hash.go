package cryptohash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func HashSHA256(data string, strKey string) (hash string) {
	var emptyByte string
	if strKey == "" {
		return emptyByte
	}

	key := []byte(strKey)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	hash = fmt.Sprintf("%x", h.Sum(nil))
	return
}
