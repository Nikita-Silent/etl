package ftp

import (
	"strings"

	"github.com/jlaffaye/ftp"
)

// FilterUnprocessedFiles filters out files that have already been processed
// This function performs a single FTP LIST operation and filters in memory,
// which is much more efficient than calling IsFileProcessed for each file (O(n) vs O(n²))
func FilterUnprocessedFiles(allFiles []*ftp.Entry) []*ftp.Entry {
	// Return empty slice for nil input
	if allFiles == nil {
		return []*ftp.Entry{}
	}

	// Create a set of processed file names (without .processed extension)
	processedSet := make(map[string]bool)

	// First pass: identify all .processed files
	for _, file := range allFiles {
		if strings.HasSuffix(file.Name, ".processed") {
			// Remove .processed extension to get original filename
			originalName := strings.TrimSuffix(file.Name, ".processed")
			processedSet[originalName] = true
		}
	}

	// Second pass: filter out processed files and .processed files themselves
	var unprocessedFiles []*ftp.Entry
	for _, file := range allFiles {
		// Skip .processed files themselves
		if strings.HasSuffix(file.Name, ".processed") {
			continue
		}

		// Skip files that have a .processed version
		if processedSet[file.Name] {
			continue
		}

		unprocessedFiles = append(unprocessedFiles, file)
	}

	return unprocessedFiles
}

// FilterFilesByName filters files by excluding specific names
func FilterFilesByName(files []*ftp.Entry, excludeNames ...string) []*ftp.Entry {
	excludeSet := make(map[string]bool)
	for _, name := range excludeNames {
		excludeSet[name] = true
	}

	var filtered []*ftp.Entry
	for _, file := range files {
		if !excludeSet[file.Name] {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// GetProcessedFileNames extracts names of files that have been processed
// Returns a map for O(1) lookup
func GetProcessedFileNames(allFiles []*ftp.Entry) map[string]bool {
	processedSet := make(map[string]bool)

	for _, file := range allFiles {
		if strings.HasSuffix(file.Name, ".processed") {
			originalName := strings.TrimSuffix(file.Name, ".processed")
			processedSet[originalName] = true
		}
	}

	return processedSet
}

// IsFileInProcessedSet checks if a file name is in the processed set
func IsFileInProcessedSet(fileName string, processedSet map[string]bool) bool {
	return processedSet[fileName]
}
