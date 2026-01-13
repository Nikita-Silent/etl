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

	// Clear all kassa folders (both request and response)
	fmt.Println("Clearing all kassa folders...")
	if err := client.ClearAllKassaFolders(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Failed to close FTP client: %v", closeErr)
		}
		log.Fatalf("Failed to clear folders: %v", err)
	}

	fmt.Println("All kassa folders cleared successfully!")
}
