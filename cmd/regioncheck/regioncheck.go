package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sa7mon/s3scanner/provider"
	"io"
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
	doc.Find("h3#other-products + table thead tr th").Each(func(_ int, t *goquery.Selection) {
		regions = append(regions, t.Text())
	})

	spacesSupported := []bool{}
	doc.Find("h3#other-products + table tbody tr").Each(func(_ int, t *goquery.Selection) {
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

// GetRegionsLinode fetches region names from Linode docs HTML page. Linode also provides this info via
// unauthenticated API (https://api.linode.com/v4/regions) but the region names do not include the trailing digit "-1".
func GetRegionsLinode() ([]string, error) {
	requestURL := "https://techdocs.akamai.com/cloud-computing/docs/object-storage"

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	// server drops requests if user agent is set to the default go-http-client:
	// `HTTP/2 stream 1 was not closed cleanly: INTERNAL_ERROR (err 2)`
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")

	client := &http.Client{}
	res, err := client.Do(req)
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
	doc.Find(".rdmd-table:nth-of-type(1) tbody tr td:nth-of-type(2)").Each(func(_ int, t *goquery.Selection) {
		regions = append(regions, t.Text())
	})

	return regions, nil
}

func GetRegionsScaleway() ([]string, error) {
	var re = regexp.MustCompile(`Region: \x60(.+)\x60`)
	requestURL := "https://raw.githubusercontent.com/scaleway/docs-content/main/storage/object/how-to/create-a-bucket.mdx"
	res, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	bytes, bErr := io.ReadAll(res.Body)
	if bErr != nil {
		return nil, bErr
	}

	var regions []string
	for _, a := range re.FindAllSubmatch(bytes, -1) {
		regions = append(regions, string(a[1]))
	}
	return regions, nil
}

func main() {
	results := map[string][]string{}
	errors := map[string]error{}

	p := map[string]func() ([]string, error){
		"digitalocean": GetRegionsDO,
		"linode":       GetRegionsLinode,
		"scaleway":     GetRegionsScaleway,
	}

	wg := sync.WaitGroup{}
	wg.Add(len(p))

	for name, get := range p {
		name := name
		get := get
		go func(_ *sync.WaitGroup) {
			results[name], errors[name] = get()
			wg.Done()
		}(&wg)
	}

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
			log.Printf("[%s] OK", p)
		}
	}
	os.Exit(exit)
}
