package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

type narratarrTestTransport struct {
	handler http.Handler
}

func (t narratarrTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	t.handler.ServeHTTP(recorder, r)
	return recorder.Result(), nil
}

func TestNarratarrCollect(t *testing.T) {
	var mu sync.Mutex
	logins, refreshes := 0, 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/prefix/api/v1/status":
			fmt.Fprint(w, `{"app":"narratarr","version":"0.7.0","api_version":1}`)
		case "/prefix/api/v1/auth/login":
			require.Equal(t, "POST", r.Method)
			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, map[string]string{"username": "user", "password": "pass", "client": "Exportarr"}, body)
			logins++
			fmt.Fprint(w, `{"access_token":"access1","refresh_token":"refresh1","expires_in":3600}`)
		case "/prefix/api/v1/auth/refresh":
			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "refresh1", body["refresh_token"])
			refreshes++
			fmt.Fprint(w, `{"access_token":"access2","refresh_token":"refresh2","expires_in":3600}`)
		case "/prefix/api/v1/requests":
			expectedToken := "access1"
			if refreshes > 0 {
				expectedToken = "access2"
			}
			require.Equal(t, "Bearer "+expectedToken, r.Header.Get("Authorization"))
			require.Equal(t, "all", r.URL.Query().Get("format"))
			require.Equal(t, "100", r.URL.Query().Get("limit"))
			switch r.URL.Query().Get("offset") {
			case "0":
				fmt.Fprint(w, `{"items":[{"status":"wanted","format":"audiobook"},{"status":"available","format":"ebook"}],"total":3}`)
			case "2":
				fmt.Fprint(w, `{"items":[{"status":"failed","format":"audiobook"}],"total":3}`)
			default:
				t.Errorf("unexpected offset: %s", r.URL.Query().Get("offset"))
				http.Error(w, "bad offset", http.StatusBadRequest)
			}
		default:
			http.NotFound(w, r)
		}
	})
	serverURL := "http://narratarr.local"

	c := NewNarratarrCollector(&config.ArrConfig{
		URL: serverURL + "/prefix", ApiVersion: "v1", AuthUsername: "user", AuthPassword: "pass",
	})
	c.httpClient.Transport = narratarrTestTransport{handler}
	expected := fmt.Sprintf(`
		# HELP narratarr_info Narratarr version and API version
		# TYPE narratarr_info gauge
		narratarr_info{api_version="1",url="%s/prefix",version="0.7.0"} 1
		# HELP narratarr_requests Number of requests by status and format
		# TYPE narratarr_requests gauge
		narratarr_requests{format="audiobook",status="available",url="%s/prefix"} 0
		narratarr_requests{format="ebook",status="available",url="%s/prefix"} 1
		narratarr_requests{format="audiobook",status="downloading",url="%s/prefix"} 0
		narratarr_requests{format="ebook",status="downloading",url="%s/prefix"} 0
		narratarr_requests{format="audiobook",status="failed",url="%s/prefix"} 1
		narratarr_requests{format="ebook",status="failed",url="%s/prefix"} 0
		narratarr_requests{format="audiobook",status="wanted",url="%s/prefix"} 1
		narratarr_requests{format="ebook",status="wanted",url="%s/prefix"} 0
		# HELP narratarr_requests_total Total number of requests
		# TYPE narratarr_requests_total gauge
		narratarr_requests_total{url="%s/prefix"} 3
	`, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL)
	for i := 0; i < 2; i++ {
		require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected)))
	}
	require.Equal(t, 1, logins)
	require.Equal(t, 0, refreshes)
	c.expiresAt = time.Time{}
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected)))
	require.Equal(t, 1, logins)
	require.Equal(t, 1, refreshes)
}

