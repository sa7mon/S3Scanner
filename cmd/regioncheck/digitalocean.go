package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
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

	regions := []string{}
	doc.Find("h2#other-digitalocean-products + table thead tr th").Each(func(_ int, t *goquery.Selection) {
		regions = append(regions, t.Text())
	})

	spacesSupported := []bool{}
	doc.Find("h2#other-digitalocean-products + table tbody tr").Each(func(_ int, t *goquery.Selection) {
		// For each row, check the first cell for a value of "Spaces"
		rowHeader := t.Find("td").First().Text()
		if rowHeader == "Spaces" {
			// For each cell in the "Spaces" row, check if the contents are not empty - meaning Spaces is supported
			t.Find("td").Each(func(_ int, v *goquery.Selection) {
				supported := v.Text() != ""
				spacesSupported = append(spacesSupported, supported)
			})
		}
	})

	supportedRegions := []string{}
	for i := 0; i < len(regions); i++ {
		if regions[i] == "Product" {
			continue
		}
		if spacesSupported[i] {
			supportedRegions = append(supportedRegions, strings.ToLower(regions[i]))
		}
	}

	// Return slice of region names
	return supportedRegions, nil
}
