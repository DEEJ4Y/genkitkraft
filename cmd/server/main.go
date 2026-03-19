package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DEEJ4Y/genkitkraft/internal/config"
	"github.com/DEEJ4Y/genkitkraft/internal/services"
)

func main() {
	cfg := config.Load()

	srv, err := services.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Graceful shutdown on interrupt
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		srv.Stop()
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
