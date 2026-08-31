package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// HashSHA256 returns the hex-encoded SHA-256 checksum of an input string.
func HashSHA256(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

// VerifySHA256 checks if a raw string matches an expected hex hash in constant time.
func VerifySHA256(rawInput, expectedHexHash string) bool {
	computed := HashSHA256(rawInput)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(expectedHexHash)) == 1
}

// GenerateRandomBytes reads n cryptographically secure random bytes from crypto/rand.
func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomHex generates a secure random hex string of length 2*n.
func GenerateRandomHex(n uint32) (string, error) {
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateClientID creates a unique client identifier (e.g., "srv_a81f09c2...").
func GenerateClientID() (string, error) {
	hexStr, err := GenerateRandomHex(16)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("srv_%s", hexStr), nil
}

// GenerateClientSecret generates a raw secret, its lookup prefix, and SHA-256 hash.
func GenerateClientSecret() (rawSecret string, prefix string, hash string, err error) {
	hexStr, err := GenerateRandomHex(32)
	if err != nil {
		return "", "", "", err
	}

	rawSecret = fmt.Sprintf("sec_live_%s", hexStr)
	prefix = rawSecret[:10] // e.g., "sec_live_a"
	hash = HashSHA256(rawSecret)

	return rawSecret, prefix, hash, nil
}
