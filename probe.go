// Copyright 2022 The Prometheus Authors
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

	"github.com/prometheus-community/gaussdb_exporter/collector"
	"github.com/prometheus-community/gaussdb_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handleProbe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		params := r.URL.Query()
		instanceId := params.Get("instance_id")
		if instanceId == "" {
			http.Error(w, "instance_id is required", http.StatusBadRequest)
			utils.GetLogger().Warn("instance_id is required")
			return
		}

		dsn, err := getDataSourceById(instanceId)
		if err != nil {
			utils.GetLogger().Error("failed to configure target", "err", err)
			http.Error(w, fmt.Sprintf("could not configure dsn for target: %v", err), http.StatusBadRequest)
			return
		}

		registry := prometheus.NewRegistry()

		opts := []ExporterOpt{
			//返回func(e *Exporter)函数对象，设定传入的Exporter对象里面属性的值
			DisableDefaultMetrics(*disableDefaultMetrics),
			DisableSettingsMetrics(*disableSettingsMetrics),
		}
		//将opts传进来，其实就是初始化了exporter
		exporter := NewExporter([]string{dsn}, opts...)
		defer func() {
			exporter.servers.Close()
		}()

		registry.MustRegister(exporter)

		// Run the probe
		pc, err := collector.NewProbeCollector(registry, dsn)
		if err != nil {
			utils.GetLogger().Error("Error creating probe collector", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Cleanup underlying connections to prevent connection leaks
		defer pc.Close()

		// TODO(@sysadmind): Remove the registry.MustRegister() call below and instead handle the collection here. That will allow
		// for the passing of context, handling of timeouts, and more control over the collection.
		// The current NewProbeCollector() implementation relies on the MustNewConstMetric() call to create the metrics which is not
		// ideal to use without the registry.MustRegister() call.
		_ = ctx

		registry.MustRegister(pc)

		// TODO check success, etc
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
