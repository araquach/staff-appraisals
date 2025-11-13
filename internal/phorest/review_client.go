package phorest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"staff-appraisals/internal/models"
)

type reviewAPIResponse struct {
	Embedded struct {
		Reviews []struct {
			ReviewID        string `json:"reviewId"`
			ClientID        string `json:"clientId"`
			ClientFirstName string `json:"clientFirstName"`
			ClientLastName  string `json:"clientLastName"`
			ReviewDate      string `json:"reviewDate"` // "YYYY-MM-DD"
			VisitDate       string `json:"visitDate"`  // "YYYY-MM-DD"
			StaffID         string `json:"staffId"`
			StaffFirstName  string `json:"staffFirstName"`
			StaffLastName   string `json:"staffLastName"`
			Text            string `json:"text"`
			Rating          int    `json:"rating"`
			FacebookReview  bool   `json:"facebookReview"`
			TwitterReview   bool   `json:"twitterReview"`
		} `json:"reviews"`
	} `json:"_embedded"`
	Page struct {
		Size          int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages    int `json:"totalPages"`
		Number        int `json:"number"`
	} `json:"page"`
}

type ReviewsClient struct {
	BaseURL  string
	User     string
	Pass     string
	Business string
	HTTP     *http.Client
}

func NewReviewsClient(user, pass, business string) *ReviewsClient {
	return &ReviewsClient{
		BaseURL:  "https://api-gateway-eu.phorest.com/third-party-api-server/api",
		User:     user,
		Pass:     pass,
		Business: business,
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

// FetchReviews fetches reviews for a branch.
// If sinceDate is non-empty (YYYY-MM-DD), it will use it as a lower bound on reviewDate (if supported).
func (c *ReviewsClient) FetchReviews(branchID, sinceDate string, page, size int) ([]models.Review, int, error) {
	// Paging params (Phorest list endpoints are usually size/page based)
	if size <= 0 {
		size = 200
	}
	url := fmt.Sprintf("%s/business/%s/branch/%s/review?size=%d&page=%d",
		c.BaseURL, c.Business, branchID, size, page)

	// (If the endpoint supports filtering by reviewDate, you can append `&reviewDateStart=%s`)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.SetBasicAuth(c.User, c.Pass)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("phorest reviews %s: status %d: %s", branchID, resp.StatusCode, string(b))
	}

	var api reviewAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return nil, 0, err
	}

	out := make([]models.Review, 0, len(api.Embedded.Reviews))
	parse := func(s string) *time.Time {
		if s == "" {
			return nil
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil
		}
		return &t
	}

	for _, r := range api.Embedded.Reviews {
		out = append(out, models.Review{
			ReviewID:        r.ReviewID,
			BranchID:        branchID,
			ClientID:        r.ClientID,
			ClientFirstName: r.ClientFirstName,
			ClientLastName:  r.ClientLastName,
			ReviewDate:      parse(r.ReviewDate),
			VisitDate:       parse(r.VisitDate),
			StaffID:         r.StaffID,
			StaffFirstName:  r.StaffFirstName,
			StaffLastName:   r.StaffLastName,
			Text:            r.Text,
			Rating:          r.Rating,
			FacebookReview:  r.FacebookReview,
			TwitterReview:   r.TwitterReview,
		})
	}
	return out, api.Page.TotalPages, nil
}

func (c *ReviewsClient) FetchLatestN(branchID string, n int) ([]models.Review, error) {
	if n <= 0 {
		n = 10
	}
	// page=0 with size=n
	rows, _, err := c.FetchReviews(branchID, "", 0, n)
	return rows, err
}
