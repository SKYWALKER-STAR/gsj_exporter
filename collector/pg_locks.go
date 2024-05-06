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

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const locksSubsystem = "locks"

func init() {
	registerCollector(locksSubsystem, defaultEnabled, NewPGLocksCollector)
}

type PGLocksCollector struct {
	log log.Logger
}

func NewPGLocksCollector() (Collector, error) {
	return &PGLocksCollector{}, nil
}

var (
	pgLocksDesc = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			locksSubsystem,
			"count",
		),
		"Number of locks",
		[]string{"datname", "mode"}, nil,
	)

	pgLocksAvgWaitSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			locksSubsystem,
			"avg_wait_seconds",
		),
		"Average wait time for each session waiting for a lock",
		[]string{"datname", "mode"}, nil,
	)

	pgLocksTotalWaitSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			locksSubsystem,
			"total_wait_seconds",
		),
		"Total wait time for each session waiting for a lock",
		[]string{"datname", "mode"}, nil,
	)

	pgLocksQuery = `
		SELECT 
			pg_database.datname as datname,
			tmp.mode as mode,
			COALESCE(count, 0) as count,
			COALESCE(total_wait_seconds,0) AS total_wait_seconds,
			(CASE 
			WHEN COALESCE(count, 0)=0 THEN 0
			WHEN COALESCE(count, 0)!=0 THEN COALESCE(total_wait_seconds,0)/count
			END
			) AS avg_wait_seconds
		FROM 
			(
			VALUES 
				('accesssharelock'), 
				('rowsharelock'), 
				('rowexclusivelock'), 
				('shareupdateexclusivelock'), 
				('sharelock'), 
				('sharerowexclusivelock'), 
				('exclusivelock'), 
				('accessexclusivelock'), 
				('sireadlock')
			) AS tmp(mode)
			CROSS JOIN pg_database
			LEFT JOIN (
			SELECT 
				l.database, 
				lower(mode) AS mode, 
				count(*) AS count,
				sum(EXTRACT(EPOCH FROM (now()-a.query_start))) AS total_wait_seconds
			FROM 
				pg_locks l,
				pg_stat_activity a
			WHERE 1 = 1
				AND l.pid = a.pid
				AND l.database IS NOT NULL
				AND NOT l.GRANTED    --关键的指标是等待的锁
			GROUP BY 
				database, 
				lower(mode)
			) AS tmp2 ON tmp.mode = tmp2.mode
			and pg_database.oid = tmp2.DATABASE
		WHERE datname NOT IN ('template0','template1','security')
		ORDER BY
			1
	`
)

// Update implements Collector and exposes database locks.
// It is called by the Prometheus registry when collecting metrics.
func (c PGLocksCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	// Query the list of databases
	rows, err := db.QueryContext(ctx,
		pgLocksQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var datname, mode sql.NullString
	var count sql.NullInt64
	var totalWaitSeconds, avgWaitSeconds sql.NullFloat64

	for rows.Next() {
		if err := rows.Scan(&datname, &mode, &count, &totalWaitSeconds, &avgWaitSeconds); err != nil {
			return err
		}

		if !datname.Valid || !mode.Valid {
			continue
		}

		countMetric := 0.0
		if count.Valid {
			countMetric = float64(count.Int64)
		}

		totalWaitSecondsMetric := 0.0
		if totalWaitSeconds.Valid {
			totalWaitSecondsMetric = float64(totalWaitSeconds.Float64)
		}

		avgWaitSecondsMetric := 0.0
		if avgWaitSeconds.Valid {
			avgWaitSecondsMetric = float64(avgWaitSeconds.Float64)
		}

		ch <- prometheus.MustNewConstMetric(
			pgLocksDesc,
			prometheus.GaugeValue, countMetric,
			datname.String, mode.String,
		)

		ch <- prometheus.MustNewConstMetric(
			pgLocksTotalWaitSeconds,
			prometheus.GaugeValue, totalWaitSecondsMetric,
			datname.String, mode.String,
		)

		ch <- prometheus.MustNewConstMetric(
			pgLocksAvgWaitSeconds,
			prometheus.GaugeValue, avgWaitSecondsMetric,
			datname.String, mode.String,
		)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
