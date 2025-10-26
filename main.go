package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rom8726/floxy-ui/internal/config"
	"github.com/rom8726/floxy-ui/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}
	defer srv.Close()

	// Handle a graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		srv.Close()
		os.Exit(0)
	}()

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
