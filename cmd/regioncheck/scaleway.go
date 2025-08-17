package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

func GetRegionsScaleway() ([]string, error) {
	var re = regexp.MustCompile(`Region: \x60(.+)\x60`)
	requestURL := "https://raw.githubusercontent.com/scaleway/docs-content/refs/heads/main/pages/object-storage/how-to/create-a-bucket.mdx"
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
