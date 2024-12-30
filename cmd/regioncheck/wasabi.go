package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func getRegionsWasabi() ([]string, error) {
	requestURL := "https://wasabi.com/company/storage-regions"
	res, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	regions := []string{}
	doc.Find("main h3~p:has(b)").Contents().Each(func(_ int, t *goquery.Selection) {
		if goquery.NodeName(t) == "#text" {
			rangeNames := strings.Split(t.Text(), "&")
			for _, r := range rangeNames {
				regions = append(regions, strings.TrimSpace(r))
			}
		}
	})
	return regions, nil
}
