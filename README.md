# go-sql-db-stats


[![GoDoc](https://godoc.org/github.com/krpn/go-sql-db-stats?status.svg)](http://godoc.org/github.com/krpn/go-sql-db-stats)
[![Go Report](https://goreportcard.com/badge/github.com/krpn/go-sql-db-stats)](https://goreportcard.com/report/github.com/krpn/go-sql-db-stats)
[![License](https://img.shields.io/github/license/krpn/go-sql-db-stats.svg)](https://github.com/krpn/go-sql-db-stats/blob/master/LICENSE)

* A Go library for collecting [sql.DBStats](https://golang.org/pkg/database/sql/#DBStats) taken from *sql.DB connection
* Can collect metrics in Prometheus format
* Can repeatedly pass [sql.DBStats](https://golang.org/pkg/database/sql/#DBStats) in your own Collector (see [docs](http://godoc.org/github.com/krpn/go-sql-db-stats))

# Exposed Prometheus Metrics

| Name           | Description                    | Labels             |
|----------------|--------------------------------|--------------------|
| `sql_db_stats` | sql.DBStats metrics and values | `db_name` `metric` |


## Installation

```
go get github.com/krpn/go-sql-db-stats
```

## Example Usage

```go
package main

import (
	"database/sql"
	"fmt"
	"github.com/krpn/go-sql-db-stats"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

// docker run --name some-postgres -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres:11.2
var exampleConnString = fmt.Sprintf(
	"host=%v port=%v user=%v password=%v dbname=%v sslmode=disable binary_parameters=yes",
	"localhost", 5432, "postgres", "mysecretpassword", "postgres",
)

func main() {
	db, err := sql.Open("postgres", exampleConnString)
	if err != nil {
		panic(err)
	}
	defer func() {
		panic(db.Close())
	}()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	_ = sqldbstats.StartCollectPrometheusMetrics(db, 30*time.Second, "main_db")
	// If you want to stop collecting at any moment (for example: after close db), you may use this code:
	// collector := sqldbstats.StartCollectPrometheusMetrics(db, 30*time.Second, "main_db")
	// defer collector.Stop()

	http.HandleFunc("/handler", someHandler(db))
	http.Handle("/metrics", promhttp.Handler())
	panic(http.ListenAndServe(":8080", nil))
}

func someHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t time.Time
		err := db.QueryRow("select now();").Scan(&t)
		if err != nil {
			_, _ = fmt.Fprintf(w, "Result: %v", err)
			return
		}
		_, _ = fmt.Fprintf(w, "Result: %v", t)
	}
}

```