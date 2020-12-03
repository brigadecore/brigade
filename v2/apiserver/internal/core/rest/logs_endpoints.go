package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// LogsEndpoints implements restmachinery.Endpoints to provide log-related URL
// --> action mappings to a restmachinery.Server.
type LogsEndpoints struct {
	AuthFilter restmachinery.Filter
	Service    core.LogsService
}

// Register is invoked by restmachinery.Server to register log-related URL
// --> action mappings to a restmachinery.Server.
func (l *LogsEndpoints) Register(router *mux.Router) {
	// Stream logs
	router.HandleFunc(
		"/v2/events/{id}/logs",
		l.AuthFilter.Decorate(l.stream),
	).Methods(http.MethodGet)
}

func (l *LogsEndpoints) stream(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	// nolint: errcheck
	follow, _ := strconv.ParseBool(r.URL.Query().Get("follow"))

	selector := core.LogsSelector{
		Job:       r.URL.Query().Get("job"),
		Container: r.URL.Query().Get("container"),
	}
	opts := core.LogStreamOptions{
		Follow: follow,
	}

	logEntryCh, err := l.Service.Stream(r.Context(), id, selector, opts)
	if err != nil {
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			restmachinery.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving log stream for event %q", id),
		)
		restmachinery.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			&meta.ErrInternalServer{},
		)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.(http.Flusher).Flush()
	for logEntry := range logEntryCh {
		logEntryBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling log entry"))
			return
		}
		fmt.Fprint(w, string(logEntryBytes))
		w.(http.Flusher).Flush()
	}
}
