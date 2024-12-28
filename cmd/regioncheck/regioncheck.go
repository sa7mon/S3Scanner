package main

import (
	"compress/gzip"
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

// eq compares sorted string slices. TODO: Once we move to Golang 1.21, use slices.Equal instead.
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

func GetRegionsDreamhost() ([]string, error) {
	return []string{"us-east-1"}, nil
}

// GetRegionsLinode fetches region names from Linode docs HTML page. Linode also provides this info via
// unauthenticated API (https://api.linode.com/v4/regions) but the region names do not include the trailing digit "-1".
func GetRegionsLinode() ([]string, error) {
	// Akamai docs return a strange HTTP2 internal error if you don't request HTTP/2 with compression
	req, err := http.NewRequest(http.MethodGet, "https://techdocs.akamai.com/cloud-computing/docs/object-storage", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:128.0) Gecko/20100101 Firefox/128.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Connection", "keep-alive")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader) //nolint:gosec
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(buf.String()))
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
		"dreamhost":    GetRegionsDreamhost,
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
