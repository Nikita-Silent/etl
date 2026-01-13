package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/user/go-frontol-loader/pkg/parser"
)

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file_path>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s /path/to/frontol_export.txt\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Parse file
	fmt.Printf("Parsing file: %s\n", filePath)

	// Use "test" as source folder for parser testing
	transactions, header, err := parser.ParseFile(filePath, "test")
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	// Debug: print all transaction types found
	fmt.Printf("Debug: Found %d transaction types in map\n", len(transactions))
	for key, value := range transactions {
		fmt.Printf("Debug: Key '%s' has value of type %T\n", key, value)
	}

	// Print file header information
	fmt.Printf("File header:\n")
	fmt.Printf("  Processed: %t\n", header.Processed)
	fmt.Printf("  DB ID: %s\n", header.DBID)
	fmt.Printf("  Report Number: %s\n", header.ReportNum)
	fmt.Printf("\n")

	// Print transaction statistics
	fmt.Printf("=== PARSING STATISTICS ===\n")

	totalTransactions := 0
	for transactionType, transactionList := range transactions {
		rv := reflect.ValueOf(transactionList)
		if rv.Kind() != reflect.Slice {
			fmt.Printf("Unknown type for %s: %T\n", transactionType, transactionList)
			continue
		}
		count := rv.Len()
		totalTransactions += count
		fmt.Printf("Type %s: %d transactions\n", transactionType, count)
	}

	fmt.Printf("Total transactions parsed: %d\n", totalTransactions)
	fmt.Printf("==========================\n")

	fmt.Printf("Successfully parsed file: %s\n", filePath)
}
