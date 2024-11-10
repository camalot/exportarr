package collector

import (
	"strconv"

	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type downloadClientCollector struct {
	config        *config.ArrConfig // App configuration
	downloadclientMetric *prometheus.Desc  // Total number of root folders
	errorMetric   *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewDownloadClientCollector(conf *config.ArrConfig) *downloadClientCollector {
	return &downloadClientCollector{
		config: conf,
		downloadclientMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "", "downloadclient"),
			"Download Client Metrics.",
			[]string{"protocol", "name", "priority", "implementation", "removeCompletedDownloads", "removeFailedDownloads"},
			prometheus.Labels{"url": conf.URL},
		),

		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "indexer", "collector_error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": conf.URL},
		),
	}
}

func (collector *downloadClientCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.downloadclientMetric
}

func (collector *downloadClientCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "downloadclient")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	clients := model.DownloadClient{}
	if err := c.DoRequest("downloadclient", &clients); err != nil {
		log.Errorw("Error getting download clients",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	if len(clients) > 0 {
		for _, client := range clients {
			status := 0
			if client.Enable {
				status += 1
			}
			// []string{"protocol", "name", "priority", "implementation", "removeCompletedDownloads", "removeFailedDownloads"},

			ch <- prometheus.MustNewConstMetric(
				collector.downloadclientMetric,
				prometheus.GaugeValue,
				float64(status),
				client.Protocol,
				client.Name,
				strconv.Itoa(client.Priority),
				client.Implementation,
				strconv.FormatBool(client.RemoveCompletedDownloads),
				strconv.FormatBool(client.RemoveFailedDownloads),
			)
		}
	}
}
