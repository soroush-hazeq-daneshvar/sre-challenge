package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sre-challenge/geoip-api/internal/metrics"
	"github.com/sre-challenge/geoip-api/internal/repository"
)

type Provider interface {
	Lookup(ctx context.Context, ip string) (*repository.GeoIPRecord, error)
}

type IPAPIProvider struct {
	baseURL string
	client  *http.Client
}

func NewIPAPIProvider(baseURL string, timeout time.Duration) *IPAPIProvider {
	return &IPAPIProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type ipAPIResponse struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country_name"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Error       bool    `json:"error"`
	Reason      string  `json:"reason"`
}

func (p *IPAPIProvider) Lookup(ctx context.Context, ip string) (*repository.GeoIPRecord, error) {
	start := time.Now()
	url := fmt.Sprintf("%s/%s/json/", p.baseURL, ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		metrics.ExternalAPICallsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "sre-challenge-geoip-api/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		metrics.ExternalAPICallsTotal.WithLabelValues("error").Inc()
		metrics.ExternalAPIDuration.Observe(time.Since(start).Seconds())
		return nil, fmt.Errorf("external api call: %w", err)
	}
	defer resp.Body.Close()

	metrics.ExternalAPIDuration.Observe(time.Since(start).Seconds())

	if resp.StatusCode != http.StatusOK {
		metrics.ExternalAPICallsTotal.WithLabelValues("error").Inc()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("external api returned %d: %s", resp.StatusCode, string(body))
	}

	var apiResp ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		metrics.ExternalAPICallsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if apiResp.Error {
		metrics.ExternalAPICallsTotal.WithLabelValues("error").Inc()
		return nil, fmt.Errorf("external api error: %s", apiResp.Reason)
	}

	metrics.ExternalAPICallsTotal.WithLabelValues("success").Inc()

	return &repository.GeoIPRecord{
		IP:          ip,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		City:        apiResp.City,
		Region:      apiResp.Region,
		Latitude:    apiResp.Latitude,
		Longitude:   apiResp.Longitude,
		Provider:    "ipapi",
	}, nil
}
