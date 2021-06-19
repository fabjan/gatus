package postgres

import (
	"errors"
	"time"

	"github.com/TwinProduction/gatus/core"
)

// TODO Are these overkill? It seems like a good idea to separate the peristed
//      data from the things that are "live" in the server, to e.g. enable
//      migrations and versioning.

// StoredServiceStatus is the things in core.ServiceStatus we need to persist
// to survive server restarts, upgrades, and (futurely) distributing Gatus
type StoredServiceStatus struct {
	Key         string
	Name, Group string
	// has at most 100 elements since the ServiceStatus code cleans itself up
	Results []*StoredResult
	// has at most 50 elements since the ServiceStatus code cleans itself up
	Events []*StoredEvent
	Uptime StoredUptime
}

// StoredResult is the things in core.Result we need to persist
type StoredResult struct {
	Hostname  string
	Timestamp time.Time
	Duration  time.Duration

	HTTPStatus       int
	ConditionResults []*StoredConditionResult
	Errors           []string
	Success          bool
}

// StoredEvent is the things in core.Event we need to persist
type StoredEvent struct {
	Type      core.EventType
	Timestamp time.Time
}

// StoredUptime is the things in core.Uptime we need to persist
type StoredUptime struct {
	LastSevenDays       float64
	LastTwentyFourHours float64
	LastHour            float64

	// has at most 240 elements since the Uptime code cleans itself up
	HourlyStatistics map[int64]*core.HourlyUptimeStatistics
}

// StoredConditionResult is the things in core.ConditionResult we need to persist
type StoredConditionResult struct {
	Condition core.Condition
	Success   bool
}

// Store that leverages a Postgres database
type Store struct {
	table string
}

// NewStore creates a new store
func NewStore(table string) (*Store, error) {
	if table == "" {
		return nil, errors.New("table name can't be empty")
	}

	store := &Store{
		table: table,
	}

	// If we add in-memory cache to support the database and minimize latency,
	// we could initialize a load here. Note that if we add support for
	// distributing Gatus over several nodes/instances, they need a way to
	// keep their caches in sync.

	// Perhaps the consistency requirements are note very high, a few seconds
	// lag should not affect make the app less useful.

	return store, nil
}

// GetAllServiceStatusesWithResultPagination returns the JSON encoding of all monitored core.ServiceStatus
// with a subset of core.Result defined by the page and pageSize parameters
func (s *Store) GetAllServiceStatusesWithResultPagination(page, pageSize int) map[string]*core.ServiceStatus {
	// if we relax our consistency we can leverage the memory store here
	panic("TODO: implement")
}

// GetServiceStatus returns the service status for a given service name in the given group
func (s *Store) GetServiceStatus(groupName, serviceName string) *core.ServiceStatus {
	// if we relax our consistency we can leverage the memory store here
	panic("TODO: implement")
}

// GetServiceStatusByKey returns the service status for a given key
func (s *Store) GetServiceStatusByKey(key string) *core.ServiceStatus {
	// if we relax our consistency we can leverage the memory store here
	panic("TODO: implement")
}

// Insert adds the observed result for the specified service into the store
func (s *Store) Insert(service *core.Service, result *core.Result) {
	// if we relax our consistency we can leverage the memory store here
	// and only save to database to the next Save()
	panic("TODO: implement")
}

// DeleteAllServiceStatusesNotInKeys removes all ServiceStatus that are not within the keys provided
//
// Used to delete services that have been persisted but are no longer part of the configured services
func (s *Store) DeleteAllServiceStatusesNotInKeys(keys []string) int {
	// if we relax our consistency we can leverage the memory store here
	// and defer the purge from the database to the next Save()
	panic("TODO: implement")
}

// Clear deletes everything from the store
func (s *Store) Clear() {
	// if we relax our consistency we can leverage the memory store here
	// and defer the purge from the database to the next Save()
	panic("TODO: implement")
}

// Save persists the data if and where it needs to be persisted
func (s *Store) Save() error {
	// we can no longer defer database updates!
	panic("TODO: implement")
}
