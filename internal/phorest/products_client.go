package phorest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ProductsClient struct {
	BaseURL    string
	BusinessID string
	Username   string
	Password   string
	HTTPClient *http.Client
}

func NewProductsClient(baseURL, businessID, username, password string) *ProductsClient {
	if baseURL == "" {
		baseURL = "https://api-gateway-eu.phorest.com/third-party-api-server"
	}
	return &ProductsClient{
		BaseURL:    baseURL,
		BusinessID: businessID,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type PhorestProduct struct {
	ProductID           string    `json:"productId"`
	ParentProductID     string    `json:"parentProductId"`
	Name                string    `json:"name"`
	BrandID             string    `json:"brandId"`
	BrandName           string    `json:"brandName"`
	CategoryID          string    `json:"categoryId"`
	CategoryName        string    `json:"categoryName"`
	Archived            bool      `json:"archived"`
	Price               float64   `json:"price"`
	MinQuantity         float64   `json:"minQuantity"`
	MaxQuantity         float64   `json:"maxQuantity"`
	Type                string    `json:"type"`
	Barcode             string    `json:"barcode"`
	MeasurementQuantity float64   `json:"measurementQuantity"`
	MeasurementUnit     string    `json:"measurementUnit"`
	ReorderCount        float64   `json:"reorderCount"`
	ReorderCost         float64   `json:"reorderCost"`
	QuantityInStock     float64   `json:"quantityInStock"`
	Code                string    `json:"code"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type listProductsResponse struct {
	Embedded struct {
		Products []PhorestProduct `json:"products"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

type ListProductsOptions struct {
	BranchID      string
	ProductType   string
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	Page          int
	Size          int
}

func (c *ProductsClient) ListProducts(ctx context.Context, opts ListProductsOptions) (*listProductsResponse, error) {
	if opts.Size <= 0 {
		opts.Size = 100
	}
	u, err := url.Parse(fmt.Sprintf("%s/api/business/%s/branch/%s/product", c.BaseURL, c.BusinessID, opts.BranchID))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("size", fmt.Sprint(opts.Size))
	q.Set("page", fmt.Sprint(opts.Page))
	if opts.ProductType != "" {
		q.Set("productType", opts.ProductType)
	}

	// Only send date filters if we have BOTH after & before,
	// because the API insists on both.
	if opts.UpdatedAfter != nil && opts.UpdatedBefore != nil {
		after := opts.UpdatedAfter.UTC()
		before := opts.UpdatedBefore.UTC()

		// Use a precise format similar to their JSON timestamps, e.g. "2025-11-29T05:44:49.265Z"
		layout := "2006-01-02T15:04:05.000Z"

		q.Set("updatedAfter", after.Format(layout))
		q.Set("updatedBefore", before.Format(layout))
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = "<empty body>"
		}
		return nil, fmt.Errorf("phorest: list products %s returned %d: %s", u.String(), resp.StatusCode, msg)
	}

	var out listProductsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
