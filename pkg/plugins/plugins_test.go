package plugins

import (
	"github.com/bazuker/browserbro/pkg/fs/local"
	"github.com/go-rod/rod"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	defer func() {
		customPlugins = []Plugin{}
	}()
	browser := &rod.Browser{}
	fs := &local.FileStore{}
	require.Len(t, Initialize(browser, fs), 2)
	AddCustom(&mockPlugin{name: "test"})
	require.Len(t, Initialize(browser, fs), 3)
}

func TestAddCustom(t *testing.T) {
	defer func() {
		customPlugins = []Plugin{}
	}()
	require.Len(t, customPlugins, 0)
	AddCustom(&mockPlugin{name: "test"})
	require.Len(t, customPlugins, 1)
	AddCustom(nil)
	require.Len(t, customPlugins, 1)
}

type mockPlugin struct {
	name  string
	runFn func(params map[string]interface{}) (map[string]interface{}, error)
}

func (mp *mockPlugin) Name() string {
	return mp.name
}

func (mp *mockPlugin) Run(params map[string]interface{}) (
	map[string]interface{},
	error,
) {
	if mp.runFn == nil {
		return nil, nil
	}
	return mp.runFn(params)
}
