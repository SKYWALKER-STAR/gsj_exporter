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
	"log"
	"os"
	"net/http"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/gaussdb_exporter/utils"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"

	"myencrypt"
	"config"
)

var (
	cfgHandler = config.Handler{}
	logger = log.New(os.Stdout,"",log.Lshortfile | log.Ldate | log.Ltime)

	listenAddress          = kingpin.Flag("web.listen-address", "sever listen port.").Default(":9334").Int16()
	metricsPath            = kingpin.Flag("web.metrics-path", "Path under which to expose metrics.").Default("/metrics").Envar("PG_EXPORTER_WEB_TELEMETRY_PATH").String()
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

	var target_info myencrypt.Target_info

	kingpin.Version(version.Print(exporterName))
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	/* 一体化监控平台纳管逻辑 */
	if fi,err := os.Stat("HISTORY");err == nil || os.IsExist(err) {
		logger.Println("监测到HISTORY文件，执行新模式")

		if fi.Size() < 1 {
			logger.Println("HISTORY文件是空的，部署有问题，无法继续")
			os.Exit(1)
		}

		/* 从一体化监控平台接收参数并且生成配置文件 */
		target_info = myencrypt.CheckInstall(fmt.Sprintf(":%d",*listenAddress))
		config_file := cfgHandler.InitFromUrl(target_info.InstanceId,target_info.Excludedbs,target_info.Hosts,target_info.Port,target_info.DB,target_info.User,target_info.Password,"0",1)
		cfgHandler.WriteConfigFile(config_file)
		if err := cfgHandler.ReloadConfig(); err != nil {
			logger.Println(err)
			utils.GetLogger().Warn("Error loading config", "err", err)
		}


	} else {
		logger.Println("未监测到部署文件，不做任何改动，当前行为与官方包一致")
		if err := cfgHandler.ReloadConfig(); err != nil {
			logger.Println(err)
			utils.GetLogger().Warn("Error loading config", "err", err)
		}
	}

	/* 注册/metrics接口，handleProbe作为处理函数 */
	http.HandleFunc(*metricsPath, handleProbe(target_info.InstanceId))

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
		logger.Println("ListenAndServe err:", err)
		utils.GetLogger().Error("msg:", "无法启动web容器服务", "error:", err)
	}
}
