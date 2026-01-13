package ftp

import (
	"github.com/jlaffaye/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

// FTPClient defines the interface for FTP operations
// This allows for easier testing with mocks
type FTPClient interface {
	Close() error
	ListFiles(path string) ([]*ftp.Entry, error)
	DownloadFile(remotePath, localPath string) error
	MarkFileAsProcessed(remotePath string) error
	IsFileProcessed(remotePath string) bool
	DeleteProcessedFiles(path string) error
	ClearAllKassaResponseProcessedFiles() error
	UploadFile(localPath, remotePath string) error
	SendRequestToKassa(kassaFolder models.KassaFolder, date string) error
	ClearDirectory(path string) error
	ClearAllKassaRequestFolders() error
	ClearAllKassaResponseFolders() error
	ClearAllKassaFolders() error
	SendRequestsToAllKassas() error
	SendRequestsToAllKassasWithDate(date string) error
	EnsureDirectoryExists(path string) error
	EnsureKassaFoldersExist() error
}

// Ensure Client implements FTPClient interface
var _ FTPClient = (*Client)(nil)
