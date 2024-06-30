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
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const waitEventsSubsystem = "wait_events"

func init() {
	registerCollector(waitEventsSubsystem, defaultEnabled, NewGSWaitEventsCollector)
}

type GSWaitEventsCollector struct {
	log log.Logger
}

func NewGSWaitEventsCollector(config collectorConfig) (Collector, error) {
	return &GSWaitEventsCollector{
		log: config.logger,
	}, nil
}

var (
	gsWaitEventsWait = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			waitEventsSubsystem,
			"wait",
		),
		"",
		[]string{"type", "event"}, nil,
	)
	gsWaitEventsFailedWait = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			waitEventsSubsystem,
			"failed_wait",
		),
		"",
		[]string{"type", "event"}, nil,
	)
	gsWaitEventsTotalWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			waitEventsSubsystem,
			"total_wait_time",
		),
		"",
		[]string{"type", "event"}, nil,
	)

	gsWaitEventsQuery = `select type,event,wait,failed_wait,total_wait_time from dbe_perf.wait_events where wait <> 0`
)

func (c *GSWaitEventsCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		gsWaitEventsQuery,
	)

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var wait_type, event sql.NullString
		var wait, failed_wait, total_wait_time sql.NullInt64

		err := rows.Scan(
			&wait_type,
			&event,
			&wait,
			&failed_wait,
			&total_wait_time,
		)
		if err != nil {
			return err
		}

		if !wait_type.Valid {
			level.Info(c.log).Log("Skipping collecting metric because it has no type")
			continue
		}
		if !event.Valid {
			level.Info(c.log).Log("Skipping collecting metric because it has no event")
			continue
		}

		waitMetric := 0.0
		if wait.Valid {
			waitMetric = float64(wait.Int64)
		}
		failedWaitMetric := 0.0
		if failed_wait.Valid {
			failedWaitMetric = float64(failed_wait.Int64)
		}
		total_wait_timeMetric := 0.0
		if total_wait_time.Valid {
			total_wait_timeMetric = float64(total_wait_time.Int64)
		}
		ch <- prometheus.MustNewConstMetric(
			gsWaitEventsWait,
			prometheus.GaugeValue,
			waitMetric,
			wait_type.String, event.String,
		)
		ch <- prometheus.MustNewConstMetric(
			gsWaitEventsFailedWait,
			prometheus.GaugeValue,
			failedWaitMetric,
			wait_type.String, event.String,
		)
		ch <- prometheus.MustNewConstMetric(
			gsWaitEventsTotalWaitTime,
			prometheus.GaugeValue,
			total_wait_timeMetric,
			wait_type.String, event.String,
		)
	}

	return nil
}
