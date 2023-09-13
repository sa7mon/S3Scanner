package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sa7mon/s3scanner/collection"
	"github.com/sa7mon/s3scanner/provider"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// eq compares sorted string slices. Once we move to Golang 1.21, use slices.Equal instead.
func eq(f []string, s []string) bool {
	if len(f) != len(s) {
		return false
	}
	for i := range f {
		if f[i] != s[i] {
			return false
		}
	}
	return true
}

// GetRegionsDO fetches regions from the DigitalOcean docs HTML page.
func GetRegionsDO() ([]string, error) {
	requestURL := "https://docs.digitalocean.com/products/platform/availability-matrix/#other-product-availability"
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
	doc.Find("h2#other-product-availability + table thead tr th").Each(func(i int, t *goquery.Selection) {
		regions = append(regions, t.Text())
	})

	spaces_supported := []bool{}
	doc.Find("h2#other-product-availability + table tbody tr").Each(func(i int, t *goquery.Selection) {
		// For each row, check the first cell for a value of "Spaces"
		rowHeader := t.Find("td").First().Text()
		if rowHeader == "Spaces" {
			// For each cell in the "Spaces" row, check if the contents are not empty - meaning Spaces is supported
			t.Find("td").Each(func(j int, v *goquery.Selection) {
				supported := v.Text() != ""
				spaces_supported = append(spaces_supported, supported)
			})
		}
	})

	supported_regions := []string{}
	for i := 0; i < len(regions); i++ {
		if regions[i] == "Product" {
			continue
		}
		if spaces_supported[i] {
			supported_regions = append(supported_regions, strings.ToLower(regions[i]))
		}
	}

	// Return slice of region names
	return supported_regions, nil
}

// GetRegionsLinode fetches region names from Linode docs HTML page. Linode also provides this info via
// unauthenticated API (https://api.linode.com/v4/regions) but the region names do not include the trailing digit "-1".
func GetRegionsLinode() ([]string, error) {
	requestURL := "https://www.linode.com/docs/products/storage/object-storage/"
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
	doc.Find("h2#availability + p + table tbody tr td:nth-of-type(2)").Each(func(i int, t *goquery.Selection) {
		regions = append(regions, t.Text())
	})

	return regions, nil
}

// GetRegionsDreamhost fetches subdomains of dream.io like 'objects-us-east-1.dream.io' via crt.sh since Dreamhost
// doesn't have a documentation page listing the regions.
func GetRegionsDreamhost() ([]string, error) {
	var domainRe = regexp.MustCompile(`objects-([^\.]+)\.dream\.io`)
	requestURL := "https://crt.sh/?q=.dream.io"
	// Request the HTML page.
	res, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	certCNs := collection.StringSet{}
	// For each cell in the Common Name column
	doc.Find("body > table table tbody tr > td:nth-of-type(5)").Each(func(i int, t *goquery.Selection) {
		matches := domainRe.FindAllStringSubmatch(t.Text(), -1)
		for _, match := range matches {
			if !strings.HasPrefix(match[1], "website-") { // regions like 'objects-website-us-east-1' are not for Object Storage
				certCNs.Add(match[1])
			}
		}
	})

	return certCNs.Slice(), nil
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(3)

	results := map[string][]string{}
	errors := map[string]error{}

	go func(w *sync.WaitGroup) {
		results["digitalocean"], errors["digitalocean"] = GetRegionsDO()
		wg.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		results["dreamhost"], errors["dreamhost"] = GetRegionsDreamhost()
		wg.Done()
	}(&wg)
	go func(w *sync.WaitGroup) {
		results["linode"], errors["linode"] = GetRegionsLinode()
		wg.Done()
	}(&wg)
	wg.Wait()

	exit := 0

	for p, knownRegions := range provider.ProviderRegions {
		if errors[p] != nil {
			log.Printf("[%s]: %v\n", p, errors[p])
			continue
		}
		foundRegions := results[p]
		sort.Strings(foundRegions)
		sort.Strings(knownRegions)

		if !eq(foundRegions, knownRegions) {
			log.Printf("[%s] regions differ! Existing: %v, found: %v", p, knownRegions, foundRegions)
			exit = 1
		} else {
			log.Printf("[%s} OK", p)
		}
	}
	os.Exit(exit)
}
