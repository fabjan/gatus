package postgres

import (
	"testing"
)

// TODO Share the test data? (firstCondition, testService, etc.) It's the same
//      in memory_test and store_bench_test, and we can use the same here.

func TestStore_Insert(t *testing.T) {
	// TODO inject a mock db (https://github.com/DATA-DOG/go-sqlmock)
	store, _ := NewStore("TestStore_Insert")

	t.Errorf("%v does not implement anything yet!", store)
}
