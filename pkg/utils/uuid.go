package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID returns a random 8-char hex string
func GenerateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
