package postgres

import (
	"encoding/json"
	"time"
)

// TODO Are these overkill? It seems like a good idea to separate the peristed
//      data from the things that are "live" in the server, to e.g. enable
//      migrations and versioning.

// StoredServiceStatus is the things in core.ServiceStatus we need to persist
// to survive server restarts, upgrades, and (futurely) distributing Gatus
type StoredServiceStatus struct {
	Key             string
	Name, GroupName string
	// has at most 100 elements since the ServiceStatus code cleans itself up
	Results []*StoredResult // TODO diff between insert and select with no smart ORM
	// has at most 50 elements since the ServiceStatus code cleans itself up
	Events []*StoredEvent
	Uptime *StoredUptime
}

// String makes debugging and logging easier
func (s *StoredServiceStatus) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

// StoredResult is the things in core.Result we need to persist
type StoredResult struct {
	ID         int
	ServiceKey string
	Hostname   string
	Timestamp  time.Time
	Duration   time.Duration

	HTTPStatus       int
	ConditionResults []*StoredConditionResult
	Errors           []string
	Success          bool
}

// StoredConditionResult is a persisted core.ConditionResult
type StoredConditionResult struct {
	Condition string
	Success   bool
}

// StoredEvent is a persisted core.Event
type StoredEvent struct {
	Type      string
	Timestamp time.Time
}

// StoredUptime is a persisted core.Uptime
type StoredUptime struct {
	LastWeek float64
	LastDay  float64
	LastHour float64

	// has at most 240 elements since the Uptime code cleans itself up
	HourlyUptime map[int64]*StoredHourlyUptimeStatistics
}

// StoredHourlyUptimeStatistics is a persisted core.HourlyUptimeStatistics
type StoredHourlyUptimeStatistics struct {
	TotalExecutions             uint64
	SuccessfulExecutions        uint64
	TotalExecutionsResponseTime uint64
}
