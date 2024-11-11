package collector

import (
	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type rootFolderCollector struct {
	config           *config.ArrConfig // App configuration
	freeSpaceMetric  *prometheus.Desc  // freespace of root folder
	totalSpaceMetric  *prometheus.Desc  // total space of root folder
	errorMetric      *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewRootFolderCollector(c *config.ArrConfig) *rootFolderCollector {
	return &rootFolderCollector{
		config: c,
		freeSpaceMetric: prometheus.NewDesc(
			prometheus.BuildFQName(c.App, "rootfolder", "freespace_bytes"),
			"Root folder free space in bytes by path",
			[]string{"path"},
			prometheus.Labels{"url": c.URL},
		),
		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(c.App, "rootfolder", "collector_error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": c.URL},
		),
	}
}

func (collector *rootFolderCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.freeSpaceMetric
}

func (collector *rootFolderCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "rootfolder")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	rootFolders := model.RootFolder{}
	if err := c.DoRequest("rootfolder", &rootFolders); err != nil {
		log.Errorw("Error getting rootfolder",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	// Group metrics by path
	if len(rootFolders) > 0 {
		for _, rootFolder := range rootFolders {
			ch <- prometheus.MustNewConstMetric(collector.freeSpaceMetric, prometheus.GaugeValue, float64(rootFolder.FreeSpace),
				rootFolder.Path,
			)
		}
	}
}
