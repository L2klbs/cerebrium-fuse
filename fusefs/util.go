package fusefs

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0444)
}

// hashContent returns the SHA-256 hash of a byte slice as a string
func hashContent(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}