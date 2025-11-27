package phorest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	JobTypeTransactionsCSV = "TRANSACTIONS_CSV"
	JobTypeClientsCSV      = "CLIENT_CSV"
)

type ExportRequest struct {
	JobType          string `json:"jobType"`
	StartFilter      string `json:"startFilter,omitempty"`
	FinishFilter     string `json:"finishFilter,omitempty"`
	FilterExpression string `json:"filterExpression,omitempty"`
}

type ExportResponse struct {
	JobID              string  `json:"jobId"`
	JobType            string  `json:"jobType"`
	JobStatus          string  `json:"jobStatus"`
	Started            *string `json:"started,omitempty"`
	Finished           *string `json:"finished,omitempty"`
	StartFilter        string  `json:"startFilter"`
	FinishFilter       string  `json:"finishFilter"`
	FilterExpression   string  `json:"filterExpression,omitempty"`
	TotalRows          *int32  `json:"totalRows,omitempty"`
	SucceededRows      *int32  `json:"succeededRows,omitempty"`
	TempCSVExternalURL *string `json:"tempCsvExternalUrl,omitempty"`
	FailureReason      *string `json:"failureReason,omitempty"`
}

type ExportClient struct {
	http     *http.Client
	username string
	password string
	business string
}

func NewExportClient(username, password, business string) *ExportClient {
	return &ExportClient{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		username: username,
		password: password,
		business: business,
	}
}

func (c *ExportClient) CreateCSVExport(
	ctx context.Context,
	branchID string,
	jobType string,
	filterExpression string,
	startFilter string,
	finishFilter string,
) (*ExportResponse, error) {

	url := fmt.Sprintf(
		"https://api-gateway-eu.phorest.com/third-party-api-server/api/business/%s/branch/%s/csvexportjob",
		c.business, branchID,
	)

	reqPayload := ExportRequest{
		JobType:          jobType,
		FilterExpression: filterExpression,
		StartFilter:      startFilter,
		FinishFilter:     finishFilter,
	}

	body, _ := json.Marshal(reqPayload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create export job: %w", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("CSV export create failed: %s — %s", resp.Status, string(b))
	}

	var out ExportResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("decode create response: %w — body=%s", err, string(b))
	}

	return &out, nil
}

// Wait for a job to finish:
func (c *ExportClient) WaitForCSVJob(
	businessID string,
	branchID string,
	jobID string,
	maxWait time.Duration,
) (*ExportResponse, error) {

	url := fmt.Sprintf(
		"https://api-gateway-eu.phorest.com/third-party-api-server/api/business/%s/branch/%s/csvexportjob/%s",
		businessID, branchID, jobID,
	)

	deadline := time.Now().Add(maxWait)
	backoff := 2 * time.Second

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for job %s", jobID)
		}

		req, _ := http.NewRequest("GET", url, nil)
		req.SetBasicAuth(c.username, c.password)
		req.Header.Set("Accept", "application/json")

		res, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("poll failed: %w", err)
		}
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode > 299 {
			return nil, fmt.Errorf("poll non-2xx: %s — %s", res.Status, string(b))
		}

		var out ExportResponse
		if err := json.Unmarshal(b, &out); err != nil {
			return nil, fmt.Errorf("decode poll: %w — body=%s", err, string(b))
		}

		switch out.JobStatus {
		case "DONE":
			return &out, nil

		case "FAILED":
			// 👇 this is the important bit
			if out.FailureReason != nil && *out.FailureReason != "" {
				return &out, fmt.Errorf(
					"CSV job FAILED: %s — reason: %s",
					jobID, *out.FailureReason,
				)
			}
			return &out, fmt.Errorf("CSV job FAILED: %s", jobID)

		default:
			time.Sleep(backoff)
			if backoff < 10*time.Second {
				backoff += 2 * time.Second
			}
		}
	}
}

// Download the CSV (signed URL is usually enough)
func (c *ExportClient) DownloadCSV(url string, outPath string) error {

	// first try unsigned
	if err := c.tryDownload(url, outPath, false); err == nil {
		return nil
	}

	// fallback: BasicAuth
	return c.tryDownload(url, outPath, true)
}

func (c *ExportClient) tryDownload(url string, outPath string, withAuth bool) error {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if withAuth {
		req.SetBasicAuth(c.username, c.password)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("download non-2xx: %s — %s", res.Status, string(b))
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	return err
}
