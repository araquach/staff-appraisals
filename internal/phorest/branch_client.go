package phorest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"staff-appraisals/internal/models"
)

type BranchAPIResponse struct {
	Embedded struct {
		Branches []struct {
			BranchID     string   `json:"branchId"`
			Name         string   `json:"name"`
			TimeZone     string   `json:"timeZone"`
			Latitude     *float64 `json:"latitude"`
			Longitude    *float64 `json:"longitude"`
			Street1      string   `json:"streetAddress1"`
			Street2      string   `json:"streetAddress2"`
			City         string   `json:"city"`
			State        string   `json:"state"`
			PostalCode   string   `json:"postalCode"`
			Country      string   `json:"country"`
			CurrencyCode string   `json:"currencyCode"`
			AccountID    *int64   `json:"accountId"`
		} `json:"branches"`
	} `json:"_embedded"`
}

type BranchClient struct {
	BaseURL  string
	User     string
	Pass     string
	Business string
	HTTP     *http.Client
	Logger   *log.Logger
}

func NewBranchClient(user, pass, business string, lg *log.Logger) *BranchClient {
	return &BranchClient{
		BaseURL:  "https://api-gateway-eu.phorest.com/third-party-api-server/api",
		User:     user,
		Pass:     pass,
		Business: business,
		Logger:   lg,
		HTTP: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

func (c *BranchClient) FetchBranches() ([]models.Branch, error) {
	url := fmt.Sprintf("%s/business/%s/branch", c.BaseURL, c.Business)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.User, c.Pass)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("phorest branches: status %d: %s", resp.StatusCode, string(b))
	}

	var api BranchAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return nil, err
	}

	out := make([]models.Branch, 0, len(api.Embedded.Branches))
	for _, b := range api.Embedded.Branches {
		out = append(out, models.Branch{
			BranchID:     b.BranchID,
			Name:         b.Name,
			TimeZone:     b.TimeZone,
			Latitude:     b.Latitude,
			Longitude:    b.Longitude,
			Street1:      b.Street1,
			Street2:      b.Street2,
			City:         b.City,
			State:        b.State,
			PostalCode:   b.PostalCode,
			Country:      b.Country,
			CurrencyCode: b.CurrencyCode,
			AccountID:    b.AccountID,
		})
	}
	return out, nil
}
