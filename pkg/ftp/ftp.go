package ftp

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

// Client represents FTP client
type Client struct {
	conn *ftp.ServerConn
	cfg  *models.Config
}

// NewClient creates a new FTP client
func NewClient(cfg *models.Config) (*Client, error) {
	// Connect to FTP server
	conn, err := ftp.Dial(cfg.FTPHost+":"+fmt.Sprintf("%d", cfg.FTPPort), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to FTP server: %w", err)
	}

	// Login
	if err := conn.Login(cfg.FTPUser, cfg.FTPPassword); err != nil {
		_ = conn.Quit()
		return nil, fmt.Errorf("failed to login to FTP server: %w", err)
	}

	client := &Client{
		conn: conn,
		cfg:  cfg,
	}

	// Ensure all kassa folders exist
	slog.Info("Ensuring kassa folder structure exists on FTP server",
		"event", "ftp_ensure_folders",
	)
	if err := client.EnsureKassaFoldersExist(); err != nil {
		_ = conn.Quit()
		return nil, fmt.Errorf("failed to create kassa folders: %w", err)
	}

	return client, nil
}

// Close closes FTP connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Quit()
	}
	return nil
}

// ListFiles lists files in a directory
func (c *Client) ListFiles(path string) ([]*ftp.Entry, error) {
	// Ensure path starts with / for absolute path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Try to change to directory first to ensure it exists
	if err := c.conn.ChangeDir(path); err != nil {
		return nil, fmt.Errorf("failed to change to directory %s: %w", path, err)
	}

	// List files in the directory
	entries, err := c.conn.List(".")
	if err != nil {
		// Try with full path as fallback
		entries, err = c.conn.List(path)
		if err != nil {
			return nil, fmt.Errorf("failed to list files in %s: %w", path, err)
		}
	}

	// Filter only files (not directories)
	var files []*ftp.Entry
	for _, entry := range entries {
		if entry.Type == ftp.EntryTypeFile && !strings.HasPrefix(entry.Name, ".") {
			files = append(files, entry)
		}
	}

	return files, nil
}

// DownloadFile downloads a file from FTP server
func (c *Client) DownloadFile(remotePath, localPath string) error {
	// Create local directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Download file
	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return fmt.Errorf("failed to retrieve file %s: %w", remotePath, err)
	}
	defer func() {
		if err := resp.Close(); err != nil {
			slog.Warn("Failed to close FTP response",
				"error", err.Error(),
				"event", "ftp_response_close_error",
			)
		}
	}()

	// Create local file
	// #nosec G304 -- localPath is controlled by configuration.
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer func() {
		if err := localFile.Close(); err != nil {
			slog.Warn("Failed to close local file",
				"error", err.Error(),
				"event", "ftp_local_file_close_error",
			)
		}
	}()

	// Copy data
	_, err = io.Copy(localFile, resp)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	return nil
}

// MarkFileAsProcessed marks a file as processed by renaming it
func (c *Client) MarkFileAsProcessed(remotePath string) error {
	// Rename file to add .processed suffix
	processedPath := remotePath + ".processed"
	return c.conn.Rename(remotePath, processedPath)
}

// IsFileProcessed checks if a file is already processed
func (c *Client) IsFileProcessed(remotePath string) bool {
	processedPath := remotePath + ".processed"
	_, err := c.conn.List(processedPath)
	return err == nil
}

