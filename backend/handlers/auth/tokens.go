package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func GenerateResetToken() (string, error) {
	resetToken := make([]byte, 32)

	_, err := rand.Read(resetToken)
	if err != nil {
		return "", errors.New("error creating random slice of bytes")
	}

	encoded := hex.EncodeToString(resetToken)

	return encoded, nil
}

func HashToken(token string) string {
	bytesToken := sha256.Sum256([]byte(token))
	encoded := hex.EncodeToString(bytesToken[:])
	return encoded
}
