package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// SystemEndpoints implements restmachinery.Endpoints to provide
// System-related URL --> action mappings to a restmachinery.Server.
type SystemEndpoints struct {
	APIServerVersion string
	DatabaseClient   *mongo.Client
	WriterFactory    queue.WriterFactory
}

// Register is invoked by restmachinery.Server to register System-related
// URL --> action mappings to a restmachinery.Server.
func (h *SystemEndpoints) Register(router *mux.Router) {
	// Get /healthz
	router.HandleFunc(
		"/healthz",
		h.healthz,
	).Methods(http.MethodGet)

	// Get /v2/ping
	router.HandleFunc(
		"/v2/ping",
		h.ping,
	).Methods(http.MethodGet)

	// Get /ping (unversioned)
	router.HandleFunc(
		"/ping",
		h.unversionedPing,
	).Methods(http.MethodGet)
}

// healthz is the main healthcheck endpoint for the api server.  The endpoint
// logic verifies connectivity to crucial dependencies, e.g. the database
// and messaging queues.
func (h *SystemEndpoints) healthz(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// Check Database
				err := h.DatabaseClient.Ping(ctx, nil)
				if err != nil {
					return false, errors.New("error communicating with database")
				}

				// Check messaging queue
				writer, err := h.WriterFactory.NewWriter("healthz")
				if err != nil {
					return false, errors.New("error instantiating messaging queue writer")
				}
				defer writer.Close(ctx)
				if err := writer.Write(
					ctx,
					"ping",
					&queue.MessageOptions{},
				); err != nil {
					return false, errors.New("error writing to messaging queue")
				}

				return []byte("ok"), nil
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// unversionedPing returns the api server version and http.StatusOK.
// This is handy for auxiliary components to verify their connectivity.
func (h *SystemEndpoints) unversionedPing(
	w http.ResponseWriter,
	r *http.Request,
) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return []byte(h.APIServerVersion), nil
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// ping returns the api server version and http.StatusOK.  This is handy for
// auxiliary components to verify their connectivity.
func (h *SystemEndpoints) ping(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return PingResponse{Version: h.APIServerVersion}, nil
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// PingResponse represents the response object returned by the ping endpoint
type PingResponse struct {
	Version string
}

// MarshalJSON amends PingResponse instances with type metadata.
func (p PingResponse) MarshalJSON() ([]byte, error) {
	type Alias PingResponse
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "PingResponse",
			},
			Alias: (Alias)(p),
		},
	)
}
