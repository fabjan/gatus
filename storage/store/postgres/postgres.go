package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/TwinProduction/gatus/core"
	"github.com/TwinProduction/gatus/storage/store/memory"
	"github.com/TwinProduction/gatus/util"
	"github.com/jackc/pgtype"
)

// Store that leverages a Postgres database
type Store struct {
	db        *sql.DB
	namespace string
	cache     *memory.Store
}

// NewStore creates a new store
func NewStore(db *sql.DB, namespace string) (*Store, error) {

	if namespace == "" {
		return nil, errors.New("Table namespace can't be empty")
	}

	// We leverage the functinality of the in-memory store for
	// everything but persistence, so we won't give it a filename.
	cache, err := memory.NewStore("")
	if err != nil {
		return nil, fmt.Errorf("Cannot intialize in-memory cache: %w", err)
	}

	store := &Store{
		db:        db,
		namespace: namespace,
		cache:     cache,
	}

	err = store.mustCreateTables()
	if err != nil {
		return nil, fmt.Errorf("Cannot create tables: %w", err)
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

// InsertServiceStatus encodes the service status as a database row, and inserts it in the services table.
// N.B. This is only public for testing purposes.
func (s *Store) InsertServiceStatus(ss *StoredServiceStatus) error {

	var eventsJSON stringslice
	for _, e := range ss.Events {
		bytes, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("Cannot marshal service status event (%v): %w", e, err)
		}
		eventsJSON = append(eventsJSON, string(bytes))
	}

	bytes, err := json.Marshal(ss.Uptime)
	if err != nil {
		return fmt.Errorf("Cannot marshal service uptime (%v): %w", ss.Uptime, err)
	}
	uptimeJSON := string(bytes)

	key := util.ConvertGroupAndServiceToKey(ss.GroupName, ss.Name)

	insert := fmt.Sprintf(`INSERT INTO %s (
		key,
		name,
		group_name,
		events_json,
		uptime_json
	) VALUES ($1, $2, $3, $4, $5)`,
		s.ServicesTable(),
	)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("Cannot open transaction: %w", err)
	}

	res, err := tx.Exec(insert, key, ss.Name, ss.GroupName, eventsJSON, uptimeJSON)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Cannot insert service status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Cannot verify service status was inserted: %w", err)
	}
	if n != 1 {
		tx.Rollback()
		return fmt.Errorf("Expected 1 row affected by service status insert, got %v", n)
	}

	for _, r := range ss.Results {
		r.ServiceKey = key
		err := s.insertResult(tx, r)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("Cannot insert results from service status: %v", err)
		}
	}

	tx.Commit()
	return nil
}

// FindServiceStatus reads from the service status table and decodes the column data.
// N.B. This is only public for testing purposes.
func (s *Store) FindServiceStatus(serviceKey string) (*StoredServiceStatus, error) {

	q := fmt.Sprintf(`SELECT
			name,
			group_name,
			events_json,
			uptime_json
		FROM %s
		WHERE key = $1`,
		s.ServicesTable(),
	)

	rows, err := s.db.Query(q, serviceKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot select service status: %w", err)
	}
	defer rows.Close()

	ss := StoredServiceStatus{}

	for rows.Next() {
		var eventsJSON pgtype.TextArray
		var uptimeJSON string
		err := rows.Scan(&ss.Name, &ss.GroupName, &eventsJSON, &uptimeJSON)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan a service status from the database: %w", err)
		}

		ss.Uptime = &StoredUptime{}
		json.Unmarshal([]byte(uptimeJSON), ss.Uptime)

		ss.Events = make([]*StoredEvent, len(eventsJSON.Elements))
		for i, j := range eventsJSON.Elements {
			e := StoredEvent{}
			err := json.Unmarshal([]byte(j.String), &e)
			if err != nil {
				return nil, fmt.Errorf("Failed to unmarshal service status events: %w", err)
			}
			ss.Events[i] = &e
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Cannot read service status from the database: %w", err)
	}

	results, err := s.findResults(serviceKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot read service status results from the database: %w", err)
	}

	ss.Results = results

	return &ss, nil
}

