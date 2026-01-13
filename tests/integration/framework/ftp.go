//go:build integration
// +build integration

package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/user/go-frontol-loader/pkg/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

const (
	ftpImage    = "fauria/vsftpd:latest"
	ftpUser     = "frontol"
	ftpPassword = "test_password"
)

// FTPContainer wraps testcontainers FTP instance
type FTPContainer struct {
	Container testcontainers.Container
	Config    *models.Config
	Client    *ftp.Client
}

// NewFTPContainer creates and starts an FTP test container
func NewFTPContainer(ctx context.Context) (*FTPContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        ftpImage,
		ExposedPorts: []string{"21/tcp", "21000-21010/tcp"},
		Env: map[string]string{
			"FTP_USER": ftpUser,
			"FTP_PASS": ftpPassword,
			"PASV_ADDRESS": "127.0.0.1",
			"PASV_MIN_PORT": "21000",
			"PASV_MAX_PORT": "21010",
		},
		WaitingFor: wait.ForListeningPort("21/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start FTP container: %w", err)
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "21")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	// Wait a bit for FTP server to be fully ready
	time.Sleep(2 * time.Second)

	cfg := &models.Config{
		FTPHost:        host,
		FTPPort:        mappedPort.Int(),
		FTPUser:        ftpUser,
		FTPPassword:    ftpPassword,
		FTPRequestDir:  "/request",
		FTPResponseDir: "/response",
		FTPPoolSize:    2,
		KassaStructure: map[string][]string{
			"001": {"folder1"},
			"002": {"folder1"},
		},
		LocalDir: "/tmp/frontol_test",
	}

	// Create FTP client
	client, err := ftp.NewClient(cfg)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create FTP client: %w", err)
	}

	return &FTPContainer{
		Container: container,
		Config:    cfg,
		Client:    client,
	}, nil
}

// Close closes the FTP client and terminates the container
func (fc *FTPContainer) Close(ctx context.Context) error {
	if fc.Client != nil {
		fc.Client.Close()
	}
	if fc.Container != nil {
		return fc.Container.Terminate(ctx)
	}
	return nil
}

// SetupFolderStructure creates the expected folder structure on FTP server
func (fc *FTPContainer) SetupFolderStructure(ctx context.Context) error {
	if err := fc.Client.EnsureKassaFoldersExist(); err != nil {
		return fmt.Errorf("failed to create folder structure: %w", err)
	}
	return nil
}

// CleanFolders removes all files from FTP folders
func (fc *FTPContainer) CleanFolders(ctx context.Context) error {
	if err := fc.Client.ClearAllKassaFolders(); err != nil {
		return fmt.Errorf("failed to clean folders: %w", err)
	}
	return nil
}

// GetConnectionString returns FTP connection info
func (fc *FTPContainer) GetConnectionString() string {
	return fmt.Sprintf("ftp://%s:%s@%s:%d",
		fc.Config.FTPUser,
		fc.Config.FTPPassword,
		fc.Config.FTPHost,
		fc.Config.FTPPort,
	)
}
