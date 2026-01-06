package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/podoru/podoru-chain/internal/node"
	"github.com/sirupsen/logrus"
)

// Server represents the REST API server
type Server struct {
	node       *node.Node
	router     *mux.Router
	httpServer *http.Server
	logger     *logrus.Logger
}

// NewServer creates a new REST API server
func NewServer(n *node.Node, bindAddr string, port int, logger *logrus.Logger) *Server {
	if logger == nil {
		logger = logrus.New()
	}

	server := &Server{
		node:   n,
		router: mux.NewRouter(),
		logger: logger,
	}

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", bindAddr, port)
	server.httpServer = &http.Server{
		Addr:         addr,
		Handler:      server.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// setupRoutes sets up all API routes
func (s *Server) setupRoutes() {
	// Chain endpoints
	s.router.HandleFunc("/api/v1/chain/info", s.handleGetChainInfo).Methods("GET")
	s.router.HandleFunc("/api/v1/block/{hash}", s.handleGetBlockByHash).Methods("GET")
	s.router.HandleFunc("/api/v1/block/height/{height}", s.handleGetBlockByHeight).Methods("GET")
	s.router.HandleFunc("/api/v1/block/latest", s.handleGetLatestBlock).Methods("GET")

	// Transaction endpoints
	s.router.HandleFunc("/api/v1/transaction/{hash}", s.handleGetTransaction).Methods("GET")
	s.router.HandleFunc("/api/v1/transaction", s.handleSubmitTransaction).Methods("POST")

	// State endpoints
	s.router.HandleFunc("/api/v1/state/{key}", s.handleGetState).Methods("GET")
	s.router.HandleFunc("/api/v1/state/batch", s.handleBatchGetState).Methods("POST")
	s.router.HandleFunc("/api/v1/state/query/prefix", s.handleQueryByPrefix).Methods("POST")

	// Node endpoints
	s.router.HandleFunc("/api/v1/node/info", s.handleGetNodeInfo).Methods("GET")
	s.router.HandleFunc("/api/v1/node/peers", s.handleGetPeers).Methods("GET")
	s.router.HandleFunc("/api/v1/node/health", s.handleHealthCheck).Methods("GET")

	// Mempool endpoints
	s.router.HandleFunc("/api/v1/mempool", s.handleGetMempool).Methods("GET")

	// Add logging middleware
	s.router.Use(s.loggingMiddleware)
}

// Start starts the API server
func (s *Server) Start() error {
	s.logger.Infof("Starting REST API server on %s", s.httpServer.Addr)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Errorf("REST API server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the API server
func (s *Server) Stop() error {
	s.logger.Info("Stopping REST API server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown API server: %w", err)
	}

	s.logger.Info("REST API server stopped")
	return nil
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Infof("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
