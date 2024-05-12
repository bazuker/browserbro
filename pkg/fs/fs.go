package fs

import (
	"errors"
)

type FileStore interface {
	PutObject(object []byte, key string) error
	GetObject(key string) ([]byte, error)
	DeleteObject(key string) error
	BasePath() string
}

var (
	ErrorFileNotFound = errors.New("file not found")
)
