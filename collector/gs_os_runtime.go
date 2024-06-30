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

const osRuntimeSubsystem = "os_runtime"

func init() {
	registerCollector(osRuntimeSubsystem, defaultEnabled, NewGSOSRuntimeCollector)
}

type GSOSRuntimeCollector struct {
	log log.Logger
}

func NewGSOSRuntimeCollector(config collectorConfig) (Collector, error) {
	return &GSOSRuntimeCollector{
		log: config.logger,
	}, nil
}

var (
	gsOSRuntimeNumCPUS = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"num_cpus",
		),
		"",
		[]string{}, nil,
	)

	gsOSRuntimeNumCPUCores = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"num_cpu_cores",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeNumCPUSockets = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"num_cpu_sockets",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeCPUSecondTotal = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"cpu_seconds_total",
		),
		"",
		[]string{"mode"}, nil,
	)
	gsOSRuntimeIdleTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"idle_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeBusyTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"busy_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeUserTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"user_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeSysTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"sys_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeIowaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"iowait_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeNiceTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"nice_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeAvgIdleTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"avg_idle_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeAvgBusyTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"avg_busy_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeAvgUserTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"avg_user_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeAvgSysTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"avg_sys_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeAvgIowaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"avg_iowait_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeAvgNiceTime = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"avg_nice_time",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeVMPageInBytes = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"vm_page_in_bytes",
		),
		"",
		[]string{}, nil,
	)

	gsOSRuntimeVMPageOutBytes = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"vm_page_out_bytes",
		),
		"",
		[]string{}, nil,
	)
	gsOSRuntimeLOAD = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			osRuntimeSubsystem,
			"load",
		),
		"",
		[]string{}, nil,
	)

	gsOSRuntimeQuery = `select
	max(
	(case 
	  when name='NUM_CPUS' then value
	  else null
	end)) NUM_CPUS,
	max((case 
	  when name='NUM_CPU_CORES' then value
	  else null
	end)) NUM_CPU_CORES,
	max((case 
	  when name='NUM_CPU_SOCKETS' then value
	  else null
	end)) NUM_CPU_SOCKETS,
	max((case 
	  when name='IDLE_TIME' then value
	  else null
	end)) IDLE_TIME,
	max((case 
	  when name='BUSY_TIME' then value
	  else null
	end)) BUSY_TIME,
	max((case 
	  when name='USER_TIME' then value
	  else null
	end)) USER_TIME,
	max((case 
	  when name='SYS_TIME' then value
	  else null
	end)) SYS_TIME,
	max((case 
	  when name='IOWAIT_TIME' then value
	  else null
	end)) IOWAIT_TIME,
	max((case 
	  when name='NICE_TIME' then value
	  else null
	end)) NICE_TIME,
	max((case 
	  when name='AVG_IDLE_TIME' then value
	  else null
	end)) AVG_IDLE_TIME,
	max((case 
	  when name='AVG_BUSY_TIME' then value
	  else null
	end)) AVG_BUSY_TIME,
	max((case 
	  when name='AVG_USER_TIME' then value
	  else null
	end)) AVG_USER_TIME,
	max((case 
	  when name='AVG_SYS_TIME' then value
	  else null
	end)) AVG_SYS_TIME,
	max((case 
	  when name='AVG_IOWAIT_TIME' then value
	  else null
	end)) AVG_IOWAIT_TIME,
	max((case 
	  when name='AVG_NICE_TIME' then value
	  else null
	end)) AVG_NICE_TIME,
	max((case 
	  when name='VM_PAGE_IN_BYTES' then value
	  else null
	end)) VM_PAGE_IN_BYTES,
	max((case 
	  when name='VM_PAGE_OUT_BYTES' then value
	  else null
	end)) VM_PAGE_OUT_BYTES,
	max((case 
	  when name='LOAD' then value
	  else null
	end)) SYSLOAD
	from dbe_perf.os_runtime;`
)

func (c *GSOSRuntimeCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	row := db.QueryRowContext(ctx,
		gsOSRuntimeQuery)

	var num_cpus, num_cpu_cores, num_cpu_sockets sql.NullInt64
	var idle_time, busy_time, user_time, sys_time, iowait_time, nice_time, avg_idle_time, avg_busy_time, avg_user_time, avg_sys_time, avg_iowait_time, avg_nice_time, vm_page_in_bytes, vm_page_out_bytes, load sql.NullFloat64

	err := row.Scan(&num_cpus, &num_cpu_cores, &num_cpu_sockets, &idle_time, &busy_time, &user_time, &sys_time, &iowait_time, &nice_time, &avg_idle_time, &avg_busy_time, &avg_user_time, &avg_sys_time, &avg_iowait_time, &avg_nice_time, &vm_page_in_bytes, &vm_page_out_bytes, &load)
	if err != nil {
		return err
	}

	num_cpusMetric := 0.0
	if num_cpus.Valid {
		num_cpusMetric = float64(num_cpus.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeNumCPUS,
		prometheus.GaugeValue,
		num_cpusMetric,
	)

	num_cpu_coresMetric := 0.0
	if num_cpu_cores.Valid {
		num_cpu_coresMetric = float64(num_cpu_cores.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeNumCPUCores,
		prometheus.GaugeValue,
		num_cpu_coresMetric,
	)

	num_cpu_socketsMetric := 0.0
	if num_cpu_sockets.Valid {
		num_cpu_socketsMetric = float64(num_cpu_sockets.Int64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeNumCPUSockets,
		prometheus.GaugeValue,
		num_cpu_socketsMetric,
	)

	idle_timeMetric := 0.0
	if idle_time.Valid {
		idle_timeMetric = float64(idle_time.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeCPUSecondTotal,
		prometheus.GaugeValue,
		idle_timeMetric,
		"idle",
	)

	busy_timeMetric := 0.0
	if busy_time.Valid {
		busy_timeMetric = float64(busy_time.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeCPUSecondTotal,
		prometheus.GaugeValue,
		busy_timeMetric,
		"busy",
	)

	user_timeMetric := 0.0
	if user_time.Valid {
		user_timeMetric = float64(user_time.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeCPUSecondTotal,
		prometheus.GaugeValue,
		user_timeMetric,
		"user",
	)

	sys_timeMetric := 0.0
	if sys_time.Valid {
		sys_timeMetric = float64(sys_time.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeCPUSecondTotal,
		prometheus.GaugeValue,
		sys_timeMetric,
		"sys",
	)

	iowait_timeMetric := 0.0
	if iowait_time.Valid {
		iowait_timeMetric = float64(iowait_time.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeCPUSecondTotal,
		prometheus.GaugeValue,
		iowait_timeMetric,
		"iowait",
	)

	nice_timeMetric := 0.0
	if nice_time.Valid {
		nice_timeMetric = float64(nice_time.Float64)
	}
	ch <- prometheus.MustNewConstMetric(
		gsOSRuntimeCPUSecondTotal,
		prometheus.GaugeValue,
		nice_timeMetric,
		"nice",
	)

	return nil
}
