package db_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/skydb/sky/db"
	"github.com/stretchr/testify/assert"
)

// Ensure that the database can open and set its path correctly.
func TestDBOpen(t *testing.T) {
	withDB(func(db *DB, path string) {
		assert.Equal(t, path, db.Path())
	})
}

// Ensure that the database returns an error if opened while already open.
func TestDBOpenAlreadyOpen(t *testing.T) {
	withDB(func(db *DB, path string) {
		assert.Equal(t, db.Open("/foo/bar"), ErrDatabaseOpen)
	})
}

// Ensure that the database can create a table.
func TestDBCreateTable(t *testing.T) {
	withDB(func(db *DB, path string) {
		table, err := db.CreateTable("foo", 0)
		if assert.NoError(t, err) {
			assert.NotNil(t, table)
			assert.Equal(t, table.Name(), "foo")
			assert.Equal(t, filepath.Base(table.Path()), "foo")
		}
	})
}

// Ensure that creating a table that already exists returns an error.
func TestDBCreateTableAlreadyExists(t *testing.T) {
	withDB(func(db *DB, path string) {
		db.CreateTable("foo", 0)
		table, err := db.CreateTable("foo", 0)
		assert.Equal(t, err, ErrTableExists)
		assert.Nil(t, table)
	})
}

// Ensure that creating a table while the database is closed returns an error.
func TestDBCreateTableWhileClosed(t *testing.T) {
	var db DB
	table, err := db.CreateTable("foo", 0)
	assert.Equal(t, err, ErrDatabaseNotOpen)
	assert.Nil(t, table)
}

// Ensure that opening a table returns a cached reference.
func TestDBOpenTable(t *testing.T) {
	withDB(func(db *DB, path string) {
		t0, _ := db.CreateTable("foo", 0)

		// Opening it should return the reference from create.
		t1, err := db.OpenTable("foo")
		if assert.NoError(t, err) {
			assert.True(t, t0 == t1)
		}

		// Open it again just for kicks.
		t2, err := db.OpenTable("foo")
		if assert.NoError(t, err) {
			assert.True(t, t0 == t2)
		}
	})
}

// Ensure that opening an unnamed table returned an error.
func TestDBOpenTableNameRequired(t *testing.T) {
	withDB(func(db *DB, path string) {
		table, err := db.CreateTable("", 0)
		assert.Equal(t, err, ErrTableNameRequired)
		assert.Nil(t, table)
	})
}

// Ensure that a table can be dropped.
func TestDBDropTable(t *testing.T) {
	withDB(func(db *DB, path string) {
		db.CreateTable("foo", 0)
		err := db.DropTable("foo")
		assert.NoError(t, err)

		// Opening it should return "not found".
		table, err := db.OpenTable("foo")
		assert.Equal(t, err, ErrTableNotFound)
		assert.Nil(t, table)
	})
}

// Ensure that dropping a table while the database is closed returns an error.
func TestDBDropTableNotOpen(t *testing.T) {
	withDB(func(db *DB, path string) {
		db.CreateTable("foo", 0)
		db.Close()
		err := db.DropTable("foo")
		assert.Equal(t, err, ErrDatabaseNotOpen)
	})
}

// Ensure that dropping a non-existent table returns an error.
func TestDBDropTableNotFound(t *testing.T) {
	withDB(func(db *DB, path string) {
		err := db.DropTable("foo")
		assert.Equal(t, err, ErrTableNotFound)
	})
}

func withDB(f func(db *DB, path string)) {
	path, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(path)

	db := &DB{NoSync: true}
	if err := db.Open(path); err != nil {
		panic(err.Error())
	}
	defer db.Close()

	f(db, path)
}