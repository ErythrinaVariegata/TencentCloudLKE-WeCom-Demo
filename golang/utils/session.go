package utils

import (
	"crypto/rand"
	"fmt"

	"github.com/google/uuid"
)

const sessionIDLength = 24

func GetRequestID() string {
	b := make([]byte, sessionIDLength)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", b)
}

func GetSessionID() string {
	return uuid.NewString()
}
