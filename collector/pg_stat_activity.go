// Copyright 2023 The Prometheus Authors
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

package collector

import (
	"context"
	"database/sql"

	"github.com/prometheus-community/gaussdb_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const statActivitySubsystem = "stat_activity"

func init() {
	registerCollector(statActivitySubsystem, defaultEnabled, NewPGStatActivityCollector)
}

type PGStatActivityCollector struct {
}

func NewPGStatActivityCollector() (Collector, error) {
	return &PGStatActivityCollector{}, nil
}

var (
	pgStatActivityCount = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statActivitySubsystem,
			"count",
		),
		"",
		[]string{"datname", "state", "usename", "application_name"}, nil,
	)

	pgStatActivityMaxTxDuration = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			statActivitySubsystem,
			"max_tx_duration",
		),
		"",
		[]string{"datname", "state", "usename", "application_name"}, nil,
	)

	pgStatActivityQuery = `SELECT
	pg_database.datname,
	tmp.state,
	tmp2.usename,
	tmp2.application_name,
	COALESCE(count,0) as count,
	COALESCE(max_tx_duration,0) as max_tx_duration
FROM
	(
	  VALUES ('active'),
			   ('idle'),
			   ('idle in transaction'),
			   ('idle in transaction (aborted)'),
			   ('fastpath function call'),
			   ('disabled')
	) AS tmp(state) CROSS JOIN pg_database
LEFT JOIN
(
	SELECT
		datname,
		state,
		usename,
		application_name,
		count(*) AS count,
		MAX(EXTRACT(EPOCH FROM now() - xact_start))::float AS max_tx_duration
	FROM pg_stat_activity GROUP BY datname,state,usename,application_name) AS tmp2
	ON tmp.state = tmp2.state AND pg_database.datname = tmp2.datname
	WHERE pg_database.datname NOT IN ('template0','template1','security')`
)

func (c *PGStatActivityCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		pgStatActivityQuery,
	)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var datname, state, usename, application_name sql.NullString
		var count, max_tx_duration sql.NullFloat64

		err := rows.Scan(
			&datname,
			&state,
			&usename,
			&application_name,
			&count,
			&max_tx_duration,
		)
		if err != nil {
			return err
		}

		if !datname.Valid {
			utils.GetLogger().Debug("Skipping collecting metric because it has no datid")
			continue
		}
		if !usename.Valid {
			utils.GetLogger().Debug("Skipping collecting metric because it has no usename")
			continue
		}
		if !count.Valid {
			utils.GetLogger().Debug("Skipping collecting metric because it has no count")
			continue
		}
		if !max_tx_duration.Valid {
			utils.GetLogger().Debug("Skipping collecting metric because it has no max_tx_duration")
			continue
		}

		labels := []string{datname.String, state.String, usename.String, application_name.String}
		ch <- prometheus.MustNewConstMetric(
			pgStatActivityCount,
			prometheus.GaugeValue, count.Float64, labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			pgStatActivityMaxTxDuration,
			prometheus.GaugeValue, max_tx_duration.Float64, labels...,
		)
	}

	return nil
}
