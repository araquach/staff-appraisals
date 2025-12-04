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

	"github.com/araquach/phorest-datahub/internal/models"
)

type StaffAPIResponse struct {
	Embedded struct {
		Staffs []struct {
			StaffID                   string  `json:"staffId"`
			StaffCategoryID           string  `json:"staffCategoryId"`
			UserID                    string  `json:"userId"`
			StaffCategoryName         string  `json:"staffCategoryName"`
			FirstName                 string  `json:"firstName"`
			LastName                  string  `json:"lastName"`
			BirthDate                 *string `json:"birthDate"`
			StartDate                 *string `json:"startDate"`
			SelfEmployed              bool    `json:"selfEmployed"`
			Archived                  bool    `json:"archived"`
			Mobile                    string  `json:"mobile"`
			Email                     string  `json:"email"`
			Gender                    string  `json:"gender"`
			Notes                     string  `json:"notes"`
			OnlineProfile             string  `json:"onlineProfile"`
			HideFromOnlineBookings    bool    `json:"hideFromOnlineBookings"`
			HideFromAppointmentScreen bool    `json:"hideFromAppointmentScreen"`
			ImageURL                  string  `json:"imageUrl"`
		} `json:"staffs"`
	} `json:"_embedded"`
}

type StaffClient struct {
	BaseURL  string
	User     string
	Pass     string
	Business string
	HTTP     *http.Client
	Logger   *log.Logger
}

func NewStaffClient(user, pass, business string, lg *log.Logger) *StaffClient {
	return &StaffClient{
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

func (c *StaffClient) FetchStaff(branchID string) ([]models.Staff, error) {
	url := fmt.Sprintf("%s/business/%s/branch/%s/staff?fetch_archived=true&size=%d",
		c.BaseURL, c.Business, branchID, 200)

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
		return nil, fmt.Errorf("phorest staff %s: %d: %s", branchID, resp.StatusCode, string(b))
	}

	var api StaffAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return nil, err
	}

	out := make([]models.Staff, 0, len(api.Embedded.Staffs))
	for _, s := range api.Embedded.Staffs {
		out = append(out, models.Staff{
			StaffID:  s.StaffID,
			BranchID: branchID,

			StaffCategoryID:           s.StaffCategoryID,
			UserID:                    s.UserID,
			StaffCategoryName:         s.StaffCategoryName,
			FirstName:                 s.FirstName,
			LastName:                  s.LastName,
			BirthDate:                 parseDatePtr(s.BirthDate),
			StartDate:                 parseDatePtr(s.StartDate),
			SelfEmployed:              s.SelfEmployed,
			Archived:                  s.Archived,
			Mobile:                    s.Mobile,
			Email:                     s.Email,
			Gender:                    s.Gender,
			Notes:                     s.Notes,
			OnlineProfile:             s.OnlineProfile,
			HideFromOnlineBookings:    s.HideFromOnlineBookings,
			HideFromAppointmentScreen: s.HideFromAppointmentScreen,
			ImageURL:                  s.ImageURL,
		})
	}
	return out, nil
}

func parseDatePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, *s); err == nil {
			d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			return &d
		}
	}
	return nil
}