func (s *Store) insertResult(tx *sql.Tx, r *StoredResult) error {

	var conditionsJSON stringslice
	for _, c := range r.ConditionResults {
		bytes, err := json.Marshal(c)
		if err != nil {
			return fmt.Errorf("Cannot marshal condition result (%v): %w", c, err)
		}
		conditionsJSON = append(conditionsJSON, string(bytes))
	}

	insert := fmt.Sprintf(`INSERT INTO %s (
		service_key,
		hostname,
		timestamp_ns,
		duration_ns,
		http_status,
		conditions_json,
		errors,
		success
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ResultsTable(),
	)

	res, err := tx.Exec(insert, r.ServiceKey, r.Hostname, r.Timestamp.UnixNano(), r.Duration.Nanoseconds(), r.HTTPStatus, conditionsJSON, r.Errors, r.Success)
	if err != nil {
		return fmt.Errorf("Cannot insert result: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Cannot verify result was inserted: %w", err)
	}
	if n != 1 {
		return fmt.Errorf("Expected 1 row affected by result insert, got %v", n)
	}

	return nil
}

func (s *Store) findResults(serviceKey string) ([]*StoredResult, error) {

	// TODO This could probably be batched, taking a slice of service keys
	//      and returning a map from service key to results instead.

	q := fmt.Sprintf(`SELECT
			id,
			service_key,
			hostname,
			timestamp_ns,
			duration_ns,
			http_status,
			conditions_json,
			errors,
			success
		FROM %s
		WHERE service_key = $1`,
		s.ResultsTable(),
	)

	rows, err := s.db.Query(q, serviceKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot select results: %w", err)
	}
	defer rows.Close()

	results := []*StoredResult{}

	for rows.Next() {
		r := StoredResult{}
		var ts, dur int64
		var conditionsJSON, errors pgtype.TextArray
		err := rows.Scan(&r.ID, &r.ServiceKey, &r.Hostname, &ts, &dur, &r.HTTPStatus, &conditionsJSON, &errors, &r.Success)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan a result from the database: %w", err)
		}

		r.Timestamp = time.Unix(ts/int64(time.Second), ts%int64(time.Second))
		r.Duration = time.Duration(dur)

		r.ConditionResults = make([]*StoredConditionResult, len(conditionsJSON.Elements))
		for i, j := range conditionsJSON.Elements {
			cr := StoredConditionResult{}
			err := json.Unmarshal([]byte(j.String), &cr)
			if err != nil {
				return nil, fmt.Errorf("Failed to unmarshal condition result: %w", err)
			}
			r.ConditionResults[i] = &cr
		}

		r.Errors = make([]string, len(errors.Elements))
		for _, e := range errors.Elements {
			r.Errors = append(r.Errors, e.String)
		}

		results = append(results, &r)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Cannot read results from the database: %w", err)
	}

	return results, nil
}

// ServicesTable returns the name of the services table
func (s *Store) ServicesTable() string {
	return s.namespace + "__services"
}

// ResultsTable returns the name of the results table
func (s *Store) ResultsTable() string {
	return s.namespace + "__results"
}

func (s *Store) mustCreateTables() error {

	createServices := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			key           TEXT PRIMARY KEY NOT NULL,
			name          TEXT NOT NULL,
			group_name    TEXT NOT NULL,
			events_json   TEXT[],
			uptime_json   TEXT NOT NULL,

			UNIQUE(group_name, name)
		);`,
		s.ServicesTable(),
	)

	createResults := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id              SERIAL PRIMARY KEY NOT NULL,
			service_key     TEXT,
			hostname        TEXT NOT NULL,
			timestamp_ns    BIGINT,
			duration_ns     BIGINT,
			http_status     INTEGER,
			conditions_json TEXT[],
			errors          TEXT[],
			success         BOOLEAN,

			UNIQUE(service_key, timestamp_ns),
			CONSTRAINT fk_service
				FOREIGN KEY(service_key) 
				REFERENCES %s(key)
				ON DELETE CASCADE
		);`,
		s.ResultsTable(),
		s.ServicesTable(),
	)

	createResultsIndex := fmt.Sprintf(`CREATE INDEX ON %s (service_key)`, s.ResultsTable())

	_, err := s.db.Exec(createServices)
	if err != nil {
		return fmt.Errorf("Cannot create services table: %w", err)
	}

	_, err = s.db.Exec(createResults)
	if err != nil {
		return fmt.Errorf("Cannot create results table: %w", err)
	}

	_, err = s.db.Exec(createResultsIndex)
	if err != nil {
		return fmt.Errorf("Cannot create results index: %w", err)
	}

	return nil
}
