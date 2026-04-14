package pipeline

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	ftplib "github.com/jlaffaye/ftp"
	ftpclient "github.com/user/go-frontol-loader/pkg/ftp"
	"github.com/user/go-frontol-loader/pkg/models"
)

func TestFolderLockManagerTimeout(t *testing.T) {
	manager := newFolderLockManager()
	release, _, err := manager.acquire(context.Background(), "L32/L32_INTER", time.Millisecond, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire first lock: %v", err)
	}
	defer release()

	_, _, err = manager.acquire(context.Background(), "L32/L32_INTER", time.Millisecond, 5*time.Millisecond)
	if err != errFolderLockTimeout {
		t.Fatalf("acquire second lock error = %v, want %v", err, errFolderLockTimeout)
	}
}

func TestProcessFolderLoadFailsIfResponseFolderStaysNonEmpty(t *testing.T) {
	folder := models.KassaFolder{KassaCode: "L32", FolderName: "L32_INTER", RequestPath: "/request/L32/L32_INTER", ResponsePath: "/response/L32/L32_INTER"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	responseCalls := 0
	requestSent := false

	mock := &ftpclient.MockClient{
		ListFilesFunc: func(path string) ([]*ftplib.Entry, error) {
			switch path {
			case folder.ResponsePath:
				responseCalls++
				if responseCalls <= 2 {
					return []*ftplib.Entry{{Name: "response.txt", Type: ftplib.EntryTypeFile, Size: 10}}, nil
				}
				return nil, nil
			case folder.RequestPath:
				return nil, nil
			default:
				return nil, nil
			}
		},
		ClearDirectoryFunc: func(path string) error { return nil },
		SendRequestToKassaFunc: func(models.KassaFolder, string) error {
			requestSent = true
			return nil
		},
	}

	result := processFolderLoad(context.Background(), mock, &mockFileLoader{}, &models.Config{RetryDelay: time.Millisecond}, "2026-03-23", folder, logger)
	if requestSent {
		t.Fatal("request should not be sent when response folder stays non-empty")
	}
	if result.Detail.LastIssueStage != "response_preflight_failed" {
		t.Fatalf("last issue stage = %q, want response_preflight_failed", result.Detail.LastIssueStage)
	}
	if result.ErrorBreakdown["response_preflight_failed"] != 1 {
		t.Fatalf("error breakdown = %+v", result.ErrorBreakdown)
	}
}

func TestProcessFolderLoadClassifiesEmptyResponse(t *testing.T) {
	folder := models.KassaFolder{KassaCode: "L32", FolderName: "L32_INTER", RequestPath: "/request/L32/L32_INTER", ResponsePath: "/response/L32/L32_INTER"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	responseCalls := 0

	mock := &ftpclient.MockClient{
		ListFilesFunc: func(path string) ([]*ftplib.Entry, error) {
			switch path {
			case folder.ResponsePath:
				responseCalls++
				switch responseCalls {
				case 1:
					return []*ftplib.Entry{{Name: "old.txt", Type: ftplib.EntryTypeFile, Size: 1}}, nil
				case 2:
					return nil, nil
				case 3:
					return []*ftplib.Entry{{Name: "response.txt", Type: ftplib.EntryTypeFile, Size: 0}}, nil
				default:
					return nil, nil
				}
			case folder.RequestPath:
				return nil, nil
			default:
				return nil, nil
			}
		},
		ClearDirectoryFunc:     func(path string) error { return nil },
		SendRequestToKassaFunc: func(models.KassaFolder, string) error { return nil },
		DownloadFileFunc: func(remotePath, localPath string) error {
			if err := os.MkdirAll(filepath.Dir(localPath), 0750); err != nil {
				return err
			}
			return os.WriteFile(localPath, []byte{}, 0600)
		},
	}

	result := processFolderLoad(context.Background(), mock, &mockFileLoader{}, &models.Config{LocalDir: t.TempDir(), RetryDelay: time.Millisecond}, "2026-03-23", folder, logger)
	if result.Detail.LastIssueStage != "empty_response" {
		t.Fatalf("last issue stage = %q, want empty_response", result.Detail.LastIssueStage)
	}
	if result.ErrorBreakdown["empty_response"] != 1 {
		t.Fatalf("error breakdown = %+v", result.ErrorBreakdown)
	}
}

func TestProcessFolderLoadMarksNoResponse(t *testing.T) {
	folder := models.KassaFolder{KassaCode: "L32", FolderName: "L32_INTER", RequestPath: "/request/L32/L32_INTER", ResponsePath: "/response/L32/L32_INTER"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	responseCalls := 0

	mock := &ftpclient.MockClient{
		ListFilesFunc: func(path string) ([]*ftplib.Entry, error) {
			switch path {
			case folder.ResponsePath:
				responseCalls++
				switch responseCalls {
				case 1:
					return []*ftplib.Entry{{Name: "old.txt", Type: ftplib.EntryTypeFile, Size: 1}}, nil
				case 2, 3:
					return nil, nil
				default:
					return nil, nil
				}
			case folder.RequestPath:
				return nil, nil
			default:
				return nil, nil
			}
		},
		ClearDirectoryFunc:     func(path string) error { return nil },
		SendRequestToKassaFunc: func(models.KassaFolder, string) error { return nil },
	}

	result := processFolderLoad(context.Background(), mock, &mockFileLoader{}, &models.Config{RetryDelay: time.Millisecond}, "2026-03-23", folder, logger)
	if result.Detail.Status != "no_response" {
		t.Fatalf("status = %q, want no_response", result.Detail.Status)
	}
	if result.ErrorBreakdown["no_response"] != 1 {
		t.Fatalf("error breakdown = %+v", result.ErrorBreakdown)
	}
}
