package googlesearch

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
)

const (
	pluginName = "googlesearch"
)

type GoogleSearch struct {
	browser *rod.Browser
}

func New(browser *rod.Browser) *GoogleSearch {
	return &GoogleSearch{
		browser: browser,
	}
}

func (p *GoogleSearch) Name() string {
	return pluginName
}

func (p *GoogleSearch) Run(params map[string]any) (map[string]any, error) {

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
			case "all", "images", "news", "videos":
				searchTypesMap[loweredType] = true
			default:
				return nil, errors.New("invalid search type: " + typeString)
			}
		}
	} else {
		searchTypesMap["all"] = true
	}

	page, err := stealth.Page(p.browser)
	if err != nil {
		return nil, fmt.Errorf("failed to create stealth page: %w", err)
	}
	defer page.MustClose()

	output := make(map[string]any)
	urlParams := url.Values{}
	urlParams.Set("q", queryString)

	if searchTypesMap["all"] {
		err = page.Navigate("https://www.google.com/search?" + urlParams.Encode())
		if err != nil {
			return nil, fmt.Errorf("failed to navigate to 'all' page: %w", err)
		}
		page.MustWaitLoad()

		blocks := page.MustElements(".g")
		searchResults := make([]map[string]any, 0, len(blocks))
		for _, block := range blocks {
			link := block.MustElement("a").MustAttribute("href")
			title := block.MustElement("h3").MustText()
			spans := block.MustElements("span")
			description := spans[len(spans)-1].MustText()
			searchResults = append(searchResults, map[string]any{
				"link":        link,
				"title":       title,
				"description": description,
			})
		}
		output["all"] = searchResults
	}

	if searchTypesMap["videos"] {
		urlParams.Set("tbm", "vid")
		err = page.Navigate("https://www.google.com/search?" + urlParams.Encode())
		if err != nil {
			return nil, fmt.Errorf("failed to navigate to 'videos' page: %w", err)
		}
		page.MustWaitLoad()

		blocks := page.MustElements(".g")
		searchResults := make([]map[string]any, 0, len(blocks))
		for _, block := range blocks {
			link := block.MustElement("a").MustAttribute("href")
			title := block.MustElement("h3").MustText()
			spans := block.MustElements("span")
			description := spans[len(spans)-1].MustText()
			searchResults = append(searchResults, map[string]any{
				"link":        link,
				"title":       title,
				"description": description,
			})
		}
		output["videos"] = searchResults
	}

	return output, nil
}
