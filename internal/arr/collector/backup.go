package collector

import (
	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type backupCollector struct {
	config        *config.ArrConfig // App configuration
	backupMetric *prometheus.Desc  // Total number of root folders
	errorMetric   *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewBackupCollector(conf *config.ArrConfig) *backupCollector {
	return &backupCollector{
		config: conf,
		backupMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "system", "backup_bytes"),
			"Backup metrics for the application.",
			[]string{"name", "type", "time"},
			prometheus.Labels{"url": conf.URL},
		),

		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "system_backup", "collector_error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": conf.URL},
		),
	}
}

func (collector *backupCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.backupMetric
}

func (collector *backupCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "backup")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	backups := []model.Backup{}
	if err := c.DoRequest("system/backup", &backups); err != nil {
		log.Errorw("Error getting backups",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	if len(backups) > 0 {
		for _, backup := range backups {
			ch <- prometheus.MustNewConstMetric(
				collector.backupMetric,
				prometheus.GaugeValue,
				float64(backup.Size),
				backup.Name,
				backup.Type,
				backup.Time,
			)
		}
	}
}
