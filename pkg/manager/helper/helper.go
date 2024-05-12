package helper

import (
	"crypto/rand"
	"encoding/base64"
)

const (
	ContextFileStore = "fileStore"
)

type HTTPMessage struct {
	Message string `json:"message"`
}

type SessionData struct {
	UserID      string
	AccessLevel string
}

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
