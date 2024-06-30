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

const instanceTimeSubsystem = "instance_time"

func init() {
	registerCollector(instanceTimeSubsystem, defaultEnabled, NewGSInstanceTimeCollector)
}

type GSInstanceTimeCollector struct {
	log log.Logger
}

func NewGSInstanceTimeCollector(config collectorConfig) (Collector, error) {
	return &GSInstanceTimeCollector{
		log: config.logger,
	}, nil
}

var (
	gsInstanceTimeDBTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"db_time",
		),
		"",
		[]string{}, nil,
	)

	gsInstanceTimeCPUTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"cpu_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimeExecutionTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"execution_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimeParseTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"parse_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimePlanTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"plan_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimeRewriteTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"rewrite_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimePLExecutionTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"pl_execution_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimePLCompilationTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"pl_compilation_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimeNetSendTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"net_send_time",
		),
		"",
		[]string{}, nil,
	)
	gsInstanceTimeDataIOTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			instanceTimeSubsystem,
			"data_io_time",
		),
		"",
		[]string{}, nil,
	)

	gsInstanceTimeQuery = `select
	max(
	(case 
	  when stat_name='DB_TIME' then value
	  else null
	end)) DB_TIME,
	max((case 
	  when stat_name='CPU_TIME' then value
	  else null
	end)) CPU_TIME,
	max((case 
	  when stat_name='EXECUTION_TIME' then value
	  else null
	end)) EXECUTION_TIME,
	max((case 
	  when stat_name='PARSE_TIME' then value
	  else null
	end)) PARSE_TIME,
	max((case 
	  when stat_name='PLAN_TIME' then value
	  else null
	end)) PLAN_TIME,
	max((case 
	  when stat_name='REWRITE_TIME' then value
	  else null
	end)) REWRITE_TIME,
	max((case 
	  when stat_name='PL_EXECUTION_TIME' then value
	  else null
	end)) PL_EXECUTION_TIME,
	max((case 
	  when stat_name='PL_COMPILATION_TIME' then value
	  else null
	end)) PL_COMPILATION_TIME,
	max((case 
	  when stat_name='NET_SEND_TIME' then value
	  else null
	end)) NET_SEND_TIME,
	max((case 
	  when stat_name='DATA_IO_TIME' then value
	  else null
	end)) DATA_IO_TIME
	from dbe_perf.instance_time;`
)

func (c *GSInstanceTimeCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	row := db.QueryRowContext(ctx,
		gsInstanceTimeQuery)

	var db_time, cpu_time, execution_time, parse_time, plan_time, rewrite_time, pl_execution_time, pl_compilation_time, net_send_time, data_io_time sql.NullInt64

	err := row.Scan(&db_time, &cpu_time, &execution_time, &parse_time, &plan_time, &rewrite_time, &pl_execution_time, &pl_compilation_time, &net_send_time, &data_io_time)
	if err != nil {
		return err
	}

	db_timeMetric := 0.0
	if db_time.Valid {
		db_timeMetric = float64(db_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimeDBTime,
		prometheus.GaugeValue,
		db_timeMetric,
	)

	cpu_timeMetric := 0.0
	if cpu_time.Valid {
		cpu_timeMetric = float64(cpu_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimeCPUTime,
		prometheus.GaugeValue,
		cpu_timeMetric,
	)

	execution_timeMetric := 0.0
	if execution_time.Valid {
		execution_timeMetric = float64(execution_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimeExecutionTime,
		prometheus.GaugeValue,
		execution_timeMetric,
	)

	parse_timeMetric := 0.0
	if parse_time.Valid {
		parse_timeMetric = float64(parse_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimeParseTime,
		prometheus.GaugeValue,
		parse_timeMetric,
	)

	plan_timeMetric := 0.0
	if plan_time.Valid {
		plan_timeMetric = float64(plan_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimePlanTime,
		prometheus.GaugeValue,
		plan_timeMetric,
	)

	rewrite_timeMetric := 0.0
	if rewrite_time.Valid {
		rewrite_timeMetric = float64(rewrite_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimeRewriteTime,
		prometheus.GaugeValue,
		rewrite_timeMetric,
	)

	pl_execution_timeMetric := 0.0
	if pl_execution_time.Valid {
		pl_execution_timeMetric = float64(pl_execution_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimePLExecutionTime,
		prometheus.GaugeValue,
		pl_execution_timeMetric,
	)

	pl_compilation_timeMetric := 0.0
	if pl_compilation_time.Valid {
		pl_compilation_timeMetric = float64(pl_compilation_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimePLCompilationTime,
		prometheus.GaugeValue,
		pl_compilation_timeMetric,
	)

	net_send_timeMetric := 0.0
	if net_send_time.Valid {
		net_send_timeMetric = float64(net_send_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimeNetSendTime,
		prometheus.GaugeValue,
		net_send_timeMetric,
	)

	data_io_timeMetric := 0.0
	if data_io_time.Valid {
		data_io_timeMetric = float64(data_io_time.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsInstanceTimeDataIOTime,
		prometheus.GaugeValue,
		data_io_timeMetric,
	)

	return nil
}
