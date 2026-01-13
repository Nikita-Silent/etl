package ftp

import (
	"github.com/jlaffaye/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

// MockClient is a mock implementation of FTPClient for testing
type MockClient struct {
	CloseFunc                               func() error
	ListFilesFunc                           func(path string) ([]*ftp.Entry, error)
	DownloadFileFunc                        func(remotePath, localPath string) error
	MarkFileAsProcessedFunc                 func(remotePath string) error
	IsFileProcessedFunc                     func(remotePath string) bool
	DeleteProcessedFilesFunc                func(path string) error
	ClearAllKassaResponseProcessedFilesFunc func() error
	UploadFileFunc                          func(localPath, remotePath string) error
	SendRequestToKassaFunc                  func(kassaFolder models.KassaFolder, date string) error
	ClearDirectoryFunc                      func(path string) error
	ClearAllKassaRequestFoldersFunc         func() error
	ClearAllKassaResponseFoldersFunc        func() error
	ClearAllKassaFoldersFunc                func() error
	SendRequestsToAllKassasFunc             func() error
	SendRequestsToAllKassasWithDateFunc     func(date string) error
	EnsureDirectoryExistsFunc               func(path string) error
	EnsureKassaFoldersExistFunc             func() error
}

func (m *MockClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockClient) ListFiles(path string) ([]*ftp.Entry, error) {
	if m.ListFilesFunc != nil {
		return m.ListFilesFunc(path)
	}
	return []*ftp.Entry{}, nil
}

func (m *MockClient) DownloadFile(remotePath, localPath string) error {
	if m.DownloadFileFunc != nil {
		return m.DownloadFileFunc(remotePath, localPath)
	}
	return nil
}

func (m *MockClient) MarkFileAsProcessed(remotePath string) error {
	if m.MarkFileAsProcessedFunc != nil {
		return m.MarkFileAsProcessedFunc(remotePath)
	}
	return nil
}

func (m *MockClient) IsFileProcessed(remotePath string) bool {
	if m.IsFileProcessedFunc != nil {
		return m.IsFileProcessedFunc(remotePath)
	}
	return false
}

func (m *MockClient) DeleteProcessedFiles(path string) error {
	if m.DeleteProcessedFilesFunc != nil {
		return m.DeleteProcessedFilesFunc(path)
	}
	return nil
}

func (m *MockClient) ClearAllKassaResponseProcessedFiles() error {
	if m.ClearAllKassaResponseProcessedFilesFunc != nil {
		return m.ClearAllKassaResponseProcessedFilesFunc()
	}
	return nil
}

func (m *MockClient) UploadFile(localPath, remotePath string) error {
	if m.UploadFileFunc != nil {
		return m.UploadFileFunc(localPath, remotePath)
	}
	return nil
}

func (m *MockClient) SendRequestToKassa(kassaFolder models.KassaFolder, date string) error {
	if m.SendRequestToKassaFunc != nil {
		return m.SendRequestToKassaFunc(kassaFolder, date)
	}
	return nil
}

func (m *MockClient) ClearDirectory(path string) error {
	if m.ClearDirectoryFunc != nil {
		return m.ClearDirectoryFunc(path)
	}
	return nil
}

func (m *MockClient) ClearAllKassaRequestFolders() error {
	if m.ClearAllKassaRequestFoldersFunc != nil {
		return m.ClearAllKassaRequestFoldersFunc()
	}
	return nil
}

func (m *MockClient) ClearAllKassaResponseFolders() error {
	if m.ClearAllKassaResponseFoldersFunc != nil {
		return m.ClearAllKassaResponseFoldersFunc()
	}
	return nil
}

func (m *MockClient) ClearAllKassaFolders() error {
	if m.ClearAllKassaFoldersFunc != nil {
		return m.ClearAllKassaFoldersFunc()
	}
	return nil
}

func (m *MockClient) SendRequestsToAllKassas() error {
	if m.SendRequestsToAllKassasFunc != nil {
		return m.SendRequestsToAllKassasFunc()
	}
	return nil
}

func (m *MockClient) SendRequestsToAllKassasWithDate(date string) error {
	if m.SendRequestsToAllKassasWithDateFunc != nil {
		return m.SendRequestsToAllKassasWithDateFunc(date)
	}
	return nil
}

func (m *MockClient) EnsureDirectoryExists(path string) error {
	if m.EnsureDirectoryExistsFunc != nil {
		return m.EnsureDirectoryExistsFunc(path)
	}
	return nil
}

func (m *MockClient) EnsureKassaFoldersExist() error {
	if m.EnsureKassaFoldersExistFunc != nil {
		return m.EnsureKassaFoldersExistFunc()
	}
	return nil
}
