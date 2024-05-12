package mock

type FileStore struct {
	PutObjectFn    func(object []byte, key string) error
	GetObjectFn    func(key string) ([]byte, error)
	DeleteObjectFn func(key string) error
	BasePathValue  string
}

func (m *FileStore) PutObject(object []byte, key string) error {
	if m.PutObjectFn == nil {
		return nil
	}
	return m.PutObjectFn(object, key)
}

func (m *FileStore) GetObject(key string) ([]byte, error) {
	if m.GetObjectFn == nil {
		return nil, nil
	}
	return m.GetObjectFn(key)
}

func (m *FileStore) DeleteObject(key string) error {
	if m.DeleteObjectFn == nil {
		return nil
	}
	return m.DeleteObjectFn(key)
}

func (m *FileStore) BasePath() string {
	return m.BasePathValue
}
