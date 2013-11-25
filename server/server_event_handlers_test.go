package server

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

// Ensure that we can put an event on the server.
// func TestServerUpdateEvents(t *testing.T) {
// 	runTestServer(func(s *Server) {
// 		setupTestTable("foo")
// 		setupTestProperty("foo", "bar", false, "factor")
// 		setupTestProperty("foo", "baz", true, "integer")

// 		// Send two new events.
// 		resp, _ := sendTestHttpRequest("PUT", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T02:00:00Z", "application/json", `{"data":{"bar":"myValue", "baz":12}}`)
// 		assertResponse(t, resp, 200, "", "PUT /tables/:name/objects/:objectId/events failed.")
// 		resp, _ = sendTestHttpRequest("PUT", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T03:00:00Z", "application/json", `{"data":{"bar":"myValue2"}}`)
// 		assertResponse(t, resp, 200, "", "PUT /tables/:name/objects/:objectId/events failed.")

// 		// Replace the first one.
// 		resp, _ = sendTestHttpRequest("PUT", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T02:00:00Z", "application/json", `{"data":{"bar":"myValue3", "baz":1000}}`)
// 		assertResponse(t, resp, 200, "", "PUT /tables/:name/objects/:objectId/events failed.")

// 		// Merge the second one.
// 		resp, _ = sendTestHttpRequest("PATCH", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T03:00:00Z", "application/json", `{"data":{"bar":"myValue2", "baz":20}}`)
// 		assertResponse(t, resp, 200, "", "PATCH /tables/:name/objects/:objectId/events failed.")

// 		// Check our work.
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue3","baz":1000},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2","baz":20},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")

// 		// Grab a single event.
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T03:00:00Z", "application/json", "")
// 		assertResponse(t, resp, 200, `{"data":{"bar":"myValue2","baz":20},"timestamp":"2012-01-01T03:00:00Z"}`+"\n", "GET /tables/:name/objects/:objectId/events/:timestamp failed.")
// 	})
// }

// // Ensure that we can delete all events for an object.
// func TestServerDeleteEvent(t *testing.T) {
// 	runTestServer(func(s *Server) {
// 		setupTestTable("foo")
// 		setupTestProperty("foo", "bar", false, "string")

// 		// Send two new events.
// 		resp, _ := sendTestHttpRequest("PUT", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T02:00:00Z", "application/json", `{"data":{"bar":"myValue"}}`)
// 		assertResponse(t, resp, 200, "", "PUT /tables/:name/objects/:objectId/events failed.")
// 		resp, _ = sendTestHttpRequest("PUT", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T03:00:00Z", "application/json", `{"data":{"bar":"myValue2"}}`)
// 		assertResponse(t, resp, 200, "", "PUT /tables/:name/objects/:objectId/events failed.")

// 		// Delete one of the events.
// 		resp, _ = sendTestHttpRequest("DELETE", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T02:00:00Z", "application/json", "")
// 		assertResponse(t, resp, 200, "", "DELETE /tables/:name/objects/:objectId/events failed.")

// 		// Check our work.
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")
// 	})
// }

// // Ensure that we can delete all events for an object.
// func TestServerDeleteEvents(t *testing.T) {
// 	runTestServer(func(s *Server) {
// 		setupTestTable("foo")
// 		setupTestProperty("foo", "bar", false, "string")

// 		// Send two new events.
// 		resp, _ := sendTestHttpRequest("PUT", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T02:00:00Z", "application/json", `{"data":{"bar":"myValue"}}`)
// 		assertResponse(t, resp, 200, "", "PUT /tables/:name/objects/:objectId/events failed.")
// 		resp, _ = sendTestHttpRequest("PUT", "http://localhost:8586/tables/foo/objects/xyz/events/2012-01-01T03:00:00Z", "application/json", `{"data":{"bar":"myValue2"}}`)
// 		assertResponse(t, resp, 200, "", "PUT /tables/:name/objects/:objectId/events failed.")

// 		// Delete the events.
// 		resp, _ = sendTestHttpRequest("DELETE", "http://localhost:8586/tables/foo/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, "", "DELETE /tables/:name/objects/:objectId/events failed.")

// 		// Check our work.
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, "[]\n", "GET /tables/:name/objects/:objectId/events failed.")
// 	})
// }

// // Ensure that we can put multiple events on the server at once.
// func TestServerStreamUpdateEvents(t *testing.T) {
// 	runTestServer(func(s *Server) {
// 		setupTestTable("foo")
// 		setupTestProperty("foo", "bar", false, "string")
// 		setupTestProperty("foo", "baz", true, "integer")

// 		// Send two new events in one request.
// 		resp, _ := sendTestHttpRequest("PATCH", "http://localhost:8586/tables/foo/events", "application/json", `{"id":"xyz","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}`)
// 		assertResponse(t, resp, 200, `{"events_written":2}`, "PATCH /tables/:name/events failed.")

// 		// Check our work.
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")
// 	})
// }

// // Ensure that we can put multiple events on the server at once, using table agnostic event stream.
// func TestServerStreamUpdateEventsTableAgnostic(t *testing.T) {
// 	runTestServer(func(s *Server) {
// 		setupTestTable("foo_1")
// 		setupTestProperty("foo_1", "bar", false, "string")
// 		setupTestProperty("foo_1", "baz", true, "integer")

// 		setupTestTable("foo_2")
// 		setupTestProperty("foo_2", "bar", false, "string")
// 		setupTestProperty("foo_2", "baz", true, "integer")

