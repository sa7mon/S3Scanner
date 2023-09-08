package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sa7mon/s3scanner/collection"
	"github.com/sa7mon/s3scanner/provider"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
)

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

func GetRegionsDO() ([]string, error) {
	requestURL := "https://docs.digitalocean.com/products/platform/availability-matrix/#other-product-availability"
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

	// Find the review items
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

type linodeRegionsResp struct {
	Data []struct {
		ID           string   `json:"id"`
		Label        string   `json:"label"`
		Country      string   `json:"country"`
		Capabilities []string `json:"capabilities"`
		Status       string   `json:"status"`
		Resolvers    struct {
			Ipv4 string `json:"ipv4"`
			Ipv6 string `json:"ipv6"`
		} `json:"resolvers"`
	} `json:"data"`
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	Results int `json:"results"`
}

func GetRegionsLinode() ([]string, error) {
	requestURL := "https://api.linode.com/v4/regions"
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	var r linodeRegionsResp
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	objectStorageRegions := []string{}
	for _, d := range r.Data {
		objectStorageSupport := false
		for _, dc := range d.Capabilities {
			if dc == "Object Storage" {
				objectStorageSupport = true
				break
			}
		}
		if objectStorageSupport {
			objectStorageRegions = append(objectStorageRegions, d.ID)
		}
	}

	return objectStorageRegions, nil
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

	for p, knownRegions := range provider.ProviderRegions {
		if errors[p] != nil {
			log.Printf("[%s]: %v\n", p, errors[p])
			continue
		}
		foundRegions := results[p]
		sort.Strings(foundRegions)
		sort.Strings(knownRegions)

		if !eq(foundRegions, knownRegions) {
			log.Printf("[%s] regions differ! Existing: %v, found; %v", p, knownRegions, foundRegions)
		} else {
			log.Printf("[%s} OK", p)
		}
	}
}
