package metriks

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/armon/go-metrics"

	"github.com/netlify/netlify-commons/util"
)

type DBStats interface {
	Start()
	Stop()
}

type dBStats struct {
	db     *sql.DB
	name   string
	labels []metrics.Label
	sched  util.ScheduledExecutor
}

const defaultTickTime = 2 * time.Second

// NewDBStats returns a managed object that when Start() is invoked, will periodically
// report the stats from the passed DB object.
//
// Stop() should be called before the DB is closed.
func NewDBStats(db *sql.DB, name string, labels []metrics.Label) DBStats {
	dbstats := &dBStats{
		db:     db,
		name:   fmt.Sprintf("dbstats.%s", name),
		labels: labels,
	}

	dbstats.sched = util.NewScheduledExecutor(defaultTickTime, dbstats.emitStats)

	return dbstats
}

func (d *dBStats) Start() {
	d.sched.Start()
}

func (d *dBStats) Stop() {
	d.sched.Stop()
}

func (d *dBStats) emitStats() {
	stats := d.db.Stats()

	d.emitStat("MaxOpenConnections", float32(stats.MaxOpenConnections))

	d.emitStat("OpenConnections", float32(stats.OpenConnections))
	d.emitStat("InUse", float32(stats.InUse))
	d.emitStat("Idle", float32(stats.Idle))

	d.emitStat("WaitCount", float32(stats.WaitCount))
	d.emitStat("WaitDuration_us", float32(stats.WaitDuration.Nanoseconds()/1000))
	d.emitStat("MaxIdleClosed", float32(stats.MaxIdleClosed))
	d.emitStat("MaxLifetimeClosed", float32(stats.MaxLifetimeClosed))
}

func (d *dBStats) emitStat(statName string, value float32) {
	labels := append(d.labels, L("db_stat", statName))
	SampleLabels(d.name, labels, value)
}
