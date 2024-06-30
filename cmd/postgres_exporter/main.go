// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"reflect"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/postgres_exporter/collector"
	"github.com/prometheus-community/postgres_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/prometheus-community/postgres_exporter/myencrypt"
)

var (
	c = config.Handler{
		Config: &config.Config{},
	}

	configFile             = kingpin.Flag("config.file", "gauss exporter configuration file.").Default("gauss_exporter.yml").String()
	webConfig              = kingpinflag.AddFlags(kingpin.CommandLine, ":9103")
	metricsPath            = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	disableDefaultMetrics  = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	disableSettingsMetrics = kingpin.Flag("disable-settings-metrics", "Do not include dm_settings metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_SETTINGS_METRICS").Bool()
	autoDiscoverDatabases  = kingpin.Flag("auto-discover-databases", "Whether to discover the databases on a server dynamically. (DEPRECATED)").Default("false").Envar("PG_EXPORTER_AUTO_DISCOVER_DATABASES").Bool()
	queriesPath            = kingpin.Flag("extend.query-path", "Path to custom queries to run. (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXTEND_QUERY_PATH").String()
	onlyDumpMaps           = kingpin.Flag("dumpmaps", "Do not run, simply dump the maps.").Bool()
	constantLabelsList     = kingpin.Flag("constantLabels", "A list of label=value separated by comma(,). (DEPRECATED)").Default("").Envar("PG_EXPORTER_CONSTANT_LABELS").String()
	excludeDatabases       = kingpin.Flag("exclude-databases", "A list of databases to remove when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_EXCLUDE_DATABASES").String()
	includeDatabases       = kingpin.Flag("include-databases", "A list of databases to include when autoDiscoverDatabases is enabled (DEPRECATED)").Default("").Envar("PG_EXPORTER_INCLUDE_DATABASES").String()
	metricPrefix           = kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"dm\") prefixes for each of the metrics").Default("dm").Envar("PG_EXPORTER_METRIC_PREFIX").String()
	logger                 = log.NewNopLogger()
)

// Metric name parts.
const (
	// Namespace for all metrics.
	namespace = "gs"
	// Subsystems.
	exporter = "exporter"
	// The name of the exporter.
	exporterName = "gauss_exporter"
	// Metric label used for static string data thats handy to send to Prometheus
	// e.g. version
	staticLabelName = "static"
	// Metric label used for server identification.
	serverLabelName = "server"
)

func main() {
	kingpin.Version(version.Print(exporterName))
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger = promlog.New(promlogConfig)

	if *onlyDumpMaps {
		dumpMaps()
		return
	}

	
	level.Info(logger).Log("msg",reflect.TypeOf(webConfig))

	excludedDatabases := strings.Split(*excludeDatabases, ",")
	level.Info(logger).Log("msg", "Excluded databases", "databases", fmt.Sprintf("%v", excludedDatabases))

	if *queriesPath != "" {
		level.Warn(logger).Log("msg", "The extended queries.yaml config is DEPRECATED", "file", *queriesPath)
	}

	if *autoDiscoverDatabases || *excludeDatabases != "" || *includeDatabases != "" {
		level.Warn(logger).Log("msg", "Scraping additional databases via auto discovery is DEPRECATED")
	}

	if *constantLabelsList != "" {
		level.Warn(logger).Log("msg", "Constant labels on all metrics is DEPRECATED")
	}

	opts := []ExporterOpt{
		DisableDefaultMetrics(*disableDefaultMetrics),
		DisableSettingsMetrics(*disableSettingsMetrics),
		AutoDiscoverDatabases(*autoDiscoverDatabases),
		WithUserQueriesPath(*queriesPath),
		WithConstantLabels(*constantLabelsList),
		ExcludeDatabases(excludedDatabases),
		IncludeDatabases(*includeDatabases),
	}


	if fi,err := os.Stat("HISTORY");err == nil || os.IsExist(err) {
		level.Warn(logger).Log("监测到HISTORY文件，执行新模式")

		if fi.Size() < 1 {
			level.Warn(logger).Log("HISTORY文件是空的，部署有问题，无法继续")
			os.Exit(1)
		}

		tif := myencrypt.CheckInstall((*webConfig.WebListenAddresses)[0])

		fmt.Println(tif)

		dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", tif.Scheme, tif.Username, tif.Password, tif.Host, tif.Port, tif.Dbname)

		dsns := []string{dsn}

		fmt.Println(dsn)

		exporter := NewExporter(dsns,opts...)
		defer func() {
			exporter.servers.Close()
		}()

		prometheus.MustRegister(version.NewCollector(exporterName))

		prometheus.MustRegister(exporter)

		pe, err := collector.NewPostgresCollector(
			logger,
			excludedDatabases,
			dsn,
			[]string{},
		)
		if err != nil {
			level.Warn(logger).Log("msg", "Failed to create GaussDBCollector", "err", err.Error())
		} else {
			prometheus.MustRegister(pe)
		}
		level.Warn(logger).Log("msg","register complete")

	} else {
		level.Warn(logger).Log("未检测到部署文件，不做任何改动，当前行为与官方包一致")
		if err := c.ReloadConfig(*configFile,logger); err != nil {
			level.Warn(logger).Log("msg", "Error loading config", "err", err)
		}

		dsns, err := getDataSources()
		if err != nil {
			level.Error(logger).Log("msg", "Failed reading data sources", "err", err.Error())
			os.Exit(1)
		}


		exporter := NewExporter(dsns, opts...)
		defer func() {
			exporter.servers.Close()
		}()

		prometheus.MustRegister(version.NewCollector(exporterName))

		prometheus.MustRegister(exporter)


		level.Info(logger).Log("route /metrics success")
		// TODO(@sysadmind): Remove this with multi-target support. We are removing multiple DSN support
		dsn := ""
		if len(dsns) > 0 {
			dsn = dsns[0]
		}

		pe, err := collector.NewPostgresCollector(
			logger,
			excludedDatabases,
			dsn,
			[]string{},
		)
		if err != nil {
			level.Warn(logger).Log("msg", "Failed to create DreamCollector", "err", err.Error())
		} else {
			prometheus.MustRegister(pe)
		}
	}

	http.Handle(*metricsPath, promhttp.Handler())

	level.Info(logger).Log("route /metrics success")
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "GaussDB Exporter",
			Description: "Prometheus GaussDB server Exporter",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	http.HandleFunc("/probe", handleProbe(logger, excludedDatabases))

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error running HTTP server", "err", err)
		os.Exit(1)
	}
}
