package cryptohash

import (
	"crypto/hmac"
	"crypto/sha256"
)

func HeshSHA256(data string, strKey string) []byte {
	var emtyByte []byte
	if strKey == "" {
		return emtyByte
	}

	key := []byte(strKey)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	//return fmt.Sprintf("%x", h.Sum(nil))
	return h.Sum(nil)

}
