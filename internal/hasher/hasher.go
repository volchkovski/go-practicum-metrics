// Package hasher provides HMAC-SHA256 hashing functionality
// for securing HTTP communications between agent and server.
package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const HashHeaderKey = "HashSHA256"

type Hasher struct {
	key []byte
}

func New(key string) *Hasher {
	return &Hasher{[]byte(key)}
}

func (hs *Hasher) Hash(data []byte) string {
	h := hmac.New(sha256.New, hs.key)
	h.Write(data)
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}

func (hs *Hasher) Validate(data []byte, hash string) (bool, error) {
	expectedHash := hs.Hash(data)
	return hmac.Equal([]byte(hash), []byte(expectedHash)), nil
}
