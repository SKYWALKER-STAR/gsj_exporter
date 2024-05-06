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
	"net/url"
	"regexp"
	"strings"

	"github.com/prometheus-community/gaussdb_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) discoverDatabaseDSNs() []string {
	// connstring syntax is complex (and not sure if even regular).
	// we don't need to parse it, so just superficially validate that it starts
	// with a valid-ish keyword pair
	connstringRe := regexp.MustCompile(`^ *[a-zA-Z0-9]+ *= *[^= ]+`)

	dsns := make(map[string]struct{})
	for _, dsn := range e.dsn {
		var dsnURI *url.URL
		var dsnConnstring string

		if strings.HasPrefix(dsn, "postgresql://") || strings.HasPrefix(dsn, "postgres://") {
			var err error
			dsnURI, err = url.Parse(dsn)
			if err != nil {
				utils.GetLogger().Error("Unable to parse DSN as URI", "dsn", loggableDSN(dsn), "err", err)
				continue
			}
		} else if connstringRe.MatchString(dsn) {
			dsnConnstring = dsn
		} else {
			utils.GetLogger().Error("Unable to parse DSN as either URI or connstring", "dsn", loggableDSN(dsn))
			continue
		}

		server, err := e.servers.GetServer(dsn)
		if err != nil {
			utils.GetLogger().Error("Error opening connection to database", "dsn", loggableDSN(dsn), "err", err)
			continue
		}
		dsns[dsn] = struct{}{}

		// If autoDiscoverDatabases is true, set first dsn as master database (Default: false)
		server.master = true

		databaseNames, err := queryDatabases(server)
		if err != nil {
			utils.GetLogger().Error("Error querying databases", "dsn", loggableDSN(dsn), "err", err)
			continue
		}
		for _, databaseName := range databaseNames {
			if contains(e.excludeDatabases, databaseName) {
				continue
			}

			if len(e.includeDatabases) != 0 && !contains(e.includeDatabases, databaseName) {
				continue
			}

			if dsnURI != nil {
				dsnURI.Path = databaseName
				dsn = dsnURI.String()
			} else {
				// replacing one dbname with another is complicated.
				// just append new dbname to override.
				dsn = fmt.Sprintf("%s dbname=%s", dsnConnstring, databaseName)
			}
			dsns[dsn] = struct{}{}
		}
	}

	result := make([]string, len(dsns))
	index := 0
	for dsn := range dsns {
		result[index] = dsn
		index++
	}

	return result
}

func (e *Exporter) scrapeDSN(ch chan<- prometheus.Metric, dsn string) error {
	server, err := e.servers.GetServer(dsn)

	if err != nil {
		return &ErrorConnectToServer{fmt.Sprintf("Error opening connection to database (%s): %s", loggableDSN(dsn), err.Error())}
	}

	// Check if autoDiscoverDatabases is false, set dsn as master database (Default: false)
	if !e.autoDiscoverDatabases {
		server.master = true
	}

	// Check if map versions need to be updated
	if err := e.checkMapVersions(ch, server); err != nil {
		utils.GetLogger().Warn("Proceeding with outdated query maps, as the Postgres version could not be determined", "err", err)
	}

	return server.Scrape(ch, e.disableSettingsMetrics)
}

// try to get the DataSource
func getDataSourceById(instanceId string) (string, error) {
	cfg := cfgHandler.GetConfig()
	instances := cfg.Instances
	for i := 0; i < len(instances); i++ {
		instance := instances[i]
		if instanceId == instance.InstanceId {
			return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", instance.Host, instance.Port, instance.Db, instance.User, instance.Password), nil
		}
	}

	return "", fmt.Errorf("根据实例ID（instance_id=%s）并没有找到对应配置的实例！", instanceId)
}
