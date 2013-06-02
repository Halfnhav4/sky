package skyd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/benbjohnson/gomdb"
	"github.com/ugorji/go/codec"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"time"
)

//------------------------------------------------------------------------------
//
// Typedefs
//
//------------------------------------------------------------------------------

// A Servlet is a small wrapper around a single shard of a LevelDB data file.
type Servlet struct {
	path    string
	env     *mdb.Env
	factors *Factors
	mutex   sync.Mutex
}

//------------------------------------------------------------------------------
//
// Constructors
//
//------------------------------------------------------------------------------

// NewServlet returns a new Servlet with a data shard stored at a given path.
func NewServlet(path string, factors *Factors) *Servlet {
	return &Servlet{
		path:    path,
		factors: factors,
	}
}

//------------------------------------------------------------------------------
//
// Methods
//
//------------------------------------------------------------------------------

//--------------------------------------
// Lifecycle
//--------------------------------------

// Opens the underlying LevelDB database and starts the message loop.
func (s *Servlet) Open() error {
	// Create directory if it doesn't exist.
	if err := os.MkdirAll(s.path, 0700); err != nil {
		return err
	}

	// Create the database environment.
	var err error
	if s.env, err = mdb.NewEnv(); err != nil {
		return fmt.Errorf(fmt.Sprintf("skyd.Servlet: Unable to create LMDB environment: %v", err))
	}
	// Setup max dbs.
	if err := s.env.SetMaxDBs(1024); err != nil {
		return fmt.Errorf("skyd.Servlet: Unable to set LMDB max dbs: %v", err)
	}
	// Setup map size.
	if err := s.env.SetMapSize(10 * (2 << 30)); err != nil {
		return fmt.Errorf("skyd.Servlet: Unable to set LMDB map size: %v", err)
	}
	// Open the database.
	err = s.env.Open(s.path, mdb.NOSYNC, 0664)
	if err != nil {
		return fmt.Errorf("skyd.Servlet: Cannot open servlet: %s", err)
	}

	return nil
}

// Closes the underlying database.
func (s *Servlet) Close() {
	if s.env != nil {
		s.env.Close()
	}
}

// Deletes a table stored on the servlet.
func (s *Servlet) DeleteTable(name string) error {
	s.Lock()
	defer s.Unlock()

	// Begin a transaction.
	txn, dbi, err := s.mdbTxnBegin(name, false)
	if err != nil {
		return fmt.Errorf("skyd.Servlet: Unable to begin LMDB transaction for table deletion: %s", err)
	}

	// Drop the table.
	if err = txn.Drop(dbi, 1); err != nil {
		txn.Abort()
		return fmt.Errorf("skyd.Servlet: Unable to drop LMDB DBI: %s", err)
	}

	// Commit the transaction.
	if err = txn.Commit(); err != nil {
		txn.Abort()
		return fmt.Errorf("skyd.Servlet: Unable to commit LMDB drop: %s", err)
	}

	return nil
}

//--------------------------------------
// Lock Management
//--------------------------------------

// Locks the entire servlet.
func (s *Servlet) Lock() {
	s.mutex.Lock()
}

// Unlocks the entire servlet.
func (s *Servlet) Unlock() {
	s.mutex.Unlock()
}

//--------------------------------------
// Event Management
//--------------------------------------

// Adds an event for a given object in a table to a servlet.
func (s *Servlet) PutEvent(table *Table, objectId string, event *Event, replace bool) error {
	s.Lock()
	defer s.Unlock()

	// Make sure the servlet is open.
	if s.env == nil {
		return fmt.Errorf("Servlet is not open: %v", s.path)
	}

	// Do not allow empty events to be added.
	if event == nil {
		return errors.New("skyd.PutEvent: Cannot add nil event")
	}

	// Check the current state and perform an optimized append if possible.
	state, data, err := s.GetState(table, objectId)
	if state == nil || state.Timestamp.Before(event.Timestamp) {
		return s.appendEvent(table, objectId, event, state, data)
	}

	// Retrieve the events and state for the object.
	tmp, state, err := s.GetEvents(table, objectId)
	if err != nil {
		return err
	}

	// Remove any event matching the timestamp.
	found := false
	state = &Event{Timestamp: event.Timestamp, Data: map[int64]interface{}{}}
	events := make([]*Event, 0)
	for _, v := range tmp {
		// Replace or merge with existing event.
		if v.Timestamp.Equal(event.Timestamp) {
			// Dedupe all permanent state.
			event.Dedupe(state)

			// Replace or merge.
			if replace {
				v = event
			} else {
				v.Merge(event)
			}
			found = true
		}
		events = append(events, v)

		// Keep track of permanent state.
		state.MergePermanent(v)
	}
	// Add the event if it wasn't found.
	if !found {
		event.Dedupe(state)
		events = append(events, event)
		state.MergePermanent(event)
	}

	// Write events back to the database.
	err = s.SetEvents(table, objectId, events, state)
	if err != nil {
		return err
	}

	return nil
}

