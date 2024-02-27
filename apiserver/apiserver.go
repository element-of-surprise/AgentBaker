package apiserver

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	agentoverrides "github.com/Azure/agentbaker/pkg/agent/overrides"
)

const (
	readHeaderTimeoutSeconds = 5
)

// Options holds the options for the api server.
type Options struct {
	Addr string
}

func (o *Options) validate() error {
	if o == nil {
		return errors.New("absvc options cannot be nil")
	}

	if o.Addr == "" {
		return errors.New("addr must not be empty")
	}
	return nil
}

// APIServer contains the connections details required to run the api.
type APIServer struct {
	Options          *Options
	ServiceOverrides *agentoverrides.Overrides
}

// NewAPIServer creates an APIServer object with defaults.
func NewAPIServer(serviceOverrides *agentoverrides.Overrides, o *Options) (*APIServer, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}

	s := &APIServer{
		Options:          o,
		ServiceOverrides: serviceOverrides,
	}

	return s, nil
}

// ListenAndServe wraps http.Server and provides context-based cancelation.
func (api *APIServer) ListenAndServe(ctx context.Context) error {
	svr := http.Server{
		Addr:              api.Options.Addr,
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: readHeaderTimeoutSeconds * time.Second,
	}

	errors := make(chan error)
	go func() {
		errors <- svr.ListenAndServe()
	}()

	log.Printf("Starting APIServer at %s\n", api.Options.Addr)
	select {
	case <-ctx.Done():
		return svr.Shutdown(context.Background())
	case err := <-errors:
		return err
	}
}