// 		// Send two new events in one request.
// 		resp, _ := sendTestHttpRequest("PATCH", "http://localhost:8586/events", "application/json", `{"id":"xyz","table":"foo_1","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","table":"foo_2","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","table":"foo_1","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}{"id":"xyz","table":"foo_2","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}`)
// 		assertResponse(t, resp, 200, `{"events_written":4}`, "PATCH /events failed.")

// 		// Check our work.
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo_1/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo_2/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")
// 	})
// }

// Ensure that streamed events are flushed when count exceeds specified threshold.
func TestServerStreamUpdateEventsFlushesOnThreshold(t *testing.T) {
	runTestServer(func(s *Server) {
		setupTestTable("foo")
		setupTestProperty("foo", "bar", false, "string")
		setupTestProperty("foo", "baz", true, "integer")
		s.StreamFlushThreshold = 2

		method := "PATCH"
		endpoint := "http://localhost:8586/events"
		contentType := "application/json"

		client := &http.Client{Transport: &http.Transport{DisableKeepAlives: false}}
		in, out := io.Pipe()
		// go io.Copy(reqIn, in)
		req, err := http.NewRequest(method, endpoint, in)
		assert.NoError(t, err)
		req.Header.Add("Content-Type", contentType)

		finished := make(chan interface{})
		go func() {
			resp, err := client.Do(req)
			assert.NoError(t, err)
			finished <- resp
		}()

		// Send two new events in one request.
		body := `{"id":"xyz","table":"foo","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}` + "\n"
		_, err = io.WriteString(io.Writer(out), body)
		assert.NoError(t, err)

		// Assert that the event was NOT flushed
		resp, err := sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", contentType, "")
		assert.NoError(t, err)
		assertResponse(t, resp, 200, `[]`+"\n", "GET /events failed.")

		// resp, _ := sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", contentType, "")
		// assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")

		// resp, _ := sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", contentType, "")
		// assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")

		// in.Close()

		//,{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}
		// Send another event and ensure flushed
		// time.Sleep(10 * time.Second)

		body = `{"id":"xyz","table":"foo","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}` + "\n"
		_, err = io.WriteString(io.Writer(out), body)
		assert.NoError(t, err)

		time.Sleep(1 * time.Second)

		// Assert that the events were flushed
		resp, err = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", contentType, "")
		assert.NoError(t, err)
		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /events failed.")
		// resp, err = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", contentType, "")
		// assert.NoError(t, err)
		// assertResponse(t, resp, 200, `[{"id":"xyz","table":"foo","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","table":"foo","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}]`+"\n", "GET /events failed.")

		// resp, _ := sendTestStreamingHttpRequest(client, method, endpoint, contentType, body)
		// req, _ = http.NewRequest(method, endpoint, strings.NewReader(body))
		// req.Header.Add("Content-Type", contentType)
		// resp = client.Do(req)
		// TODO: assert flushed

		out.Close()
		ret := <-finished
		in.Close()

		resp = ret.(*http.Response)
		assertResponse(t, resp, 200, `{"events_written":2}`, "PATCH /events failed.")

		// Ensure events were flushed
		resp, err = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo/objects/xyz/events", contentType, "")
		assert.NoError(t, err)
		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /events failed.")

		// body = `{"id":"xyz","table":"foo","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}`

		// resp, _ := sendTestHttpRequest("PATCH", "http://localhost:8586/events", "application/json", `{"id":"xyz","table":"foo_1","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","table":"foo_2","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","table":"foo_1","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}{"id":"xyz","table":"foo_2","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}`)
		// assertResponse(t, resp, 200, `{"events_written":4}`, "PATCH /events failed.")

		// Check our work.
		// resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo_1/objects/xyz/events", "application/json", "")
		// assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")
	})
}

// // Ensure that we can put multiple events on the server at once, using table agnostic event stream.
// func TestServerStreamUpdateEventsFlushesOnPeriod(t *testing.T) {
// 	runTestServer(func(s *Server) {
// 		setupTestTable("foo_1")
// 		setupTestProperty("foo_1", "bar", false, "string")
// 		setupTestProperty("foo_1", "baz", true, "integer")

// 		setupTestTable("foo_2")
// 		setupTestProperty("foo_2", "bar", false, "string")
// 		setupTestProperty("foo_2", "baz", true, "integer")

// 		// Send two new events in one request.
// 		resp, _ := sendTestHttpRequest("PATCH", "http://localhost:8586/events", "application/json", `{"id":"xyz","table":"foo_1","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","table":"foo_2","timestamp":"2012-01-01T02:00:00Z","data":{"bar":"myValue", "baz":12}}{"id":"xyz","table":"foo_1","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}{"id":"xyz","table":"foo_2","timestamp":"2012-01-01T03:00:00Z","data":{"bar":"myValue2"}}`)
// 		assertResponse(t, resp, 200, `{"events_written":4}`, "PATCH /events failed.")

// 		// Check our work.
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo_1/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")
// 		resp, _ = sendTestHttpRequest("GET", "http://localhost:8586/tables/foo_2/objects/xyz/events", "application/json", "")
// 		assertResponse(t, resp, 200, `[{"data":{"bar":"myValue","baz":12},"timestamp":"2012-01-01T02:00:00Z"},{"data":{"bar":"myValue2"},"timestamp":"2012-01-01T03:00:00Z"}]`+"\n", "GET /tables/:name/objects/:objectId/events failed.")
// 	})
// }
