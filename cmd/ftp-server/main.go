package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFTPConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(logger.Config{
		Level:   cfg.LogLevel,
		Format:  "text",
		Output:  os.Stdout,
		Backend: os.Getenv("LOG_BACKEND"),
	})

	// Create FTP driver
	driver := &FTPDriver{
		rootPath: cfg.FTPRootPath,
		user:     cfg.FTPUser,
		password: cfg.FTPPassword,
		log:      log,
		cfg:      cfg,
	}

	// Create FTP server
	server := ftpserver.NewFtpServer(driver)

	// Start server
	log.Info("Starting FTP server",
		"listen_addr", fmt.Sprintf("0.0.0.0:%d", cfg.FTPPort),
		"public_host", cfg.PublicHost,
		"passive_ports", fmt.Sprintf("%d-%d", cfg.PassiveMinPort, cfg.PassiveMaxPort),
		"root_path", driver.rootPath)

	if err := server.ListenAndServe(); err != nil {
		log.Error("Failed to start FTP server", "error", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down FTP server...")
	if err := server.Stop(); err != nil {
		log.Error("Error stopping server", "error", err)
		os.Exit(1)
	}

	log.Info("FTP server stopped")
}

// FTPDriver implements ftpserver.MainDriver interface
type FTPDriver struct {
	rootPath string
	user     string
	password string
	log      *logger.Logger
	cfg      *config.FTPConfig
}

// GetSettings returns server settings
func (d *FTPDriver) GetSettings() (*ftpserver.Settings, error) {
	// Create port range for passive mode
	portRange := &ftpserver.PortRange{
		Start: d.cfg.PassiveMinPort,
		End:   d.cfg.PassiveMaxPort,
	}

	settings := &ftpserver.Settings{
		ListenAddr:               fmt.Sprintf("0.0.0.0:%d", d.cfg.FTPPort),
		PublicHost:               d.cfg.PublicHost,
		PassiveTransferPortRange: portRange,
		IdleTimeout:              300,
		ConnectionTimeout:        30,
		DisableActiveMode:        false,                     // Enable active mode for Frontol compatibility
		ActiveConnectionsCheck:   ftpserver.IPMatchDisabled, // Disable IP check for active mode (helps with NAT)
		Banner:                   "220 Welcome to Frontol FTP Server",
	}

	return settings, nil
}

// GetTLSConfig returns TLS configuration (not used, return nil)
func (d *FTPDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, nil
}

// ClientConnected is called when a client connects
func (d *FTPDriver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	d.log.Info("Client connected", "remote_addr", cc.RemoteAddr())
	return "220 Welcome to Frontol FTP Server", nil
}

// ClientDisconnected is called when a client disconnects
func (d *FTPDriver) ClientDisconnected(cc ftpserver.ClientContext) {
	d.log.Info("Client disconnected", "remote_addr", cc.RemoteAddr())
}

// AuthUser authenticates a user
func (d *FTPDriver) AuthUser(cc ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	if user != d.user || pass != d.password {
		d.log.Warn("Authentication failed", "user", user)
		return nil, fmt.Errorf("invalid credentials")
	}

	d.log.Info("User authenticated successfully", "user", user)

	// Create afero filesystem with base path
	baseFs := afero.NewBasePathFs(afero.NewOsFs(), d.rootPath)

	return &FTPClientDriver{
		Fs:  baseFs,
		log: d.log,
	}, nil
}

// FTPClientDriver implements ftpserver.ClientDriver interface (afero.Fs)
type FTPClientDriver struct {
	afero.Fs
	log *logger.Logger
}
