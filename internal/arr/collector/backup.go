package collector

import (
	"strconv"
	"time"

	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type backupCollector struct {
	config           *config.ArrConfig // App configuration
	backupsMetric    *prometheus.Desc  // 
	lastBackupMetric *prometheus.Desc  // 
	errorMetric      *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewBackupCollector(conf *config.ArrConfig) *backupCollector {
	return &backupCollector{
		config: conf,
		backupsMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "system", "backup_bytes"),
			"Backup metrics for the application.",
			[]string{"name", "type", "time", "age"},
			prometheus.Labels{"url": conf.URL},
		),

		lastBackupMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "system", "backup"),
			"The time in seconds since the last backup.",
			[]string{"name", "type"},
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
	ch <- collector.backupsMetric
	ch <- collector.lastBackupMetric
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
		// store the last backup time
		type lastBackup struct {
			age      int64
			name     string
			typeName string
		}
		var lastBackupInfo lastBackup = lastBackup{age: -1, name: "", typeName: ""}
		const layout = "2024-11-05T16:48:59Z"
		for _, backup := range backups {
			// convert the date time string to a unix timestamp
			backupTime, err := time.Parse(layout, backup.Time)
			if err != nil {
				log.Errorw("Error parsing backup time",
					"error", err)
				ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
			}
			backupUnix := backupTime.Unix()
			unixNow := time.Now().Unix()
			age := unixNow - backupUnix
			if age < lastBackupInfo.age || lastBackupInfo.age == -1 {
				lastBackupInfo.age = age
				lastBackupInfo.name = backup.Name
				lastBackupInfo.typeName = backup.Type
			}

			ch <- prometheus.MustNewConstMetric(
				collector.backupsMetric,
				prometheus.GaugeValue,
				float64(backup.Size),
				backup.Name,
				backup.Type,
				backup.Time,
				strconv.FormatInt(age, 10),
			)
		}

		if lastBackupInfo.age != -1 {
			// store the last backup time
			ch <- prometheus.MustNewConstMetric(
				collector.lastBackupMetric,
				prometheus.GaugeValue,
				float64(lastBackupInfo.age),
				lastBackupInfo.name,
				lastBackupInfo.typeName,
			)
		}
	}
}