// DeleteProcessedFiles deletes all .processed files from a directory
func (c *Client) DeleteProcessedFiles(path string) error {
	// Save current directory
	currentDir, err := c.conn.CurrentDir()
	if err != nil {
		currentDir = ""
	}

	// Change to target directory
	if err := c.conn.ChangeDir(path); err != nil {
		// Restore original directory if we changed it
		if currentDir != "" {
			_ = c.conn.ChangeDir(currentDir)
		}
		return fmt.Errorf("failed to change to directory %s: %w", path, err)
	}

	// List all files in the directory
	entries, err := c.conn.List(".")
	if err != nil {
		// Try with full path as fallback
		entries, err = c.conn.List(path)
		if err != nil {
			// Restore original directory if we changed it
			if currentDir != "" {
				_ = c.conn.ChangeDir(currentDir)
			}
			return fmt.Errorf("failed to list files in %s: %w", path, err)
		}
	}

	// Delete all .processed files
	deletedCount := 0
	for _, entry := range entries {
		if entry.Type == ftp.EntryTypeFile && strings.HasSuffix(entry.Name, ".processed") {
			// Use relative path since we're in the directory
			filePath := entry.Name
			if err := c.conn.Delete(filePath); err != nil {
				// Try with full path as fallback
				fullPath := path + "/" + entry.Name
				if err := c.conn.Delete(fullPath); err != nil {
					slog.Warn("Failed to delete processed file",
						"file", fullPath,
						"error", err.Error(),
						"event", "ftp_delete_processed_warning",
					)
				} else {
					deletedCount++
					slog.Debug("Deleted processed file",
						"file", fullPath,
						"event", "ftp_processed_file_deleted",
					)
				}
			} else {
				deletedCount++
				slog.Debug("Deleted processed file",
					"file", path+"/"+entry.Name,
					"event", "ftp_processed_file_deleted",
				)
			}
		}
	}

	// Restore original directory
	if currentDir != "" {
		_ = c.conn.ChangeDir(currentDir)
	}

	if deletedCount > 0 {
		slog.Info("Deleted processed files",
			"path", path,
			"count", deletedCount,
			"event", "ftp_processed_files_deleted",
		)
	}

	return nil
}

// ClearAllKassaResponseProcessedFiles clears all .processed files from response folders
func (c *Client) ClearAllKassaResponseProcessedFiles() error {
	folders := GetAllKassaFolders(c.cfg)

	slog.Info("Clearing all .processed files from response folders",
		"event", "ftp_clear_processed_files",
	)

	for _, folder := range folders {
		if err := c.DeleteProcessedFiles(folder.ResponsePath); err != nil {
			slog.Warn("Failed to clear processed files from response folder",
				"folder", folder.ResponsePath,
				"error", err.Error(),
				"event", "ftp_clear_processed_warning",
			)
		}
	}

	return nil
}

// GetAllKassaFolders returns all kassa folders from configuration
func GetAllKassaFolders(cfg *models.Config) []models.KassaFolder {
	var folders []models.KassaFolder

	// Parse kassa structure from configuration
	for kassaCode, folderNames := range cfg.KassaStructure {
		for _, folderName := range folderNames {
			folder := models.KassaFolder{
				KassaCode:    kassaCode,
				FolderName:   folderName,
				RequestPath:  cfg.FTPRequestDir + "/" + kassaCode + "/" + folderName,
				ResponsePath: cfg.FTPResponseDir + "/" + kassaCode + "/" + folderName,
			}
			folders = append(folders, folder)
		}
	}

	return folders
}

// UploadFile uploads a file to FTP server
func (c *Client) UploadFile(localPath, remotePath string) error {
	// Open local file
	// #nosec G304 -- localPath is controlled by configuration.
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer func() {
		if err := localFile.Close(); err != nil {
			slog.Warn("Failed to close local file",
				"error", err.Error(),
				"event", "ftp_local_file_close_error",
			)
		}
	}()

	// Upload file
	err = c.conn.Stor(remotePath, localFile)
	if err != nil {
		return fmt.Errorf("failed to upload file to %s: %w", remotePath, err)
	}

	return nil
}

// CreateRequestFile creates a request.txt file with specified date range
func CreateRequestFile(localDir string, date string) (string, error) {
	// Parse the date string (expected format: YYYY-MM-DD)
	var requestDate string
	if date == "" {
		// Use current date if not provided
		requestDate = time.Now().Format("02.01.2006")
	} else {
		// Parse the provided date
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return "", fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
		}
		requestDate = parsedDate.Format("02.01.2006")
	}

	// Create request content
	requestContent := fmt.Sprintf("$$$TRANSACTIONSBYDATERANGE\n%s; %s", requestDate, requestDate)

	// Create local file path
	requestPath := filepath.Join(localDir, "request.txt")

	// Create local directory if it doesn't exist
	if err := os.MkdirAll(localDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create local directory: %w", err)
	}

	// Write request file
	if err := os.WriteFile(requestPath, []byte(requestContent), 0600); err != nil {
		return "", fmt.Errorf("failed to write request file: %w", err)
	}

	return requestPath, nil
}

