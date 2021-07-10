package postgres

import (
	"database/sql"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/TwinProduction/gatus/util"
	_ "github.com/jackc/pgx/v4/stdlib" // for sql.Open("pgx", ...)
)

type testCase struct {
	testName string
	status   *StoredServiceStatus
	wantErr  bool
}

func TestStore_ServiceStatusRoundTrip(t *testing.T) {

	db := newDB(t)
	table := "TestStore_ServiceStatusRoundTrip"

	store, err := NewStore(db, table)
	if err != nil {
		t.Fatalf("Cannot create test store: %v", err)
	}

	cases := []testCase{
		{
			testName: "can store simple data",
			status: &StoredServiceStatus{
				Name:      "example.com",
				GroupName: "testing",
				Events:    []*StoredEvent{},
				Uptime:    &StoredUptime{},
				Results: []*StoredResult{
					{
						Hostname:         "example.com",
						Timestamp:        time.Unix(9000, 1),
						Duration:         10 * time.Millisecond,
						HTTPStatus:       200,
						ConditionResults: []*StoredConditionResult{},
						Errors:           []string{},
						Success:          true,
					},
				},
			},
		},
	}

	for _, c := range cases {
		err = store.InsertServiceStatus(c.status)
		if c.wantErr != (err != nil) {
			t.Errorf("[%s] InsertServiceStatus got err %v expected error: %v", c.testName, err, c.wantErr)
		}

		key := util.ConvertGroupAndServiceToKey(c.status.GroupName, c.status.Name)
		gotStatus, err := store.FindServiceStatus(key)
		if c.wantErr != (err != nil) {
			t.Errorf("[%s] FindServiceStatus got err %v expected error: %v", c.testName, err, c.wantErr)
		}

		wantStatus := c.status // this is a round trip test
		for i := range gotStatus.Results {
			// we don't know the ids before and don't care
			wantStatus.Results[i].ID = gotStatus.Results[i].ID
		}
		if !reflect.DeepEqual(gotStatus, wantStatus) {
			t.Errorf("[%s] got status %s expected %s", c.testName, gotStatus, wantStatus)
		}
	}

	// clean up any litter we created in the database
	dropTables(store, t)
}

func newDB(t *testing.T) *sql.DB {

	dbEnvVar := "GATUS_TEST_DATABASE"

	// This test requires a real database, because a lot of the complexity is in the database driver transforming things.
	// You can start one on the side using docker:
	// ❯ docker run -d --name gatus-test-db -p5432:5432 -e POSTGRES_PASSWORD=hunter2 postgres
	// ❯ export GATUS_TEST_DATABASE=postgres://postgres:hunter2@localhost:5432/postgres

	connString := os.Getenv(dbEnvVar)
	if connString == "" {
		t.Fatalf("No test database configured (please export '%s')", dbEnvVar)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Cannot open test database connection: %v", err)
	}

	return db
}

func dropTables(s *Store, t *testing.T) {
	_, err := s.db.Exec("DROP TABLE " + s.ResultsTable())
	if err != nil {
		t.Fatalf("Cannot clear results table: %v", err)
	}
	_, err = s.db.Exec("DROP TABLE " + s.ServicesTable())
	if err != nil {
		t.Fatalf("Cannot clear service table: %v", err)
	}
}
