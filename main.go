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
	"os"
	"net/http"

	"github.com/alecthomas/kingpin/v2"
	//"github.com/prometheus-community/gaussdb_exporter/config"
	"github.com/prometheus-community/gaussdb_exporter/utils"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"

	"myencrypt"
	"config"
)

var (
	cfgHandler = config.Handler{}

	//获取命令行输入的参数值
	//webConfig = kingpinflag.AddFlags(kingpin.CommandLine, ":9334")
	listenAddress          = kingpin.Flag("web.listen-address", "sever listen port.").Default("9334").Int16()
	metricsPath            = kingpin.Flag("web.metrics-path", "Path under which to expose metrics.").Default("/gaussdb/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
	disableDefaultMetrics  = kingpin.Flag("disable-default-metrics", "Do not include default metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_DEFAULT_METRICS").Bool()
	disableSettingsMetrics = kingpin.Flag("disable-settings-metrics", "Do not include pg_settings metrics.").Default("false").Envar("PG_EXPORTER_DISABLE_SETTINGS_METRICS").Bool()
	metricPrefix           = kingpin.Flag("metric-prefix", "A metric prefix can be used to have non-default (not \"gs\") prefixes for each of the metrics").Default("gs").Envar("PG_EXPORTER_METRIC_PREFIX").String()
)

// Metric name parts.
const (
	// Namespace for all metrics.
	namespace = "gs"
	// Subsystems.
	exporter = "exporter"
	// The name of the exporter.
	exporterName = "gaussdb_exporter"
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

	if fi,err := os.Stat("HISTORY");err == nil || os.IsExist(err) {
		fmt.Println("监测到HISTORY文件，执行新模式")

		if fi.Size() < 1 {
			fmt.Println("HISTORY文件是空的，部署有问题，无法继续")
			os.Exit(1)
		}

		target_info := myencrypt.CheckInstall(fmt.Sprintf(":%d",*listenAddress))

		config_file := cfgHandler.InitFromUrl(target_info.InstanceId,target_info.Excludedbs,target_info.Hosts,target_info.Port,target_info.DB,target_info.User,target_info.Password,"0",1)

		cfgHandler.WriteConfigFile(config_file)
		if err := cfgHandler.ReloadConfig(); err != nil {
			fmt.Println(err)
			utils.GetLogger().Warn("Error loading config", "err", err)
		}


	} else {
		fmt.Println("未监测到部署文件，不做任何改动，当前行为与官方包一致")
		if err := cfgHandler.ReloadConfig(); err != nil {
			fmt.Println(err)
			utils.GetLogger().Warn("Error loading config", "err", err)
		}
	}

	/*
		dsn, err := getDataSourceById("opengauss_instance_1")
		if err != nil {
			fmt.Println(err)
		}
		//http.HandleFunc("/metrics", handleProbe(logger, excludedDatabases))

		pe, err := collector.NewPostgresCollector(
			dsn,
			[]string{},
		)
		if err != nil {
			utils.GetLogger().Warn("Failed to create PostgresCollector", "err", err.Error())
		} else {
			prometheus.MustRegister(pe)
		}
	*/
	//http.Handle(*metricsPath, promhttp.Handler())

	http.HandleFunc(*metricsPath, handleProbe())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `<html>
		<head><title>HuaWei GaussDB Exporter</title></head>
		<body>
		<h2>Welcome to use HuaWei GaussDB Exporter</h2>
		<h2>This exporter was developed by Li Hongjie from Digitalgd Company</h2>
		</body>
		</html>`)
	})
	utils.GetLogger().Info("服务已经启动", "server listen port", *listenAddress)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *listenAddress), nil)
	if err != nil {
		fmt.Println("ListenAndServe err:", err)
		utils.GetLogger().Error("msg:", "无法启动web容器服务", "error:", err)
	}
	/*
		srv := &http.Server{}
		if err := web.ListenAndServe(srv, webConfig, nil); err != nil {
			utils.GetLogger().Error("Error running HTTP server", "err", err)
			os.Exit(1)
		} else {
			utils.GetLogger().Info("服务已经启动")
		}
	*/
}
