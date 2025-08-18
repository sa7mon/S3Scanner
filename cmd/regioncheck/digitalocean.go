package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetRegionsDO fetches regions from the DigitalOcean docs HTML page.
func GetRegionsDO() ([]string, error) {
	requestURL := "https://docs.digitalocean.com/platform/regional-availability/"
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

	var regions []string
	doc.Find("h2#other-digitalocean-products + div table thead tr th").Each(func(_ int, t *goquery.Selection) {
		if t.Text() != "Product" {
			regions = append(regions, t.Text())
		}
	})

	var supportedRegions []string
	doc.Find("h2#other-digitalocean-products + div table tbody tr").Each(func(_ int, t *goquery.Selection) {
		// For each row, check the first cell for a value of "Spaces"
		rowHeader := t.Find("td").First().Text()
		if rowHeader == "Spaces" {
			// For each cell in the "Spaces" row, check if the contents are not empty - meaning Spaces is supported
			t.Find("td").Each(func(i int, v *goquery.Selection) {
				if v.Has("i.fa-circle").Length() != 0 {
					supportedRegions = append(supportedRegions, strings.ToLower(regions[i-1]))
				}
			})
		}
	})

	return supportedRegions, nil
}
