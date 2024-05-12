package helper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateRandomString(t *testing.T) {
	values := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key := GenerateRandomString(5)
		_, ok := values[key]
		require.False(t, ok)
		values[key] = true
	}
}
