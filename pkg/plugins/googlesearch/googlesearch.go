package googlesearch

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

const (
	pluginName = "googlesearch"
)

type GoogleSearch struct {
	browser          *rod.Browser
	maxTimePerSearch time.Duration
}

func New(browser *rod.Browser) *GoogleSearch {
	return &GoogleSearch{
		browser:          browser,
		maxTimePerSearch: 15 * time.Second,
	}
}

func (p *GoogleSearch) Name() string {
	return pluginName
}

func (p *GoogleSearch) Run(params map[string]any) (output map[string]any, err error) {
	query, ok := params["query"]
	if !ok {
		return nil, errors.New("missing 'query' parameter")
	}
	queryString, ok := query.(string)
	if !ok {
		return nil, errors.New("'query' parameter must be a string")
	}

	searchTypesMap := make(map[string]bool)
	searchType, ok := params["type"]
	if ok {
		types, ok := searchType.([]any)
		if !ok {
			return nil, errors.New("'type' parameter must be an array of strings")
		}
		for _, t := range types {
			typeString, ok := t.(string)
			if !ok {
				return nil, errors.New("'type' parameter must only contain strings")
			}
			loweredType := strings.ToLower(typeString)
			switch loweredType {
			case "all", "videos":
				searchTypesMap[loweredType] = true
			default:
				return nil, errors.New("invalid search type: " + typeString)
			}
		}
	} else {
		searchTypesMap["all"] = true
	}

	var stealthPage *rod.Page
	stealthPage, err = stealth.Page(p.browser)
	if err != nil {
		return nil, fmt.Errorf("failed to create stealth page: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	page := stealthPage.Context(ctx)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to complete: %v", r)
		}
		_ = page.Close()
	}()

	go func() {
		time.Sleep(p.maxTimePerSearch)
		cancel()
	}()

	output = make(map[string]any)
	urlParams := url.Values{}
	urlParams.Set("q", queryString)

	if searchTypesMap["all"] || searchTypesMap["videos"] {
		searchKey := "all"
		if searchTypesMap["videos"] {
			urlParams.Set("tbm", "vid")
			searchKey = "videos"
		}
		err = page.Navigate("https://www.google.com/search?" + urlParams.Encode())
		if err != nil {
			return nil, fmt.Errorf("failed to navigate to 'all' page: %w", err)
		}
		err = page.WaitLoad()
		if err != nil {
			return nil, fmt.Errorf("failed to wait for page to load: %w", err)
		}
		blocks, err := page.Elements(".g")
		if err != nil {
			return nil, fmt.Errorf("failed to parse search results: %w", err)
		}
		searchResults := make([]map[string]any, 0, len(blocks))
		for _, block := range blocks {
			href, err := block.Element("a")
			if err != nil {
				continue
			}
			link, err := href.Attribute("href")
			if err != nil {
				continue
			}
			h3, err := block.Element("h3")
			if err != nil {
				continue
			}
			title, err := h3.Text()
			if err != nil {
				continue
			}
			spans, err := block.Elements("span")
			if err != nil {
				continue
			}
			description, err := spans[len(spans)-1].Text()
			if err != nil {
				description = ""
			}
			searchResults = append(searchResults, map[string]any{
				"link":        link,
				"title":       title,
				"description": description,
			})
		}
		output[searchKey] = searchResults
	}

	return output, nil
}
