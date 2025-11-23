package main

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

// GetRegionsLinode fetches region names from Linode docs HTML page. Linode also provides this info via
// unauthenticated API (https://api.linode.com/v4/regions) but the region names do not include the trailing digit "-1".
func GetRegionsLinode() ([]string, error) {
	// Linode's owner Akamai is very particular about how the HTTP request is made.
	// Using net/http proved too difficult so the burden is offloaded to github.com/imroc/req to craft a worthy request
	resp := req.
		ImpersonateFirefox().
		R().
		MustGet("https://techdocs.akamai.com/cloud-computing/docs/object-storage-product-limits")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}

	regionRe := regexp.MustCompile(`[\w-]+\.linodeobjects\.com`)

	regions := []string{}
	doc.Find(".rdmd-table:nth-of-type(1) tbody tr td:nth-of-type(4)").Each(func(_ int, t *goquery.Selection) {
		for _, r := range regionRe.FindAllString(t.Text(), -1) {
			regions = append(regions, strings.ReplaceAll(r, ".linodeobjects.com", ""))
		}
	})

	return regions, nil
}
