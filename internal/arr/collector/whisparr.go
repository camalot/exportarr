package collector

import (
	"fmt"
	"time"

	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	whisparr_namespace = "whisparr"
)

type whisparrCollector struct {
	config                  *config.ArrConfig // App configuration
	seriesMetric            *prometheus.Desc  // Total number of series
	seriesDownloadedMetric  *prometheus.Desc  // Total number of downloaded series
	seriesMonitoredMetric   *prometheus.Desc  // Total number of monitored series
	seriesUnmonitoredMetric *prometheus.Desc  // Total number of unmonitored series
	seriesFileSizeMetric    *prometheus.Desc  // Total fizesize of all series in bytes
	errorMetric             *prometheus.Desc  // Error Description for use with InvalidMetric
	diskSpaceMetric         *prometheus.Desc  // Total disk space
	blocklistMetric         *prometheus.Desc  // Total number of blocklisted items
	backupsMetric           *prometheus.Desc  // Total number of backups
	indexersMetric          *prometheus.Desc  // indexers available
	// tagsMetric              *prometheus.Desc  // tags available
}

func NewWhisparrCollector(c *config.ArrConfig) *whisparrCollector {
	return &whisparrCollector{
		config: c,
		seriesMetric: prometheus.NewDesc(
			prometheus.BuildFQName(whisparr_namespace, "series", "total"),
			"Total number of series",
			nil,
			prometheus.Labels{"url": c.URL},
		),
		seriesDownloadedMetric: prometheus.NewDesc(
			"sonarr_series_downloaded_total",
			"Total number of downloaded series",
			nil,
			prometheus.Labels{"url": c.URL},
		),
		seriesMonitoredMetric: prometheus.NewDesc(
			"sonarr_series_monitored_total",
			"Total number of monitored series",
			nil,
			prometheus.Labels{"url": c.URL},
		),
		seriesUnmonitoredMetric: prometheus.NewDesc(
			"sonarr_series_unmonitored_total",
			"Total number of unmonitored series",
			nil,
			prometheus.Labels{"url": c.URL},
		),
		seriesFileSizeMetric: prometheus.NewDesc(
			"sonarr_series_filesize_bytes",
			"Total fizesize of all series in bytes",
			nil,
			prometheus.Labels{"url": c.URL},
		),
		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(whisparr_namespace, "collector", "error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": c.URL},
		),
	}
}

func (collector *whisparrCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.seriesMetric
	ch <- collector.seriesDownloadedMetric
	ch <- collector.seriesMonitoredMetric
	ch <- collector.seriesUnmonitoredMetric
	ch <- collector.seriesFileSizeMetric
	ch <- collector.errorMetric
	ch <- collector.diskSpaceMetric
	ch <- collector.blocklistMetric
	ch <- collector.backupsMetric
	ch <- collector.indexersMetric
}

func (collector *whisparrCollector) Collect(ch chan<- prometheus.Metric) {
	// total := time.Now()
	log := zap.S().With("collector", "whisparr")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	var seriesFileSize int64
	var (
		seriesDownloaded    = 0
		seriesMonitored     = 0
		seriesUnmonitored   = 0
		seasons             = 0
		seasonsDownloaded   = 0
		seasonsMonitored    = 0
		seasonsUnmonitored  = 0

		episodes            = 0
		episodesDownloaded  = 0
		episodesMonitored   = 0
		episodesUnmonitored = 0
		episodesQualities   = map[string]int{}

	)

	cseries := []time.Duration{}
	series := model.Series{}
	if err := c.DoRequest("series", &series); err != nil {
		log.Errorw("Error getting series",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	for _, s := range series {
		tseries := time.Now()

		if s.Monitored {
			seriesMonitored++
		} else {
			seriesUnmonitored++
		}

		if s.Statistics.PercentOfEpisodes == 100 {
			seriesDownloaded++
		}

		seasons += s.Statistics.SeasonCount
		episodes += s.Statistics.TotalEpisodeCount
		episodesDownloaded += s.Statistics.EpisodeFileCount
		seriesFileSize += s.Statistics.SizeOnDisk

		for _, e := range s.Seasons {
			if e.Monitored {
				seasonsMonitored++
			} else {
				seasonsUnmonitored++
			}

			if e.Statistics.PercentOfEpisodes == 100 {
				seasonsDownloaded++
			}
		}

		if collector.config.EnableAdditionalMetrics {
			textra := time.Now()
			episodeFile := model.EpisodeFile{}
			params := map[string]string{"seriesId": fmt.Sprintf("%d", s.Id)}
			if err := c.DoRequest("episodefile", &episodeFile, params); err != nil {
				log.Errorw("Error getting episodefile",
					"error", err)
				ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
				return
			}
			for _, e := range episodeFile {
				if e.Quality.Quality.Name != "" {
					episodesQualities[e.Quality.Quality.Name]++
				}
			}

			episode := model.Episode{}
			if err := c.DoRequest("episode", &episode, params); err != nil {
				log.Errorw("Error getting episode",
					"error", err)
				ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
				return
			}
			for _, e := range episode {
				if e.Monitored {
					episodesMonitored++
				} else {
					episodesUnmonitored++
				}
			}
			log.Debugw("Extra options completed",
				"duration", time.Since(textra))
		}
		e := time.Since(tseries)
		cseries = append(cseries, e)
		log.Debugw("series completed",
			"series_id", s.Id,
			"duration", e)
	}

	episodesMissing := model.Missing{}
	params := map[string]string{"sortKey": "airDateUtc"}
	if err := c.DoRequest("wanted/missing", &episodesMissing, params); err != nil {
		log.Errorw("Error getting missing",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
}