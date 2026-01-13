package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/models"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к БД
	database, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// Проверяем транзакцию 730423
	fmt.Println("=== Проверка транзакции 730423 ===")
	checkTransaction(ctx, database, 730423)

	// Проверяем транзакцию 2066483
	fmt.Println("\n=== Проверка транзакции 2066483 ===")
	checkTransaction(ctx, database, 2066483)

	// Проверяем все записи типа 21 за дату 2025-12-21
	fmt.Println("\n=== Статистика по типу 21 за 2025-12-21 ===")
	checkType21Stats(ctx, database)
}

func checkTransaction(ctx context.Context, database *db.Pool, transactionID int64) {
	// Проверяем в tx_bill_registration_21_23
	query := `
		SELECT 
			transaction_id_unique,
			source_folder,
			transaction_date,
			transaction_time,
			transaction_type,
			cash_register_code,
			document_number
		FROM tx_bill_registration_21_23
		WHERE transaction_id_unique = $1
	`

	rows, err := database.Query(ctx, query, transactionID)
	if err != nil {
		log.Printf("Error querying bill_registrations: %v", err)
		return
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var id int64
		var sourceFolder, date, time string
		var txType, cashRegCode int64
		var docNumber string

		if err := rows.Scan(&id, &sourceFolder, &date, &time, &txType, &cashRegCode, &docNumber); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		fmt.Printf("Найдено в tx_bill_registration_21_23:\n")
		fmt.Printf("  ID: %d\n", id)
		fmt.Printf("  Source Folder: %s\n", sourceFolder)
		fmt.Printf("  Date: %s\n", date)
		fmt.Printf("  Time: %s\n", time)
		fmt.Printf("  Type: %d\n", txType)
		fmt.Printf("  Cash Register: %d\n", cashRegCode)
		fmt.Printf("  Document: %s\n", docNumber)
	}

	if !found {
		fmt.Printf("Транзакция %d НЕ найдена в tx_bill_registration_21_23\n", transactionID)

		// Проверяем в других таблицах
		checkOtherTables(ctx, database, transactionID)
	}
}

func checkOtherTables(ctx context.Context, database *db.Pool, transactionID int64) {
	tables := make([]string, 0, len(models.TxSchemas))
	for table := range models.TxSchemas {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	for _, table := range tables {
		query := fmt.Sprintf(`
			SELECT 
				transaction_id_unique,
				source_folder,
				transaction_date,
				transaction_type
			FROM %s
			WHERE transaction_id_unique = $1
			LIMIT 1
		`, table)

		rows, err := database.Query(ctx, query, transactionID)
		if err != nil {
			continue
		}

		if rows.Next() {
			var id int64
			var sourceFolder, date string
			var txType int

			if err := rows.Scan(&id, &sourceFolder, &date, &txType); err == nil {
				fmt.Printf("Найдено в %s: ID=%d, source_folder=%s, date=%s, type=%d\n",
					table, id, sourceFolder, date, txType)
			}
		}
		rows.Close()
	}
}

func checkType21Stats(ctx context.Context, database *db.Pool) {
	query := `
		SELECT 
			source_folder,
			COUNT(*) as count,
			MIN(transaction_id_unique) as min_id,
			MAX(transaction_id_unique) as max_id
		FROM tx_bill_registration_21_23
		WHERE transaction_date = '2025-12-21'::date
		  AND transaction_type = 21
		GROUP BY source_folder
		ORDER BY source_folder
	`

	rows, err := database.Query(ctx, query)
	if err != nil {
		log.Printf("Error querying stats: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("Статистика по source_folder:")
	for rows.Next() {
		var sourceFolder string
		var count, minID, maxID int64

		if err := rows.Scan(&sourceFolder, &count, &minID, &maxID); err != nil {
			log.Printf("Error scanning stats: %v", err)
			continue
		}

		fmt.Printf("  %s: всего=%d, ID от %d до %d\n",
			sourceFolder, count, minID, maxID)
	}
}
