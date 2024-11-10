package collector

import (
	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type updateCollector struct {
	config       *config.ArrConfig // App configuration
	updateMetric *prometheus.Desc  // Total number of root folders
	errorMetric  *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewUpdateCollector(conf *config.ArrConfig) *updateCollector {
	return &updateCollector{
		config: conf,
		updateMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "", "update"),
			"Indicates if there is an update available",
			[]string{"version", "branch", "releaseDate", "hash"},
			prometheus.Labels{"url": conf.URL},
		),

		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "update", "collector_error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": conf.URL},
		),
	}
}

func (collector *updateCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.updateMetric
}

func (collector *updateCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "update")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	updates := model.Update{}
	if err := c.DoRequest("update", &updates); err != nil {
		log.Errorw("Error getting updates",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	if len(updates) > 0 {
		for _, update := range updates {
			status := 0
			if !update.Latest {
				// continue if not the latest version
				continue
			}

			if update.Latest {
				if !update.Installed {
					status += 1
				}
				ch <- prometheus.MustNewConstMetric(
					collector.updateMetric,
					prometheus.GaugeValue,
					float64(status),
					update.Version,
					update.Branch,
					update.ReleaseDate,
					update.Hash,
				)
			}
		}
	}
}
