package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"strconv"
)

// GetInstanceID returns the INSTANCE_ID environment variable parsed as an unsigned 16-bit integer.
func GetInstanceID() (uint16, error) {
	if instanceID := os.Getenv("INSTANCE_ID"); len(instanceID) > 0 {
		value, err := strconv.ParseUint(instanceID, 10, 16)

		if err != nil {
			log.Fatal(err)
		}

		return uint16(value), nil
	}

	return 0, nil
}

// RandomHexString generates a random hexadecimal string using the specified byte length.
func RandomHexString(byteLength int) string {
	data := make([]byte, byteLength)

	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	return hex.EncodeToString(data)
}

// HashPassword returns an SHA256 encoded string using the input.
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))

	return hex.EncodeToString(hash[:])
}