// Appends an event for a given object in a table to a servlet. This should not
// be called directly but only through PutEvent().
func (s *Servlet) appendEvent(table *Table, objectId string, event *Event, state *Event, data []byte) error {
	if state == nil {
		state = &Event{Data: map[int64]interface{}{}}
	}
	state.Timestamp = event.Timestamp
	event.Dedupe(state)
	state.MergePermanent(event)

	// Append new event.
	buffer := bytes.NewBuffer(data)
	if err := event.EncodeRaw(buffer); err != nil {
		return err
	}

	// Write everything to the database.
	return s.SetRawEvents(table, objectId, buffer.Bytes(), state)
}

// Retrieves an event for a given object at a single point in time.
func (s *Servlet) GetEvent(table *Table, objectId string, timestamp time.Time) (*Event, error) {
	// Retrieve all events.
	events, _, err := s.GetEvents(table, objectId)
	if err != nil {
		return nil, err
	}

	// Find an event at a given point in time.
	for _, v := range events {
		if v.Timestamp.Equal(timestamp) {
			return v, nil
		}
	}

	return nil, nil
}

// Removes an event for a given object in a table to a servlet.
func (s *Servlet) DeleteEvent(table *Table, objectId string, timestamp time.Time) error {
	s.Lock()
	defer s.Unlock()

	// Make sure the servlet is open.
	if s.env == nil {
		return fmt.Errorf("Servlet is not open: %v", s.path)
	}

	// Retrieve the events for the object and append.
	tmp, _, err := s.GetEvents(table, objectId)
	if err != nil {
		return err
	}
	// Remove any event matching the timestamp.
	state := &Event{Data: map[int64]interface{}{}}
	events := make([]*Event, 0)
	for _, v := range tmp {
		if !v.Timestamp.Equal(timestamp) {
			events = append(events, v)
			state.MergePermanent(v)
		}
	}

	// Write events back to the database.
	err = s.SetEvents(table, objectId, events, state)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves the state and the remaining serialized event stream for an object.
func (s *Servlet) GetState(table *Table, objectId string) (*Event, []byte, error) {
	// Make sure the servlet is open.
	if s.env == nil {
		return nil, nil, fmt.Errorf("Servlet is not open: %v", s.path)
	}

	// Begin a transaction.
	txn, dbi, err := s.mdbTxnBegin(table.Name, false)
	if err != nil {
		return nil, nil, fmt.Errorf("skyd.Servlet: Unable to begin LMDB transaction for get: %s", err)
	}

	// Retrieve byte array.
	data, err := txn.Get(dbi, []byte(objectId))
	if err != nil && err != mdb.NotFound {
		txn.Abort()
		return nil, nil, err
	}

	// Commit the transaction.
	if err = txn.Commit(); err != nil {
		txn.Abort()
		return nil, nil, fmt.Errorf("skyd.Servlet: Unable to commit LMDB get: %s", err)
	}

	// Decode the events into a slice.
	if data != nil {
		reader := bytes.NewReader(data)

		// The first item should be the current state wrapped in a raw value.
		var raw interface{}
		var handle codec.MsgpackHandle
		handle.RawToString = true
		decoder := codec.NewDecoder(reader, &handle)
		if err := decoder.Decode(&raw); err != nil && err != io.EOF {
			return nil, nil, err
		}
		if b, ok := raw.(string); ok {
			state := &Event{}
			if err = state.DecodeRaw(bytes.NewReader([]byte(b))); err == nil {
				eventData, _ := ioutil.ReadAll(reader)
				return state, eventData, nil
			} else if err != io.EOF {
				return nil, nil, err
			}
		} else {
			return nil, nil, fmt.Errorf("skyd.Servlet: Invalid state: %v", raw)
		}
	}

	return nil, []byte{}, nil
}

// Retrieves a list of events and the current state for a given object in a table.
func (s *Servlet) GetEvents(table *Table, objectId string) ([]*Event, *Event, error) {
	state, data, err := s.GetState(table, objectId)
	if err != nil {
		return nil, nil, err
	}

	events := make([]*Event, 0)
	if data != nil {
		reader := bytes.NewReader(data)
		for {
			// Decode the event and append it to our list.
			event := &Event{}
			err = event.DecodeRaw(reader)
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				return nil, nil, err
			}
			events = append(events, event)
		}
	}

	return events, state, nil
}

