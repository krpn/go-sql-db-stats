package sqldbstats

import (
	"database/sql"
	pr "github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

var (
	initPromMetricsOnce = &sync.Once{}
	promMetric          = pr.NewGaugeVec(
		pr.GaugeOpts{
			Name: "sql_db_stats",
			Help: "SQL database stats",
		},
		[]string{"db_name", "metric"},
	)
)

type prometheus struct{ dbName string }

func newPrometheus(dbName string) *prometheus {
	initPromMetricsOnce.Do(func() { pr.MustRegister(promMetric) })
	return &prometheus{dbName: dbName}
}

func (p *prometheus) Collect(stats sql.DBStats) {
	promMetric.WithLabelValues(p.dbName, "max_open_connections").Set(float64(stats.MaxOpenConnections))
	promMetric.WithLabelValues(p.dbName, "open_connections").Set(float64(stats.OpenConnections))
	promMetric.WithLabelValues(p.dbName, "in_use_connections").Set(float64(stats.InUse))
	promMetric.WithLabelValues(p.dbName, "idle_connections").Set(float64(stats.Idle))
	promMetric.WithLabelValues(p.dbName, "wait_count").Set(float64(stats.WaitCount))
	promMetric.WithLabelValues(p.dbName, "wait_duration_seconds").Set(stats.WaitDuration.Seconds())
	promMetric.WithLabelValues(p.dbName, "idle_connections_closed").Set(float64(stats.MaxIdleClosed))
	promMetric.WithLabelValues(p.dbName, "lifetime_connections_closed").Set(float64(stats.MaxLifetimeClosed))
}

// StartCollectPrometheusMetrics starts metrics collection with Prometheus.
func StartCollectPrometheusMetrics(db StatsGetter, interval time.Duration, dbName string) CollectorStopper {
	return StartCollect(db, interval, newPrometheus(dbName))
}