// SendRequestToKassa sends request.txt to a specific kassa folder
func (c *Client) SendRequestToKassa(kassaFolder models.KassaFolder, date string) error {
	// Create request file
	requestPath, err := CreateRequestFile(c.cfg.LocalDir, date)
	if err != nil {
		return fmt.Errorf("failed to create request file: %w", err)
	}
	defer func() {
		if err := os.Remove(requestPath); err != nil {
			slog.Warn("Failed to remove request file",
				"path", requestPath,
				"error", err.Error(),
				"event", "request_file_remove_error",
			)
		}
	}()

	// Upload to FTP
	remotePath := kassaFolder.RequestPath + "/request.txt"
	if err := c.UploadFile(requestPath, remotePath); err != nil {
		return fmt.Errorf("failed to upload request to %s: %w", remotePath, err)
	}

	return nil
}

// ClearDirectory removes all files from a directory
func (c *Client) ClearDirectory(path string) error {
	// Ensure path starts with / for absolute path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Save current directory to restore it later
	currentDir, err := c.conn.CurrentDir()
	if err != nil {
		// If we can't get current directory, continue anyway
		currentDir = ""
	}

	// Change to target directory
	if err := c.conn.ChangeDir(path); err != nil {
		// Directory might not exist, which is fine
		return nil
	}

	// List all files in the directory
	entries, err := c.conn.List(".")
	if err != nil {
		// Try with full path as fallback
		entries, err = c.conn.List(path)
		if err != nil {
			// Restore original directory if we changed it
			if currentDir != "" {
				_ = c.conn.ChangeDir(currentDir)
			}
			return nil
		}
	}

	// Delete all files (not directories)
	for _, entry := range entries {
		if entry.Type == ftp.EntryTypeFile {
			// Use relative path since we're in the directory
			filePath := entry.Name
			if err := c.conn.Delete(filePath); err != nil {
				// Try with full path as fallback
				fullPath := path + "/" + entry.Name
				if err := c.conn.Delete(fullPath); err != nil {
					slog.Warn("Failed to delete file",
						"file", fullPath,
						"error", err.Error(),
						"event", "ftp_delete_warning",
					)
				} else {
					slog.Debug("Deleted file",
						"file", fullPath,
						"event", "ftp_file_deleted",
					)
				}
			} else {
				slog.Debug("Deleted file",
					"file", path+"/"+entry.Name,
					"event", "ftp_file_deleted",
				)
			}
		}
	}

	// Restore original directory
	if currentDir != "" {
		_ = c.conn.ChangeDir(currentDir)
	}

	return nil
}

// ClearAllKassaRequestFolders clears all kassa request folders
func (c *Client) ClearAllKassaRequestFolders() error {
	folders := GetAllKassaFolders(c.cfg)

	slog.Info("Clearing all kassa request folders",
		"event", "ftp_clear_request_folders",
	)
	for _, folder := range folders {
		if err := c.ClearDirectory(folder.RequestPath); err != nil {
			slog.Warn("Failed to clear request folder",
				"folder", folder.RequestPath,
				"error", err.Error(),
				"event", "ftp_clear_folder_warning",
			)
		} else {
			slog.Debug("Cleared request folder",
				"folder", folder.RequestPath,
				"event", "ftp_folder_cleared",
			)
		}
	}

	return nil
}

// ClearAllKassaResponseFolders clears all kassa response folders
func (c *Client) ClearAllKassaResponseFolders() error {
	folders := GetAllKassaFolders(c.cfg)

	slog.Info("Clearing all kassa response folders",
		"event", "ftp_clear_response_folders",
	)
	for _, folder := range folders {
		if err := c.ClearDirectory(folder.ResponsePath); err != nil {
			slog.Warn("Failed to clear response folder",
				"folder", folder.ResponsePath,
				"error", err.Error(),
				"event", "ftp_clear_folder_warning",
			)
		} else {
			slog.Debug("Cleared response folder",
				"folder", folder.ResponsePath,
				"event", "ftp_folder_cleared",
			)
		}
	}

	return nil
}

// ClearAllKassaFolders clears both request and response folders for all kassas
func (c *Client) ClearAllKassaFolders() error {
	// Clear request folders
	if err := c.ClearAllKassaRequestFolders(); err != nil {
		slog.Warn("Failed to clear some request folders",
			"error", err.Error(),
			"event", "ftp_clear_folders_warning",
		)
	}

	// Clear response folders
	if err := c.ClearAllKassaResponseFolders(); err != nil {
		slog.Warn("Failed to clear some response folders",
			"error", err.Error(),
			"event", "ftp_clear_folders_warning",
		)
	}

	return nil
}

