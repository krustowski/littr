package stats

import (
	"net/http"
	"testing"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/config"

	chi "github.com/go-chi/chi/v5"
)

var getStatsMock = func(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "stats")

	pl := struct{}{}

	l.Msg("ok, returning the application and users stats").Status(http.StatusOK).Log().Payload(pl).Write(w)
}

func TestStatsRouter(t *testing.T) {
	r := chi.NewRouter()

	// For the Streamer configuration check pkg/backend/live/streamer.go
	r.Get("/api/v1/stats", getStatsMock)

	// Fetch test net listener and test HTTP server configuration.
	listener := config.PrepareTestListener(t)
	defer listener.Close()

	ts := config.PrepareTestServer(t, listener, r)
	ts.Start()
	defer ts.Close()
}
