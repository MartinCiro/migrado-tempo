// internal/infrastructure/delivery/web/server.go
package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email/config"
)

type Server struct {
	handler    *WebhookHandler
	httpServer *http.Server
	config     *config.WebhookConfig
}

func NewServer(handler *WebhookHandler, webhookConfig *config.WebhookConfig) *Server {
	return &Server{
		handler: handler,
		config:  webhookConfig,
	}
}

func (s *Server) Start() error {
	// Configurar el servidor HTTP
	s.httpServer = &http.Server{
		Addr:         s.config.GetAddress(),
		Handler:      s.setupRoutes(),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// Manejar shutdown graceful
	go s.gracefulShutdown()

	log.Printf("🌐 Servidor webhook escuchando en %s", s.config.GetBaseURL())
	log.Printf("📋 Endpoints disponibles:")
	log.Printf("   POST %s/webhook/process-xml", s.config.GetBaseURL())

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Endpoint principal
	mux.HandleFunc("/webhook/process-xml", s.handler.HandleProcessXML)
	// Info del servidor
	mux.HandleFunc("/", s.serverInfo)

	return mux
}

func (s *Server) gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down webhook server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("❌ Error shutting down server: %v", err)
	}

	log.Println("✅ Webhook server stopped")
}

/* func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"status": "healthy",
		"service": "email-processor-webhook",
		"host": "%s",
		"port": "%s",
		"timestamp": "%s"
	}`, s.config.Host, s.config.Port, time.Now().Format(time.RFC3339))
} */

func (s *Server) serverInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"service": "Email Processor Webhook",
		"version": "1.0.0",
		"endpoints": {
			"process_xml": "%s/webhook/process-xml",
		}
	}`, s.config.GetBaseURL())
}
