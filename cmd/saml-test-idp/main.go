package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/breakroom/saml-test-idp/internal/config"
	"github.com/breakroom/saml-test-idp/internal/idp"
)

var (
	version = "dev"
)

func main() {
	// Define CLI flags
	configPath := flag.String("config", "config.yaml", "Path to YAML configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("saml-test-idp version %s\n", version)
		os.Exit(0)
	}

	// Load configuration from YAML file
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set default base URL if not provided
	if cfg.Server.BaseURL == "" {
		cfg.Server.BaseURL = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	}

	// Set default entity ID if not provided
	if cfg.IDP.EntityID == "" {
		cfg.IDP.EntityID = cfg.Server.BaseURL + "/metadata"
	}

	// Create IDP server
	idpServer, err := idp.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create IDP server: %v", err)
	}

	// Set up HTTP routes
	mux := http.NewServeMux()
	idpServer.RegisterRoutes(mux)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// Create HTTP server with explicit configuration
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Channel to listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting SAML IDP server on %s", addr)
		log.Printf("  Metadata URL: %s/metadata", cfg.Server.BaseURL)
		log.Printf("  SSO URL: %s/sso", cfg.Server.BaseURL)
		log.Printf("Press Ctrl+C to stop")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
