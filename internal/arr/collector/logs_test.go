package collector

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/test_util"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestLogsClientCollect(t *testing.T) {
	var tests = []struct {
		name   string
		config *config.ArrConfig
		paths   []string
	}{
		{
			name: "radarr",
			config: &config.ArrConfig{
				App:        "radarr",
				ApiVersion: "v3",
			},
			paths: []string{"/api/v3/log?info", "/api/v3/log?warn", "/api/v3/log?error", "/api/v3/log?fatal"},
		},
		{
			name: "sonarr",
			config: &config.ArrConfig{
				App:        "sonarr",
				ApiVersion: "v3",
			},
			paths: []string{"/api/v3/log?info", "/api/v3/log?warn", "/api/v3/log?error", "/api/v3/log?fatal"},
		},
		{
			name: "lidarr",
			config: &config.ArrConfig{
				App:        "lidarr",
				ApiVersion: "v1",
			},
			paths: []string{"/api/v1/log?info", "/api/v1/log?warn", "/api/v1/log?error", "/api/v1/log?fatal"},
		},
		{
			name: "prowlarr",
			config: &config.ArrConfig{
				App:        "prowlarr",
				ApiVersion: "v1",
			},
			paths: []string{"/api/v1/log?info", "/api/v1/log?warn", "/api/v1/log?error", "/api/v1/log?fatal"},
		},
		{
			name: "readarr",
			config: &config.ArrConfig{
				App:        "readarr",
				ApiVersion: "v1",
			},
			paths: []string{"/api/v1/log?info", "/api/v1/log?warn", "/api/v1/log?error", "/api/v1/log?fatal"},
		},
		{
			name: "whisparr",
			config: &config.ArrConfig{
				App:        "whisparr",
				ApiVersion: "v3",
			},
			paths: []string{"/api/v3/log?info", "/api/v3/log?warn", "/api/v3/log?error", "/api/v3/log?fatal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			// loop each path
			for _, path := range tt.paths {
				ts, err := test_util.NewTestSharedServer(t, func(w http.ResponseWriter, r *http.Request) {
					// remove query params
					test_path := strings.Split(path, "?")[0]
					require.Contains(r.URL.Path, test_path)
				})
				require.NoError(err)

				defer ts.Close()

				tt.config.URL = ts.URL
				tt.config.ApiKey = test_util.API_KEY

				collector := NewLogsCollector(tt.config)

				expected_metrics_file := "expected_logs_metrics.txt"

				b, err := os.ReadFile(test_util.COMMON_FIXTURES_PATH + expected_metrics_file)
				require.NoError(err)

				expected := strings.Replace(string(b), "SOMEURL", ts.URL, -1)
				expected = strings.Replace(expected, "APP", tt.config.App, -1)

				f := strings.NewReader(expected)

				require.NotPanics(func() {
					err = testutil.CollectAndCompare(collector, f)
				})
				require.NoError(err)
			}
		})
	}
}

func TestLogsCollect_FailureDoesntPanic(t *testing.T) {
	require := require.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	config := &config.ArrConfig{
		URL:    ts.URL,
		ApiKey: test_util.API_KEY,
	}
	collector := NewLogsCollector(config)

	f := strings.NewReader("")

	require.NotPanics(func() {
		err := testutil.CollectAndCompare(collector, f)
		require.Error(err)
	}, "Collecting metrics should not panic on failure")
}
