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
	_ "fmt"
	"context"
	"database/sql"

	_ "github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const MemorySubsystem = "memory"

func init() {
	registerCollector(MemorySubsystem, defaultEnabled, NewPGMemoryCollector)
}

type PGMemoryCollector struct {
}

func NewPGMemoryCollector() (Collector, error) {
	return &PGMemoryCollector{}, nil
}

var (
	gsProcessMemoryInfo = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			MemorySubsystem,
			"info",
		),
		"gaussdb memory info",
		[]string{"nodename","memtype"}, nil,
	)

	gsProcessMemoryInfoQuery	= "SELECT * FROM dbe_perf.MEMORY_NODE_DETAIL;"
)

// Update implements Collector and exposes database size.
// It is called by the Prometheus registry when collecting metrics.
// The list of databases is retrieved from pg_database and filtered
// by the excludeMemory config parameter. The tradeoff here is that
// we have to query the list of databases and then query the size of
// each database individually. This is because we can't filter the
// list of databases in the query because the list of excluded
// databases is dynamic.
func (c PGMemoryCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	
	db := instance.getDB()

	// Memory total amount of slow sql
	//rows, err := db.MemoryContext(ctx,
	//	pgSlowSQLMemory,
	//)
	rows, err := db.QueryContext(ctx,
		gsProcessMemoryInfoQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodename,memtype sql.NullString
		var membytes sql.NullFloat64

		err := rows.Scan(
			&nodename,
			&memtype,
			&membytes)

		if err != nil {
			return err
		}

		labels := []string{nodename.String,memtype.String}

		ch <- prometheus.MustNewConstMetric(
			gsProcessMemoryInfo,
			prometheus.GaugeValue, 
			membytes.Float64,
			labels...,
		)

	}
	return nil

}
