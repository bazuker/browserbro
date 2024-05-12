package screenshot

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bazuker/browserbro/pkg/fs"
	"github.com/bazuker/browserbro/pkg/manager/helper"
	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

const (
	pluginName = "screenshot"
)

type BotCheck struct {
	maxTimePerScreenshot time.Duration
	browser              *rod.Browser
	fileStore            fs.FileStore
}

func New(browser *rod.Browser, fileStore fs.FileStore) *BotCheck {
	return &BotCheck{
		maxTimePerScreenshot: 15 * time.Second,
		browser:              browser,
		fileStore:            fileStore,
	}
}

func (p *BotCheck) Name() string {
	return pluginName
}

func (p *BotCheck) Run(params map[string]any) (output map[string]any, err error) {

	var stealthPage *rod.Page
	stealthPage, err = stealth.Page(p.browser)
	if err != nil {
		return nil, fmt.Errorf("failed to create stealth page: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to complete: %v", r)
		}
		_ = stealthPage.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	page := stealthPage.Context(ctx)

	urlsList, ok := params["urls"].([]any)
	if !ok {
		cancel()
		return nil, errors.New("'urls' parameter must be an array of string")
	}
	if len(urlsList) == 0 {
		cancel()
		return nil, errors.New("empty 'urls' parameter")
	}
	waitStable, ok := params["waitStable"].(bool)
	if !ok {
		waitStable = true
	}

	go func() {
		time.Sleep(p.maxTimePerScreenshot * time.Duration(len(urlsList)))
		cancel()
	}()

	screenshots := make([]string, 0)
	for _, url := range urlsList {
		urlString, ok := url.(string)
		if !ok {
			return nil, errors.New("'urls' parameter must only contain strings")
		}

		err = page.Navigate(urlString)
		if err != nil {
			return nil, fmt.Errorf("failed to navigate to the page '%s': %w", urlString, err)
		}

		page.MustWaitLoad()
		if waitStable {
			page.MustWaitStable()
		}

		filename := helper.GenerateRandomString(6) + ".screenshot.png"
		page.MustScreenshotFullPage(filepath.Join(p.fileStore.BasePath(), filename))

		screenshots = append(screenshots, filename)
	}

	output = make(map[string]any)
	output["files"] = screenshots

	return output, err
}
