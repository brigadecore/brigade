package restmachinery

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/brigadecore/brigade/v2/internal/file"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/cors"
)

// ServerConfig represents optional configuration for a REST API server.
type ServerConfig struct {
	Port        int
	TLSEnabled  bool
	TLSCertPath string
	TLSKeyPath  string
}

// Server is an interface for the REST API server.
type Server interface {
	// Run causes the API server to start serving HTTP requests. It will block
	// until an error occurs and will return that error.
	ListenAndServe(ctx context.Context) error
}

type server struct {
	config  ServerConfig
	handler http.Handler
}

// NewServer returns a new REST API server that serves the provided Endpoints.
func NewServer(endpoints []Endpoints, config *ServerConfig) Server {
	if config == nil {
		config = &ServerConfig{}
	}
	if config.Port == 0 {
		config.Port = 8080
	}

	router := mux.NewRouter()
	router.StrictSlash(true)

	for _, eps := range endpoints {
		eps.Register(router)
	}

	return &server{
		config: *config,
		handler: cors.New(
			cors.Options{
				AllowedMethods: []string{"DELETE", "GET", "POST", "PUT"},
			},
		).Handler(router),
	}
}

// ListenAndServe runs the REST API server until it is interrupted. This
// function always returns a non-nil error.
func (s *server) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.handler,
	}

	errCh := make(chan error)

	if s.config.TLSEnabled {
		if s.config.TLSCertPath == "" {
			return errors.New(
				"TLS was enabled, but no certificate path was specified",
			)
		}

		if s.config.TLSKeyPath == "" {
			return errors.New(
				"TLS was enabled, but no key path was specified",
			)
		}

		ok, err := file.Exists(s.config.TLSCertPath)
		if err != nil {
			return errors.Wrap(err, "error checking for existence of TLS cert")
		}
		if !ok {
			return errors.Errorf(
				"no TLS certificate found at path %s",
				s.config.TLSCertPath,
			)
		}

		if ok, err = file.Exists(s.config.TLSKeyPath); err != nil {
			return errors.Wrap(err, "error checking for existence of TLS key")
		}
		if !ok {
			return errors.Errorf("no TLS key found at path %s", s.config.TLSKeyPath)
		}

		log.Printf(
			"API server is listening with TLS enabled on 0.0.0.0:%d",
			s.config.Port,
		)

		go func() {
			err := srv.ListenAndServeTLS(s.config.TLSCertPath, s.config.TLSKeyPath)
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
		}()
	} else {
		log.Printf(
			"API server is listening without TLS on 0.0.0.0:%d",
			s.config.Port,
		)

		go func() {
			err := srv.ListenAndServe()
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
		}()
	}

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// Five second grace period on shutdown
		shutdownCtx, cancel :=
			context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx) // nolint: errcheck
		return ctx.Err()
	}
}
