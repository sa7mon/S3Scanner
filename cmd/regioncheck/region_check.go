package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strings"
)

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

func main() {
	doRegions, err := GetRegionsDO()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Printf("DigitalOcean: %v\n", doRegions)

	linodeRegions, lErr := GetRegionsLinode()
	if lErr != nil {
		log.Printf("Linode: %v\n", lErr)
	} else {
		log.Printf("Linode: %v", linodeRegions)
	}
}
