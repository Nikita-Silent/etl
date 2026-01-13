package main

import (
	"fmt"
	"log"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/ftp"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create FTP client
	client, err := ftp.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create FTP client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close FTP client: %v", err)
		}
	}()

	// Send requests to all kassas
	fmt.Println("Sending request.txt files to all kassas...")
	if err := client.SendRequestsToAllKassas(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close FTP client: %v", closeErr)
		}
		log.Fatalf("Failed to send requests: %v", err)
	}

	fmt.Println("All requests sent successfully!")
}
