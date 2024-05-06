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

	"github.com/prometheus/client_golang/prometheus"
)

const replicationSubsystem = "replication"

func init() {
	registerCollector(replicationSubsystem, defaultEnabled, NewPGReplicationCollector)
}

type PGReplicationCollector struct {
}

func NewPGReplicationCollector() (Collector, error) {
	return &PGReplicationCollector{}, nil
}

var (
	pgReplicationLag = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSubsystem,
			"lag_bytes",
		),
		"Replication lag behind master in bytes",
		[]string{"replic_node_addr"}, nil,
	)
	pgReplicationIsReplica = prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			replicationSubsystem,
			"is_replica",
		),
		"is current instance replica or master",
		[]string{}, nil,
	)

	pgReplicationQuery = `SELECT
    client_addr as replic_node_addr,
	CASE
		WHEN pg_is_in_recovery() THEN 0  --主从延迟，仅以主节点的视角来看。从节点不存在主从延迟
		ELSE pg_xlog_location_diff(sender_sent_location,receiver_replay_location)
	END AS lag_bytes  --延迟字节数
FROM pg_stat_replication`

	pgReplicationQuery2 = `SELECT 
	CASE
		WHEN pg_is_in_recovery() THEN 1   --在recovery状态的，则为从节点
		ELSE 0
	END AS is_replica`
)

func (c *PGReplicationCollector) Update(ctx context.Context, instance *instance, ch chan<- prometheus.Metric) error {
	db := instance.getDB()
	rows, err := db.QueryContext(ctx,
		pgReplicationQuery,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var lag_bytes float64
		var replic_node_addr string
		err := rows.Scan(&replic_node_addr, &lag_bytes)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			pgReplicationLag,
			prometheus.GaugeValue, lag_bytes, replic_node_addr,
		)
	}

	row := db.QueryRowContext(ctx,
		pgReplicationQuery2)

	var isReplica sql.NullInt64
	err = row.Scan(&isReplica)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		pgReplicationIsReplica,
		prometheus.GaugeValue, float64(isReplica.Int64),
	)

	return nil
}
