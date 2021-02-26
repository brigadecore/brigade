package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// HealthcheckEndpoints implements restmachinery.Endpoints to provide
// Healthcheck-related URL --> action mappings to a restmachinery.Server.
type HealthcheckEndpoints struct {
	APIServerVersion string
	DatabaseClient   *mongo.Client
	WriterFactory    queue.WriterFactory
}

// Register is invoked by restmachinery.Server to register Healthcheck-related
// URL --> action mappings to a restmachinery.Server.
func (h *HealthcheckEndpoints) Register(router *mux.Router) {
	// Get /healthz
	router.HandleFunc(
		"/healthz",
		h.healthz,
	).Methods(http.MethodGet)

	// Get /ping
	router.HandleFunc(
		"/ping",
		h.ping,
	).Methods(http.MethodGet)
}

// healthz is the main healthcheck endpoint for the api server.  The endpoint
// logic verifies connectivity to crucial dependencies, e.g. the database
// and messaging queues.
func (h *HealthcheckEndpoints) healthz(w http.ResponseWriter, r *http.Request) {
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
					&queue.MessageOptions{
						Ephemeral: true,
					},
				); err != nil {
					return false, errors.New("error writing to messaging queue")
				}

				return true, nil
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// ping returns the api server version and http.StatusOK.  This is handy for
// auxiliary components to verify their connectivity.
func (h *HealthcheckEndpoints) ping(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return fmt.Sprintf(`{"version": "%s"}`, h.APIServerVersion), nil
			},
			SuccessCode: http.StatusOK,
		},
	)
}
