package main

import (
	"github.com/sa7mon/s3scanner/provider"
	"log"
	"os"
	"sort"
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

func main() {
	results := map[string][]string{}
	errors := map[string]error{}

	p := map[string]func() ([]string, error){
		"digitalocean": GetRegionsDO,
		"dreamhost":    GetRegionsDreamhost,
		"linode":       GetRegionsLinode,
		"scaleway":     GetRegionsScaleway,
		"wasabi":       getRegionsWasabi,
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
