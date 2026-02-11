package main

import (
	"flag"
	"log"

	"share-sniffer/internal/httpapi"
	"share-sniffer/internal/httpapi/httpconfig"
)

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "./config/httpconfig.toml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := httpconfig.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create and run server
	server := httpapi.NewServer(cfg)
	if err := server.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
