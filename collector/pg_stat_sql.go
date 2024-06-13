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

	_ "github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const QuerySubsystem = "query"

func init() {
	registerCollector(QuerySubsystem, defaultEnabled, NewPGQueryCollector)
}

type PGQueryCollector struct {
}

func NewPGQueryCollector() (Collector, error) {
	return &PGQueryCollector{}, nil
}

var (
	pgSlowSQL = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			QuerySubsystem,
			"slow_sql",
		),
		"slow sql information",
		[]string{"datname","pid","sessionid","usename","application_name","client_addr","client_port","query_id","query","query_start"}, nil,
	)

	pgSlowSQLTotal = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			QuerySubsystem,
			"slow_sql_total",
		),
		"Total number of slow sql",
		[]string{}, nil,
	)

	pgSlowSQLQuery			= "SELECT count(*) FROM pg_stat_activity WHERE state != 'idle' AND now() - query_start > interval '5s'"
	//pgSlowSQLQuery			= "SELECT datname,pid,sessionid,usename,application_name,client_addr,client_port,query_id,query,query_start,(SELECT count(*) FROM pg_stat_activity WHERE state != 'idle' AND now() - query_start > interval '5s') FROM pg_stat_activity WHERE state != 'idle' AND now()-query_start > interval '5s' ORDER BY query_start;"
)

// Update implements Collector and exposes database size.
// It is called by the Prometheus registry when collecting metrics.
// The list of databases is retrieved from pg_database and filtered
// by the excludeQuery config parameter. The tradeoff here is that
// we have to query the list of databases and then query the size of
// each database individually. This is because we can't filter the
// list of databases in the query because the list of excluded
// databases is dynamic.
func (c PGQueryCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	
	db := instance.getDB()

	// Query total amount of slow sql
	//rows, err := db.QueryContext(ctx,
	//	pgSlowSQLQuery,
	//)
	rows, err := db.QueryContext(ctx,
		pgSlowSQLQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {

		/*
		var datname,pid,sessionid,usename,application_name,client_addr,client_port,query_id,query,query_start sql.NullString
		var count sql.NullFloat64
		err := rows.Scan(
			&datname,
			&pid,
			&sessionid,
			&usename,
			&application_name,
			&client_addr,
			&client_port,
			&query_id,
			&query,
			&query_start,
			&count)
		*/

		var count sql.NullFloat64
		err := rows.Scan(
			&count)

		if err != nil {
			return err
		}

		//labels := []string{datname.String,pid.String,sessionid.String,usename.String,application_name.String,client_addr.String,client_port.String,query_id.String,query.String,query_start.String}

		ch <- prometheus.MustNewConstMetric(
			pgSlowSQLTotal,
			prometheus.CounterValue,
			//count.Float64,
			count.Float64,
			//labels...,
		)

	}
	return nil
}
