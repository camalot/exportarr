package collector

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type narratarrCollector struct {
	config       *config.ArrConfig
	httpClient   *http.Client
	mu           sync.Mutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	requests     *prometheus.Desc
	total        *prometheus.Desc
	status       *prometheus.Desc
	errorMetric  *prometheus.Desc
}

func NewNarratarrCollector(c *config.ArrConfig) *narratarrCollector {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if c.DisableSSLVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	labels := prometheus.Labels{"url": c.URL}
	return &narratarrCollector{
		config: c,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		requests:    prometheus.NewDesc("narratarr_requests", "Number of requests by status and format", []string{"status", "format"}, labels),
		total:       prometheus.NewDesc("narratarr_requests_total", "Total number of requests", nil, labels),
		status:      prometheus.NewDesc("narratarr_info", "Narratarr version and API version", []string{"version", "api_version"}, labels),
		errorMetric: prometheus.NewDesc("narratarr_collector_error", "Error while collecting metrics", nil, labels),
	}
}

func (c *narratarrCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.requests
	ch <- c.total
	ch <- c.status
}

type narratarrTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type narratarrAPIError struct {
	Status int
	Code   string
}

func (e *narratarrAPIError) Error() string {
	return fmt.Sprintf("Narratarr API returned HTTP %d (%s)", e.Status, e.Code)
}

func (c *narratarrCollector) request(method, endpoint string, query url.Values, body interface{}, authenticated bool, target interface{}) error {
	u, err := url.Parse(c.config.BaseURL())
	if err != nil {
		return fmt.Errorf("invalid Narratarr URL: %w", err)
	}
	u = u.JoinPath(c.config.ApiRootPath, endpoint)
	u.RawQuery = query.Encode()
	var data []byte
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authenticated {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Narratarr request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var result struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&result)
		return &narratarrAPIError{Status: resp.StatusCode, Code: result.Error.Code}
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("invalid Narratarr response: %w", err)
	}
	return nil
}

func (c *narratarrCollector) authenticate() error {
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return nil
	}
	endpoint := "auth/login"
	body := map[string]string{
		"username": c.config.AuthUsername,
		"password": c.config.AuthPassword,
		"client":   "Exportarr",
	}
	if c.refreshToken != "" {
		endpoint = "auth/refresh"
		body = map[string]string{"refresh_token": c.refreshToken}
	}
	var tokens narratarrTokens
	if err := c.request(http.MethodPost, endpoint, nil, body, false, &tokens); err != nil {
		return fmt.Errorf("Narratarr authentication failed: %w", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" || tokens.ExpiresIn <= 0 {
		return fmt.Errorf("Narratarr authentication returned invalid tokens")
	}
	c.accessToken, c.refreshToken = tokens.AccessToken, tokens.RefreshToken
	c.expiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	return nil
}

func (c *narratarrCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.collect(ch); err != nil {
		zap.S().Errorw("Error collecting Narratarr metrics", "error", err)
		ch <- prometheus.NewInvalidMetric(c.errorMetric, err)
	}
}

func (c *narratarrCollector) collect(ch chan<- prometheus.Metric) error {
	var status struct {
		App        string `json:"app"`
		Version    string `json:"version"`
		APIVersion int    `json:"api_version"`
	}
	if err := c.request(http.MethodGet, "status", nil, nil, false, &status); err != nil {
		return err
	}
	if status.App != "narratarr" || status.Version == "" || status.APIVersion != 1 {
		return fmt.Errorf("invalid Narratarr status response")
	}
	if err := c.authenticate(); err != nil {
		return err
	}

	type key struct{ status, format string }
	counts := make(map[key]int)
	for _, status := range []string{"wanted", "downloading", "available", "failed"} {
		for _, format := range []string{"audiobook", "ebook"} {
			counts[key{status, format}] = 0
		}
	}
	total := 0
	for offset := 0; ; {
		var page struct {
			Items []struct {
				Status string `json:"status"`
				Format string `json:"format"`
			} `json:"items"`
			Total *int `json:"total"`
		}
		query := url.Values{"format": {"all"}, "limit": {"100"}, "offset": {strconv.Itoa(offset)}}
		err := c.request(http.MethodGet, "requests", query, nil, true, &page)
		if apiErr, ok := err.(*narratarrAPIError); ok && apiErr.Status == http.StatusUnauthorized && apiErr.Code == "token_expired" {
			c.expiresAt = time.Time{}
			if err = c.authenticate(); err != nil {
				return err
			}
			err = c.request(http.MethodGet, "requests", query, nil, true, &page)
		}
		if err != nil {
			return err
		}
		if page.Total == nil || *page.Total < 0 || page.Items == nil {
			return fmt.Errorf("invalid Narratarr requests response")
		}
		for _, item := range page.Items {
			if item.Status == "" {
				return fmt.Errorf("Narratarr request has no status")
			}
			if item.Format == "" {
				item.Format = "audiobook" // Before Narratarr 0.7.0 all requests were audiobooks.
			}
			counts[key{item.Status, item.Format}]++
		}
		total += len(page.Items)
		offset += len(page.Items)
		if offset >= *page.Total {
			break
		}
		if len(page.Items) == 0 {
			return fmt.Errorf("Narratarr returned an incomplete requests page")
		}
	}
	ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, 1, status.Version, strconv.Itoa(status.APIVersion))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(total))
	for k, count := range counts {
		ch <- prometheus.MustNewConstMetric(c.requests, prometheus.GaugeValue, float64(count), k.status, k.format)
	}
	return nil
}