// Writes a list of events for an object in table.
func (s *Servlet) SetEvents(table *Table, objectId string, events []*Event, state *Event) error {
	// Sort the events.
	sort.Sort(EventList(events))

	// Ensure state is correct before proceeding.
	if len(events) > 0 {
		if state != nil {
			state.Timestamp = events[len(events)-1].Timestamp
		} else {
			return errors.New("skyd.Servlet: Missing state.")
		}
	} else {
		state = nil
	}

	// Encode the events.
	buffer := new(bytes.Buffer)
	for _, event := range events {
		err := event.EncodeRaw(buffer)
		if err != nil {
			return err
		}
	}

	// Set the raw bytes.
	return s.SetRawEvents(table, objectId, buffer.Bytes(), state)
}

// Writes a list of events for an object in table.
func (s *Servlet) SetRawEvents(table *Table, objectId string, data []byte, state *Event) error {
	var err error

	// Make sure the servlet is open.
	if s.env == nil {
		return fmt.Errorf("Servlet is not open: %v", s.path)
	}

	// Encode the state at the beginning.
	buffer := new(bytes.Buffer)
	var b []byte
	if state != nil {
		if b, err = state.MarshalRaw(); err != nil {
			return err
		}
	} else {
		b = []byte{}
	}
	var handle codec.MsgpackHandle
	handle.RawToString = true
	if err := codec.NewEncoder(buffer, &handle).Encode(b); err != nil {
		return err
	}

	// Encode the rest of the data.
	buffer.Write(data)

	// Begin a transaction.
	txn, dbi, err := s.mdbTxnBegin(table.Name, false)
	if err != nil {
		return fmt.Errorf("skyd.Servlet: Unable to begin LMDB transaction to set raw: %s", err)
	}

	if err = txn.Put(dbi, []byte(objectId), buffer.Bytes(), mdb.NODUPDATA); err != nil {
		txn.Abort()
		return fmt.Errorf("skyd.Servlet: Unable to put LMDB value: %s", err)
	}

	// Commit the transaction.
	if err = txn.Commit(); err != nil {
		txn.Abort()
		return fmt.Errorf("skyd.Servlet: Unable to commit LMDB get: %s", err)
	}

	return nil
}

// Deletes all events for a given object in a table.
func (s *Servlet) DeleteEvents(table *Table, objectId string) error {
	// Make sure the servlet is open.
	if s.env == nil {
		return fmt.Errorf("Servlet is not open: %v", s.path)
	}

	// Begin a transaction.
	txn, dbi, err := s.mdbTxnBegin(table.Name, false)
	if err != nil {
		return fmt.Errorf("skyd.Servlet: Unable to begin LMDB transaction for deletion: %s", err)
	}

	// Delete the key.
	if err = txn.Del(dbi, []byte(objectId), nil); err != nil {
		txn.Abort()
		return fmt.Errorf("skyd.Servlet: Unable to delete LMDB key: %s", err)
	}

	// Commit the transaction.
	if err = txn.Commit(); err != nil {
		txn.Abort()
		return fmt.Errorf("skyd.Servlet: Unable to commit LMDB get: %s", err)
	}

	return nil
}

//--------------------------------------
// Execution Engine
//--------------------------------------

// Creates and initializes an execution engine for querying this servlet.
func (s *Servlet) CreateExecutionEngine(table *Table, prefix string, source string) (*ExecutionEngine, error) {
	e, err := NewExecutionEngine(table, prefix, source)
	if err != nil {
		return nil, err
	}

	// Begin a transaction.
	txn, dbi, err := s.mdbTxnBegin(table.Name, false)
	if err != nil {
		return nil, fmt.Errorf("skyd.Servlet: Unable to begin LMDB transaction for execution engine: %s", err)
	}

	// Setup cursor.
	cursor, err := txn.CursorOpen(dbi)
	if err != nil {
		e.Destroy()
		cursor.Close()
		txn.Abort()
		return nil, fmt.Errorf("skyd.Servlet: Unable to open LMDB cursor: %s", err)
	}

	// Initialize cursor.
	if err = e.SetLmdbCursor(cursor); err != nil {
		e.Destroy()
		cursor.Close()
		return nil, fmt.Errorf("skyd.Servlet: Unable to initialize LMDB cursor: %s", err)
	}

	return e, nil
}

//--------------------------------------
// LDMB
//--------------------------------------

// Creates and initializes an execution engine for querying this servlet.
func (s *Servlet) mdbTxnBegin(name string, readOnly bool) (*mdb.Txn, mdb.DBI, error) {
	var flags uint = 0
	if readOnly {
		flags = flags | mdb.RDONLY
	}

	// Setup cursor to iterate over table data.
	txn, err := s.env.BeginTxn(nil, flags)
	if err != nil {
		return nil, 0, fmt.Errorf("skyd.Servlet: Unable to start LMDB transaction: %s", err)
	}
	dbi, err := txn.DBIOpen(&name, mdb.CREATE)
	if err != nil {
		return nil, 0, fmt.Errorf("skyd.Servlet: Unable to open LMDB DBI: %s", err)
	}

	return txn, dbi, nil
}
