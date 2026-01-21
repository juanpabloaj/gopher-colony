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

// HashString converts a string to an int64 hash.
func HashString(s string) int64 {
	var h int64
	for _, c := range s {
		h = 31*h + int64(c)
	}
	return h
}
