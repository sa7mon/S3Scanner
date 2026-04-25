package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ProductsResponse struct {
	Products []Product `json:"products"`
}

type Product struct {
	ProductCategory string `json:"product_category"`
	ProductName     string `json:"product"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	Locality        struct {
		Region string `json:"region"`
	} `json:"locality"`
}

/*
	    GetRegionsScaleway gets regions from the Scaleway API
		Documentation: https://www.scaleway.com/en/developers/api/product-catalog/public-catalog-api/
*/
func GetRegionsScaleway() ([]string, error) {
	requestURL := "https://api.scaleway.com/product-catalog/v2alpha1/public-catalog/products?product_types=object_storage"
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
	resp := ProductsResponse{}
	unmarshalErr := json.Unmarshal(bytes, &resp)
	if unmarshalErr != nil {
		return nil, err
	}

	regions := []string{}
	for _, product := range resp.Products {
		if product.ProductCategory == "Object Storage" && product.ProductName == "Standard One Zone" {
			regions = append(regions, product.Locality.Region)
		}
	}

	return regions, nil
}
