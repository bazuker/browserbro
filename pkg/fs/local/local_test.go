package local

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	dirName := t.TempDir() + "/browserBro_files"
	fs, err := New(Config{
		BasePath: dirName,
	})
	require.NoError(t, err)
	assert.NotNil(t, fs)

	_, err = os.Stat(dirName)
	assert.NoError(t, err)

	assert.Equal(t, dirName, fs.cfg.BasePath)
}

func TestFileStore_BasePath(t *testing.T) {
	dirName := t.TempDir() + "/browserBro_files"
	fs, err := New(Config{
		BasePath: dirName,
	})
	require.NoError(t, err)
	assert.NotNil(t, fs)

	assert.Equal(t, dirName, fs.BasePath())
}

func TestFileStore_PutGetDeleteObject(t *testing.T) {
	dirName := t.TempDir() + "/browserBro_files"
	fs, err := New(Config{
		BasePath: dirName,
	})
	require.NoError(t, err)
	assert.NotNil(t, fs)

	fileName := "test.txt"
	filePath := path.Join(dirName, fileName)
	fileContent := []byte("test")

	err = fs.PutObject(fileContent, fileName)
	require.NoError(t, err)
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	content, err := fs.GetObject(fileName)
	require.NoError(t, err)
	assert.Equal(t, fileContent, content)

	err = fs.DeleteObject(fileName)
	require.NoError(t, err)
	_, err = os.Stat(filePath)
	assert.Error(t, err)
}
