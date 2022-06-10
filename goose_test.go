package goose

import (
	"testing"
)

func setupTestDB() {
	if err := InitDatabase("./db/test.db"); err != nil {
		panic(err)
	}
}

func removeTestDB() {
	db = nil
	// if err := os.Remove("../db/test.db"); err != nil {
	// 	panic(err)
	// }
}

func TestDB(t *testing.T) {
	setupTestDB()
	defer removeTestDB()
}
