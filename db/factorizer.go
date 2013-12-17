package db

import (
	"fmt"
	"github.com/skydb/sky/core"
	"github.com/szferi/gomdb"
	"os"
	"reflect"
	"strconv"
	"sync"
)

// Factorizer object manages the factorization and defactorization of values.
type Factorizer interface {
	Open() error
	Close()
	Factorize(tablespace string, id string, value string, createIfMissing bool) (uint64, error)
	Defactorize(tablespace string, id string, value uint64) (string, error)
	FactorizeEvent(event *core.Event, tablespace string, propertyFile *core.PropertyFile, createIfMissing bool) error
	DefactorizeEvent(event *core.Event, tablespace string, propertyFile *core.PropertyFile) error
}

type factorizer struct {
	sync.Mutex
	env        *mdb.Env
	path       string
	noSync     bool
	maxDBs     uint
	maxReaders uint
}

// NewFactorizer returns a new LMDB-backed Factorizer.
func NewFactorizer(path string, noSync bool, maxDBs uint, maxReaders uint) Factorizer {
	return &factorizer{
		path:       path,
		noSync:     noSync,
		maxDBs:     maxDBs,
		maxReaders: maxReaders,
	}
}

// Path is the location of the factors database on disk.
func (f *factorizer) Path() string {
	return f.path
}

// Open allocates a new LMDB environment.
func (f *factorizer) Open() error {
	f.Lock()
	defer f.Unlock()
	f.close()

	if err := os.MkdirAll(f.path, 0700); err != nil {
		return err
	}

	var err error
	if f.env, err = mdb.NewEnv(); err != nil {
		return fmt.Errorf("factor env error: %s", err)
	}

	// LMDB environment settings.
	if err := f.env.SetMaxDBs(mdb.DBI(f.maxDBs)); err != nil {
		f.close()
		return fmt.Errorf("factor maxdbs error: %s", err)
	} else if err := f.env.SetMaxReaders(f.maxReaders); err != nil {
		f.close()
		return fmt.Errorf("factor maxreaders error: %s", err)
	} else if err := f.env.SetMapSize(2 << 40); err != nil {
		f.close()
		return fmt.Errorf("factor map size error: %s", err)
	}

	// Create LMDB flagset.
	options := uint(0)
	options |= mdb.NOTLS
	if f.noSync {
		options |= mdb.NOSYNC
	}

	// Open the LMDB environment.
	if err := f.env.Open(f.path, options, 0664); err != nil {
		f.close()
		return fmt.Errorf("factor env open error: %s", err)
	}

	return nil
}

// Close releases all factor resources.
func (f *factorizer) Close() {
	f.Lock()
	defer f.Unlock()
	f.close()
}

func (f *factorizer) close() {
	if f.env != nil {
		f.env.Close()
		f.env = nil
	}
}

// Converts the defactorized value for a given id in a given table to its internal representation.
func (f *factorizer) Factorize(tablespace string, id string, value string, createIfMissing bool) (uint64, error) {
	// Blank is always zero.
	if value == "" {
		return 0, nil
	}

	// Otherwise find it in the database.
	data, exists, err := f.get(tablespace, f.key(id, value))
	if err != nil {
		return 0, err
	}
	// If key does exist then parse and return it.
	if exists {
		return strconv.ParseUint(string(data), 10, 64)
	}

	// Create a new factor if requested.
	if createIfMissing {
		return f.add(tablespace, id, value)
	}

	err = NewFactorNotFound(fmt.Sprintf("skyd: Factor not found: %v", f.key(id, value)))
	return 0, err
}

// Adds a new factor to the database if it doesn't exist.
func (f *factorizer) add(tablespace string, id string, value string) (uint64, error) {
	// Lock while adding a new value.
	f.Lock()
	defer f.Unlock()

	// Retry factorize within the context of the lock.
	sequence, err := f.Factorize(tablespace, id, value, false)
	if err == nil {
		return sequence, nil
	} else if _, ok := err.(*FactorNotFound); !ok {
		return 0, err
	}

	// Retrieve next id in sequence.
	sequence, err = f.inc(tablespace, id)
	if err != nil {
		return 0, err
	}

	// Save lookup and reverse lookup.
	if err = f.put(tablespace, f.key(id, value), strconv.FormatUint(sequence, 10)); err != nil {
		return 0, err
	}
	if err = f.put(tablespace, f.revkey(id, sequence), value); err != nil {
		return 0, err
	}

	return sequence, nil
}

// Converts the factorized value for a given id in a given table to its internal representation.
func (f *factorizer) Defactorize(tablespace string, id string, value uint64) (string, error) {
	// Blank is always zero.
	if value == 0 {
		return "", nil
	}

	// Find it in the database.
	data, exists, err := f.get(tablespace, f.revkey(id, value))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("skyd: Factor value does not exist: %v", f.revkey(id, value))
	}
	return string(data), nil
}

