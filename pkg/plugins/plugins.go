package plugins

import (
	"github.com/bazuker/browserbro/pkg/fs"
	"github.com/bazuker/browserbro/pkg/plugins/googlesearch"
	"github.com/bazuker/browserbro/pkg/plugins/screenshot"
	"github.com/go-rod/rod"
)

type Plugin interface {
	Name() string
	Run(Params map[string]any) (map[string]any, error)
}

var (
	customPlugins = make([]Plugin, 0)
)

// Initialize initializes the default plugins and custom plugins.
func Initialize(browser *rod.Browser, fileStore fs.FileStore) []Plugin {
	defaultPlugins := []Plugin{
		googlesearch.New(browser),
		screenshot.New(browser, fileStore),
	}
	return append(defaultPlugins, customPlugins...)
}

// AddCustom adds a custom plugin to the list of plugins.
func AddCustom(plugin Plugin) {
	if plugin == nil {
		return
	}
	customPlugins = append(customPlugins, plugin)
}
