package ftp

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/jlaffaye/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

// Pool represents a pool of FTP connections
type Pool struct {
	cfg         *models.Config
	connections chan *Client
	size        int
	mu          sync.Mutex
	closed      bool
}

// NewPool creates a new FTP connection pool with the specified size
func NewPool(cfg *models.Config, size int) (*Pool, error) {
	if size <= 0 {
		size = 5 // Default pool size
	}

	pool := &Pool{
		cfg:         cfg,
		connections: make(chan *Client, size),
		size:        size,
		closed:      false,
	}

	// Create initial connections
	slog.Info("Creating FTP connection pool",
		"pool_size", size,
		"event", "ftp_pool_init",
	)

	for i := 0; i < size; i++ {
		client, err := NewClient(cfg)
		if err != nil {
			// Close any already created connections
			if closeErr := pool.Close(); closeErr != nil {
				slog.Warn("Failed to close FTP pool after connection error",
					"error", closeErr.Error(),
					"event", "ftp_pool_close_error",
				)
			}
			return nil, fmt.Errorf("failed to create FTP connection %d: %w", i+1, err)
		}
		pool.connections <- client
		slog.Debug("Created FTP connection",
			"connection_number", i+1,
			"event", "ftp_connection_created",
		)
	}

	slog.Info("FTP connection pool created successfully",
		"pool_size", size,
		"event", "ftp_pool_ready",
	)

	return pool, nil
}

// Get retrieves a connection from the pool (blocks if none available)
func (p *Pool) Get() (*Client, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("connection pool is closed")
	}
	p.mu.Unlock()

	// Block until a connection is available
	client := <-p.connections
	return client, nil
}

// Put returns a connection to the pool
func (p *Pool) Put(client *Client) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		// Pool is closed, close this connection
		return client.Close()
	}

	// Return connection to pool
	select {
	case p.connections <- client:
		return nil
	default:
		// Pool is full (shouldn't happen in normal operation)
		// Close the extra connection
		return client.Close()
	}
}

// Close closes all connections in the pool
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.connections)

	// Close all connections
	var lastErr error
	count := 0
	for client := range p.connections {
		if err := client.Close(); err != nil {
			lastErr = err
			slog.Warn("Failed to close FTP connection",
				"error", err.Error(),
				"event", "ftp_connection_close_error",
			)
		}
		count++
	}

	slog.Info("FTP connection pool closed",
		"connections_closed", count,
		"event", "ftp_pool_closed",
	)

	return lastErr
}

// WithConnection executes a function with a connection from the pool
// Automatically returns the connection to the pool after execution
func (p *Pool) WithConnection(fn func(*Client) error) error {
	client, err := p.Get()
	if err != nil {
		return err
	}
	defer func() {
		if putErr := p.Put(client); putErr != nil {
			slog.Warn("Failed to return connection to pool",
				"error", putErr.Error(),
				"event", "ftp_pool_put_error",
			)
		}
	}()

	return fn(client)
}

// FTPClient interface implementation - delegate to underlying connections

// ListFiles lists files in a directory using a connection from the pool
func (p *Pool) ListFiles(path string) ([]*ftp.Entry, error) {
	var result []*ftp.Entry
	err := p.WithConnection(func(client *Client) error {
		files, err := client.ListFiles(path)
		if err != nil {
			return err
		}
		result = files
		return nil
	})
	return result, err
}

// DownloadFile downloads a file from FTP server using a connection from the pool
func (p *Pool) DownloadFile(remotePath, localPath string) error {
	return p.WithConnection(func(client *Client) error {
		return client.DownloadFile(remotePath, localPath)
	})
}

// MarkFileAsProcessed marks a file as processed using a connection from the pool
func (p *Pool) MarkFileAsProcessed(remotePath string) error {
	return p.WithConnection(func(client *Client) error {
		return client.MarkFileAsProcessed(remotePath)
	})
}

// IsFileProcessed checks if a file is already processed using a connection from the pool
func (p *Pool) IsFileProcessed(remotePath string) bool {
	var result bool
	err := p.WithConnection(func(client *Client) error {
		result = client.IsFileProcessed(remotePath)
		return nil
	})
	if err != nil {
		slog.Warn("Failed to check if file is processed",
			"path", remotePath,
			"error", err.Error(),
			"event", "ftp_is_processed_error",
		)
		return false
	}
	return result
}

// DeleteProcessedFiles deletes all .processed files from a directory
func (p *Pool) DeleteProcessedFiles(path string) error {
	return p.WithConnection(func(client *Client) error {
		return client.DeleteProcessedFiles(path)
	})
}

// ClearAllKassaResponseProcessedFiles clears all .processed files from response folders
func (p *Pool) ClearAllKassaResponseProcessedFiles() error {
	return p.WithConnection(func(client *Client) error {
		return client.ClearAllKassaResponseProcessedFiles()
	})
}

// UploadFile uploads a file to FTP server using a connection from the pool
func (p *Pool) UploadFile(localPath, remotePath string) error {
	return p.WithConnection(func(client *Client) error {
		return client.UploadFile(localPath, remotePath)
	})
}

// SendRequestToKassa sends request.txt to a specific kassa folder
func (p *Pool) SendRequestToKassa(kassaFolder models.KassaFolder, date string) error {
	return p.WithConnection(func(client *Client) error {
		return client.SendRequestToKassa(kassaFolder, date)
	})
}

// ClearDirectory removes all files from a directory
func (p *Pool) ClearDirectory(path string) error {
	return p.WithConnection(func(client *Client) error {
		return client.ClearDirectory(path)
	})
}

// ClearAllKassaRequestFolders clears all kassa request folders
func (p *Pool) ClearAllKassaRequestFolders() error {
	return p.WithConnection(func(client *Client) error {
		return client.ClearAllKassaRequestFolders()
	})
}

// ClearAllKassaResponseFolders clears all kassa response folders
func (p *Pool) ClearAllKassaResponseFolders() error {
	return p.WithConnection(func(client *Client) error {
		return client.ClearAllKassaResponseFolders()
	})
}

// ClearAllKassaFolders clears both request and response folders for all kassas
func (p *Pool) ClearAllKassaFolders() error {
	return p.WithConnection(func(client *Client) error {
		return client.ClearAllKassaFolders()
	})
}

// SendRequestsToAllKassas sends request.txt to all kassa folders
func (p *Pool) SendRequestsToAllKassas() error {
	return p.WithConnection(func(client *Client) error {
		return client.SendRequestsToAllKassas()
	})
}

// SendRequestsToAllKassasWithDate sends request.txt to all kassa folders with specified date
func (p *Pool) SendRequestsToAllKassasWithDate(date string) error {
	return p.WithConnection(func(client *Client) error {
		return client.SendRequestsToAllKassasWithDate(date)
	})
}

// EnsureDirectoryExists creates a directory on FTP server if it doesn't exist
func (p *Pool) EnsureDirectoryExists(path string) error {
	return p.WithConnection(func(client *Client) error {
		return client.EnsureDirectoryExists(path)
	})
}

// EnsureKassaFoldersExist creates all kassa folder structures on FTP server
func (p *Pool) EnsureKassaFoldersExist() error {
	return p.WithConnection(func(client *Client) error {
		return client.EnsureKassaFoldersExist()
	})
}

// Ensure Pool implements FTPClient interface
var _ FTPClient = (*Pool)(nil)
