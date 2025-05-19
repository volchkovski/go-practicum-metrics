package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	actual, err := hex.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("incorrect hash format: %w", err)
	}
	expectedHash := hs.Hash(data)
	expected, err := hex.DecodeString(expectedHash)
	if err != nil {
		return false, fmt.Errorf("failed to generate hash: %w", err)
	}
	return hmac.Equal(actual, expected), nil
}