func TestNarratarrExpiredTokenAndLegacyFormat(t *testing.T) {
	refreshes := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/status":
			fmt.Fprint(w, `{"app":"narratarr","version":"0.6.0","api_version":1}`)
		case "/api/v1/auth/login":
			fmt.Fprint(w, `{"access_token":"old","refresh_token":"refresh","expires_in":3600}`)
		case "/api/v1/auth/refresh":
			refreshes++
			fmt.Fprint(w, `{"access_token":"new","refresh_token":"next","expires_in":3600}`)
		case "/api/v1/requests":
			if r.Header.Get("Authorization") == "Bearer old" {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error":{"code":"token_expired"}}`)
			} else {
				fmt.Fprint(w, `{"items":[{"status":"available"}],"total":1}`)
			}
		default:
			http.NotFound(w, r)
		}
	})
	serverURL := "http://narratarr.local"

	c := NewNarratarrCollector(&config.ArrConfig{URL: serverURL, ApiVersion: "v1"})
	c.httpClient.Transport = narratarrTestTransport{handler}
	expected := fmt.Sprintf(`
		# HELP narratarr_requests Number of requests by status and format
		# TYPE narratarr_requests gauge
		narratarr_requests{format="audiobook",status="available",url="%s"} 1
		narratarr_requests{format="ebook",status="available",url="%s"} 0
		narratarr_requests{format="audiobook",status="downloading",url="%s"} 0
		narratarr_requests{format="ebook",status="downloading",url="%s"} 0
		narratarr_requests{format="audiobook",status="failed",url="%s"} 0
		narratarr_requests{format="ebook",status="failed",url="%s"} 0
		narratarr_requests{format="audiobook",status="wanted",url="%s"} 0
		narratarr_requests{format="ebook",status="wanted",url="%s"} 0
	`, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL, serverURL)
	require.NoError(t, testutil.CollectAndCompare(c, strings.NewReader(expected), "narratarr_requests"))
	require.Equal(t, 1, refreshes)
}

func TestNarratarrCollectFailsWithoutPartialMetrics(t *testing.T) {
	for _, response := range []string{
		`{"items":[],"total":2}`,
		`{"items":[{"status":"wanted"}]}`,
		`{"items":[{"format":"ebook"}],"total":1}`,
	} {
		t.Run(response, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/status":
					fmt.Fprint(w, `{"app":"narratarr","version":"0.7.0","api_version":1}`)
				case "/api/v1/auth/login":
					fmt.Fprint(w, `{"access_token":"access","refresh_token":"refresh","expires_in":3600}`)
				case "/api/v1/requests":
					fmt.Fprint(w, response)
				}

			})
			c := NewNarratarrCollector(&config.ArrConfig{URL: "http://narratarr.local", ApiVersion: "v1"})
			c.httpClient.Transport = narratarrTestTransport{handler}
			registry := prometheus.NewRegistry()
			registry.MustRegister(c)
			metrics, err := registry.Gather()
			require.Error(t, err)
			require.Empty(t, metrics)
		})
	}
}

func TestNarratarrHTTPFailure(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/status" {
			fmt.Fprint(w, `{"app":"narratarr","version":"0.7.0","api_version":1}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"code":"invalid_credentials"}}`)
	})
	c := NewNarratarrCollector(&config.ArrConfig{URL: "http://narratarr.local", ApiVersion: "v1"})
	c.httpClient.Transport = narratarrTestTransport{handler}
	registry := prometheus.NewRegistry()
	registry.MustRegister(c)
	metrics, err := registry.Gather()
	require.ErrorContains(t, err, "invalid_credentials")
	require.Empty(t, metrics)
}

func TestNarratarrConcurrentScrapesShareAuthentication(t *testing.T) {
	logins := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/status":
			fmt.Fprint(w, `{"app":"narratarr","version":"0.7.0","api_version":1}`)
		case "/api/v1/auth/login":
			logins++
			fmt.Fprint(w, `{"access_token":"access","refresh_token":"refresh","expires_in":3600}`)
		case "/api/v1/requests":
			fmt.Fprint(w, `{"items":[],"total":0}`)
		}
	})
	c := NewNarratarrCollector(&config.ArrConfig{URL: "http://narratarr.local", ApiVersion: "v1"})
	c.httpClient.Transport = narratarrTestTransport{handler}
	registry := prometheus.NewRegistry()
	registry.MustRegister(c)
	var wg sync.WaitGroup
	errors := make(chan error, 5)
	for i := 0; i < cap(errors); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := registry.Gather()
			errors <- err
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	require.Equal(t, 1, logins)
}