// Retrieves the next available sequence number within a table for an id.
func (f *factorizer) inc(tablespace string, id string) (uint64, error) {
	data, exists, err := f.get(tablespace, f.seqkey(id))
	if err != nil {
		return 0, err
	}

	// Initialize key if it doesn't exist. Otherwise increment it.
	if !exists {
		if err := f.put(tablespace, f.seqkey(id), "1"); err != nil {
			return 0, err
		}
		return 1, nil
	}

	// Parse existing sequence.
	sequence, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("skyd: Unable to parse factor sequence: %v", data)
	}

	// Increment and save the new value.
	sequence += 1
	if err = f.put(tablespace, f.seqkey(id), strconv.FormatUint(sequence, 10)); err != nil {
		return 0, err
	}
	return sequence, nil
}

// Factorizes the values in an event.
func (f *factorizer) FactorizeEvent(event *core.Event, tablespace string, propertyFile *core.PropertyFile, createIfMissing bool) error {
	if event == nil {
		return nil
	}

	for k, v := range event.Data {
		property := propertyFile.GetProperty(k)
		if property.DataType == core.FactorDataType {
			if stringValue, ok := v.(string); ok {
				sequence, err := f.Factorize(tablespace, property.Name, stringValue, createIfMissing)
				if err != nil {
					return err
				}
				event.Data[k] = sequence
			}
		}
	}

	return nil
}

// Defactorizes the values in an event.
func (f *factorizer) DefactorizeEvent(event *core.Event, tablespace string, propertyFile *core.PropertyFile) error {
	if event == nil {
		return nil
	}

	for k, v := range event.Data {
		property := propertyFile.GetProperty(k)
		if property.DataType == core.FactorDataType {
			if sequence, ok := castUint64(v); ok {
				stringValue, err := f.Defactorize(tablespace, property.Name, sequence)
				if err != nil {
					return err
				}
				event.Data[k] = stringValue
			}
		}
	}

	return nil
}

// get retrieves the value from the database for a given key.
func (f *factorizer) get(tablespace string, key string) (string, bool, error) {
	txn, err := f.env.BeginTxn(nil, 0)
	if err != nil {
		return "", false, fmt.Errorf("skyd: Unable to start factors get txn: %s", err)
	}
	dbi, err := txn.DBIOpen(&tablespace, mdb.CREATE)
	if err != nil {
		return "", false, fmt.Errorf("skyd: Unable to open factors DBI [get]: %s", err)
	}

	// Retrieve byte array.
	data, err := txn.Get(dbi, []byte(key))
	if err != nil && err != mdb.NotFound {
		err = fmt.Errorf("skyd: Unable to get factor: %s", err)
		fmt.Fprintln(os.Stderr, err.Error())
		txn.Abort()
		return "", false, err
	}
	txn.Abort()

	return string(data), (data != nil), nil
}

// Sets the value for a given key in the database.
func (f *factorizer) put(tablespace string, key string, value string) error {
	txn, err := f.env.BeginTxn(nil, 0)
	if err != nil {
		return fmt.Errorf("skyd: Unable to start factors put txn: %s", err)
	}
	dbi, err := txn.DBIOpen(&tablespace, mdb.CREATE)
	if err != nil {
		return fmt.Errorf("skyd: Unable to open factors DBI [put]: %s", err)
	}

	// Set value for key.
	if err = txn.Put(dbi, []byte(key), []byte(value), mdb.NODUPDATA); err != nil {
		err = fmt.Errorf("skyd: Unable to put factor: %s", err)
		fmt.Fprintln(os.Stderr, err.Error())
		txn.Abort()
		return err
	}
	if err = txn.Commit(); err != nil {
		err = fmt.Errorf("skyd: Unable to commit factor: %s", err)
		fmt.Fprintln(os.Stderr, err.Error())
		txn.Abort()
		return err
	}

	return nil
}

// The key for a given id/value.
func (f *factorizer) key(id string, value string) string {
	return fmt.Sprintf("%x:%s>%s", len(id), id, value)
}

// The reverse key for a given id/value.
func (f *factorizer) revkey(id string, value uint64) string {
	return fmt.Sprintf("%x:%s<%d", len(id), id, value)
}

// The sequence key for a given id.
func (f *factorizer) seqkey(id string) string {
	return fmt.Sprintf("%x:%s!", len(id), id)
}

// Casts to a uint64 if possible.
func castUint64(value interface{}) (uint64, bool) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return uint64(v.Float()), true
	}
	return 0, false
}
