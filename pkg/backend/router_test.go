package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"go.vxn.dev/littr/pkg/backend/auth"
	"go.vxn.dev/littr/pkg/backend/common"

	//"go.vxn.dev/littr/pkg/backend/db"
	//"go.vxn.dev/littr/pkg/backend/live"
	//"go.vxn.dev/littr/pkg/backend/polls"
	//"go.vxn.dev/littr/pkg/backend/posts"
	//"go.vxn.dev/littr/pkg/backend/push"
	//"go.vxn.dev/littr/pkg/backend/stats"
	//"go.vxn.dev/littr/pkg/backend/users"
	"go.vxn.dev/littr/pkg/config"
)

const routePrefix = "/api/v1"

func TestAPIRouter(t *testing.T) {
	dummy := func(w http.ResponseWriter, r *http.Request) {
		//w.WriteHeader(200)
		//w.Write([]byte(fmt.Sprintf("dummy")))
		common.NewLogger(r, "test").Status(http.StatusOK).Payload(nil).Write(w)
	}

	root := func(w http.ResponseWriter, r *http.Request) {
		body := common.APIResponse{
			Message:   "littr JSON API service (v" + os.Getenv("APP_VERSION") + ")",
			Timestamp: time.Now().Unix(),
		}

		if os.Getenv("APP_VERSION") == "" {
			t.Errorf("APP_VERSION env var is empty")
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Error(err)
		}

		w.WriteHeader(200)
		if _, err := w.Write(jsonBody); err != nil {
			t.Error(err)
		}
	}

	r := chi.NewRouter()

	r.Use(auth.AuthMiddleware)

	// Rate limiter (see limiter in pkg/backend/router.go).
	if !config.IsLimiterDisabled {
		r.Use(limiter)
	}

	r.Get(routePrefix, root)

	// Auth bypass routes
	r.Post(routePrefix+"/users", dummy)
	for _, path := range auth.PathExceptions {
		if path == routePrefix {
			continue
		}

		r.Get(path, dummy)
	}

	//
	//  Configure and start the listener and server
	//

	// Create a custom network TCP connection listener.
	listener := config.PrepareTestListener(t)
	defer func() {
		if err := listener.Close(); err != nil {
			t.Error(err)
		}
	}()

	// Create a custom HTTP server configuration for the test server for SSE.
	ts := config.PrepareTestServer(t, listener, r)

	// Start the HTTP server.
	ts.Start()
	defer ts.Close()

	//
	//  Basic route tests
	//

	if resp, body := testRequest(t, ts, http.MethodGet, "/afdshfajshfafd", nil); resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Response: %d", resp.StatusCode)
		t.Error(body)
	}

	if resp, body := testRequest(t, ts, http.MethodGet, "/afdshfajshfafd", nil); resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Response: %d", resp.StatusCode)
		t.Error(body)
	}

	if resp, body := testRequest(t, ts, http.MethodGet, routePrefix, nil); resp.StatusCode == http.StatusOK {
		var data common.APIResponse

		err := json.Unmarshal([]byte(body), &data)
		if err != nil {
			t.Error(err.Error())
		}

		if data.Message != "littr JSON API service (v"+os.Getenv("APP_VERSION")+")" {
			t.Errorf("invalid response message")
			t.Error(body)
		}

		if data.Timestamp == 0 {
			t.Errorf("timestamp is zero")
			t.Error(body)
		}
	}

	if resp, body := testRequest(t, ts, http.MethodGet, routePrefix, nil); resp.StatusCode != http.StatusOK {
		t.Errorf("Response: %d", resp.StatusCode)
		t.Error(body)
	}

	//
	//  Auth bypass route tests
	//

	for _, path := range auth.PathExceptions {
		if resp, body := testRequest(t, ts, http.MethodGet, path, nil); resp.StatusCode != http.StatusOK {
			t.Errorf("Response: %d", resp.StatusCode)
			t.Error(body)
		}
	}

	//
	//  Limiter test
	//

	if !config.IsLimiterDisabled {
		for i := 0; i < config.LIMITER_REQS_NUM; i++ {
			_, _ = testRequest(t, ts, http.MethodGet, routePrefix, nil)
		}

		if resp, body := testRequest(t, ts, http.MethodGet, routePrefix, nil); resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Response: %d", resp.StatusCode)
			t.Error(body)
		}
	}
}

// https://github.com/go-chi/chi/blob/master/mux_test.go
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Error(err)
		}
	}()

	return resp, string(respBody)
}
