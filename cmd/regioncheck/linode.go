package main

import (
	"compress/gzip"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// GetRegionsLinode fetches region names from Linode docs HTML page. Linode also provides this info via
// unauthenticated API (https://api.linode.com/v4/regions) but the region names do not include the trailing digit "-1".
func GetRegionsLinode() ([]string, error) {
	// Akamai docs return a strange HTTP2 internal error if you don't request HTTP/2 with compression
	req, err := http.NewRequest(http.MethodGet, "https://techdocs.akamai.com/cloud-computing/docs/object-storage-product-limits", nil)
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

	regionRe := regexp.MustCompile(`[\w-]+\.linodeobjects\.com`)

	regions := []string{}
	doc.Find(".rdmd-table:nth-of-type(1) tbody tr td:nth-of-type(4)").Each(func(_ int, t *goquery.Selection) {
		for _, r := range regionRe.FindAllString(t.Text(), -1) {
			regions = append(regions, strings.Replace(r, ".linodeobjects.com", "", -1))
		}
	})

	return regions, nil
}
