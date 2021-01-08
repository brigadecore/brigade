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

	// Clients can request use of the SSE protocol instead of HTTP/2 streaming.
	// Not every potential client language has equally good support for both of
	// those, so allowing clients to pick is useful.
	sse, _ := strconv.ParseBool(r.URL.Query().Get("sse")) // nolint: errcheck

	selector := core.LogsSelector{
		Job:       r.URL.Query().Get("job"),
		Container: r.URL.Query().Get("container"),
	}
	opts := core.LogStreamOptions{
		Follow: follow,
	}

	var lastEventID int64
	// SSE has support for resuming where you left off after a
	// disconnect/reconnect. Formally, I (krancour) believe the specification says
	// clients can echo the ID of the last message received in a Last-Event-ID
	// header, but I've encountered clients that use a lastEventId query parameter
	// instead. Invoking Postel's principle, we'll support both.
	if sse {
		lastEventIDStr := r.Header.Get("Last-Event-ID")
		if lastEventIDStr == "" {
			lastEventIDStr = r.URL.Query().Get("lastEventId")
		}
		if lastEventIDStr != "" {
			var err error
			lastEventID, err = strconv.ParseInt(lastEventIDStr, 10, 64)
			if err != nil {
				restmachinery.WriteAPIResponse(
					w,
					http.StatusBadRequest,
					&meta.ErrBadRequest{
						Reason: "Value of last id was not parseable as an int",
					},
				)
				return
			}
		}
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
	var i int64
	for logEntry := range logEntryCh {
		logEntryBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling log entry"))
			continue
		}
		if sse {
			i++
			// SSE has support for resuming where you left off after a
			// disconnect/reconnect. This check is so we can "fast forward" past
			// messages the client says it already received.
			if i > lastEventID {
				fmt.Fprintf(
					w,
					"event: message\ndata: %s\nid: %d\n\n",
					string(logEntryBytes),
					i,
				)
				w.(http.Flusher).Flush()
			}
		} else {
			fmt.Fprint(w, string(logEntryBytes))
			w.(http.Flusher).Flush()
		}
	}
	// If we're using SSE, we'll explicitly send an event that denotes the end of
	// the stream.
	if sse {
		i++
		fmt.Fprintf(w, "event: done\ndata: done\nid: %d\n\n", i)
		w.(http.Flusher).Flush()
	}
}
