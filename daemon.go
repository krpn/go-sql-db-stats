package sqldbstats

import (
	"database/sql"
	"sync"
	"time"
)

// Collector collects metrics.
type Collector interface {
	Collect(sql.DBStats)
}

// CollectorStopper can stop collecting with or without metrics reset.
type CollectorStopper interface {
	// Stop stops collecting.
	Stop()
	// StopAndResetMetrics stops collecting and sends zero sql.DBStats struct to Collector.
	StopAndResetMetrics()
}

// StatsGetter gets sql.DBStats.
// It's just a *sql.DB.
type StatsGetter interface {
	Stats() sql.DBStats
}

// StartCollect starts metrics collection with passed Collector.
func StartCollect(db StatsGetter, interval time.Duration, collector Collector) CollectorStopper {
	cd := &collectorDaemon{
		stop:      make(chan struct{}),
		stopped:   false,
		wg:        &sync.WaitGroup{},
		mt:        &sync.Mutex{},
		collector: collector,
	}
	cd.wg.Add(1)
	go cd.collect(db, interval)
	return cd
}

type collectorDaemon struct {
	stop      chan struct{}
	stopped   bool
	wg        *sync.WaitGroup
	mt        *sync.Mutex
	collector Collector
}

func (cd *collectorDaemon) collect(db StatsGetter, interval time.Duration) {
	defer cd.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	cd.collector.Collect(db.Stats())
	for {
		select {
		case <-ticker.C:
			cd.collector.Collect(db.Stats())
		case <-cd.stop:
			return
		}
	}
}

func (cd *collectorDaemon) Stop() {
	cd.mt.Lock()
	defer cd.mt.Unlock()
	if cd.stopped {
		return
	}
	cd.stop <- struct{}{}
	cd.wg.Wait()
	cd.stopped = true
}

func (cd *collectorDaemon) StopAndResetMetrics() {
	cd.Stop()
	cd.collector.Collect(sql.DBStats{})
}