// SendRequestsToAllKassas sends request.txt to all kassa folders (legacy function)
func (c *Client) SendRequestsToAllKassas() error {
	return c.SendRequestsToAllKassasWithDate("")
}

// SendRequestsToAllKassasWithDate sends request.txt to all kassa folders with specified date
func (c *Client) SendRequestsToAllKassasWithDate(date string) error {
	// First clear all request folders
	if err := c.ClearAllKassaRequestFolders(); err != nil {
		slog.Warn("Failed to clear some folders",
			"error", err.Error(),
			"event", "ftp_clear_folders_warning",
		)
	}

	// Then send new requests
	folders := GetAllKassaFolders(c.cfg)

	if date == "" {
		slog.Info("Sending request.txt files to all kassas with current date",
			"event", "ftp_send_requests",
		)
	} else {
		slog.Info("Sending request.txt files to all kassas",
			"date", date,
			"event", "ftp_send_requests",
		)
	}

	for _, folder := range folders {
		if err := c.SendRequestToKassa(folder, date); err != nil {
			slog.Warn("Failed to send request to kassa",
				"kassa_code", folder.KassaCode,
				"folder", folder.FolderName,
				"error", err.Error(),
				"event", "ftp_send_request_error",
			)
			continue
		}
		slog.Info("Successfully sent request to kassa",
			"kassa_code", folder.KassaCode,
			"folder", folder.FolderName,
			"event", "ftp_request_sent",
		)
	}

	return nil
}

// EnsureDirectoryExists creates a directory on FTP server if it doesn't exist
func (c *Client) EnsureDirectoryExists(path string) error {
	// Try to list the directory - if it exists, this will succeed
	_, err := c.conn.List(path)
	if err == nil {
		// Directory exists
		return nil
	}

	// Directory doesn't exist, try to create it
	// Split path into components
	parts := strings.Split(strings.Trim(path, "/"), "/")
	currentPath := ""

	for _, part := range parts {
		if currentPath == "" {
			currentPath = "/" + part
		} else {
			currentPath = currentPath + "/" + part
		}

		// Try to create directory
		err := c.conn.MakeDir(currentPath)
		if err != nil {
			// Check if directory already exists (some FTP servers return error even if it exists)
			_, listErr := c.conn.List(currentPath)
			if listErr != nil {
				// Directory really doesn't exist and creation failed
				return fmt.Errorf("failed to create directory %s: %w", currentPath, err)
			}
			// Directory exists now (maybe created by another process), continue
		}
	}

	return nil
}

// EnsureKassaFoldersExist creates all kassa folder structures on FTP server
func (c *Client) EnsureKassaFoldersExist() error {
	folders := GetAllKassaFolders(c.cfg)

	if len(folders) == 0 {
		slog.Warn("No kassa folders found in configuration",
			"event", "ftp_no_folders",
		)
		return nil
	}

	slog.Info("Creating kassa folder structures",
		"count", len(folders),
		"event", "ftp_create_folders",
	)

	// Create base directories first
	if err := c.EnsureDirectoryExists(c.cfg.FTPRequestDir); err != nil {
		return fmt.Errorf("failed to create request directory: %w", err)
	}
	slog.Debug("Base request directory exists",
		"path", c.cfg.FTPRequestDir,
	)

	if err := c.EnsureDirectoryExists(c.cfg.FTPResponseDir); err != nil {
		return fmt.Errorf("failed to create response directory: %w", err)
	}
	slog.Debug("Base response directory exists",
		"path", c.cfg.FTPResponseDir,
	)

	// Create all kassa folders
	for _, folder := range folders {
		if err := c.EnsureDirectoryExists(folder.RequestPath); err != nil {
			return fmt.Errorf("failed to create request folder %s: %w", folder.RequestPath, err)
		}
		slog.Debug("Created request folder",
			"folder", folder.RequestPath,
			"event", "ftp_folder_created",
		)

		if err := c.EnsureDirectoryExists(folder.ResponsePath); err != nil {
			return fmt.Errorf("failed to create response folder %s: %w", folder.ResponsePath, err)
		}
		slog.Debug("Created response folder",
			"folder", folder.ResponsePath,
			"event", "ftp_folder_created",
		)
	}

	slog.Info("Successfully created all kassa folder structures",
		"count", len(folders),
		"event", "ftp_folders_created",
	)
	return nil
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(delay * time.Duration(i+1))
				continue
			}
		} else {
			return nil
		}
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}
