package local

import (
	"os"
	"path/filepath"

	"github.com/bazuker/browserbro/pkg/fs"
)

type FileStore struct {
	cfg Config
}

type Config struct {
	BasePath string
}

func New(cfg Config) (*FileStore, error) {
	if _, err := os.Stat(cfg.BasePath); os.IsNotExist(err) {
		err := os.MkdirAll(cfg.BasePath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return &FileStore{
		cfg: cfg,
	}, nil
}

func (f *FileStore) PutObject(object []byte, key string) error {
	path := filepath.Join(f.cfg.BasePath, key)
	return os.WriteFile(path, object, os.ModePerm)
}

func (f *FileStore) GetObject(key string) ([]byte, error) {
	if _, err := os.Stat(filepath.Join(f.cfg.BasePath, key)); os.IsNotExist(err) {
		return nil, fs.ErrorFileNotFound
	}
	path := filepath.Join(f.cfg.BasePath, key)
	return os.ReadFile(path)
}

func (f *FileStore) DeleteObject(key string) error {
	if _, err := os.Stat(filepath.Join(f.cfg.BasePath, key)); os.IsNotExist(err) {
		return fs.ErrorFileNotFound
	}
	path := filepath.Join(f.cfg.BasePath, key)
	return os.Remove(path)
}

func (f *FileStore) BasePath() string {
	return f.cfg.BasePath
}
