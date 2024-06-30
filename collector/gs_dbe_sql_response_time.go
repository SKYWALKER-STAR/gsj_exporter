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

package collector

import (
	"context"
	"database/sql"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const SQLResponseSubsystem = "sql_response"

func init() {
	registerCollector(SQLResponseSubsystem, defaultEnabled, NewGSSQLResponseCollector)
}

type GSSQLResponseCollector struct {
	log log.Logger
}

func NewGSSQLResponseCollector(config collectorConfig) (Collector, error) {
	return &GSSQLResponseCollector{
		log: config.logger,
	}, nil
}

var (

	pgQueryP80 = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			SQLResponseSubsystem,
			"p80",
		),
		"80% SQL's response time",
		[]string{}, nil,
	)

	pgQueryP95 = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			SQLResponseSubsystem,
			"p95",
		),
		"95% SQL's response time",
		[]string{}, nil,
	)

	pgSQLResponse  = "SELECT * FROM dbe_perf.STATEMENT_RESPONSETIME_PERCENTILE;"
)

// Update implements Collector and exposes database size.
// It is called by the Prometheus registry when collecting metrics.
// The list of databases is retrieved from pg_database and filtered
// by the excludeSQLResponse config parameter. The tradeoff here is that
// we have to query the list of databases and then query the size of
// each database individually. This is because we can't filter the
// list of databases in the query because the list of excluded
// databases is dynamic.
func (c GSSQLResponseCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()

	// QueryResponse total amount of slow sql
	//rows, err := db.QueryResponseContext(ctx,
	//	pgSlowSQLQueryResponse,
	//)
	rows, err := db.QueryContext(ctx,
		pgSQLResponse,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p80,p95 sql.NullFloat64
		err := rows.Scan(
			&p80,
			&p95)

		if err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			pgQueryP80,
			prometheus.GaugeValue, 
			p80.Float64,
		)

		ch <- prometheus.MustNewConstMetric(
			pgQueryP95,
			prometheus.GaugeValue, 
			p95.Float64,
		)

	}
	return nil

}
