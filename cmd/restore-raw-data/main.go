package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/db"
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

	// Транзакции для восстановления
	transactions := map[int64]string{
		730423:  "730423;21.12.2025;9:51:53;21;1;62853;30000004;0;;4000;1;4000;4;2406;4000;4000;2;;;4000;;0;5;0;;1/2406/62853;1;;;;;;;;;;;;;;;;;;",
		2066483: "2066483;21.12.2025;20:40:35;21;5;93409;36000007;0;;4000;1;4000;4;2274;4000;4000;1;;;4000;;0;5;0;;5/2274/93409;1;;;;;;;;;;;;;;;;;;",
	}

	// Обновляем raw_data для каждой транзакции
	for transactionID, rawData := range transactions {
		// Проверяем, существует ли запись
		var exists bool
		checkQuery := `
			SELECT EXISTS(
				SELECT 1 FROM bill_registrations 
				WHERE transaction_id_unique = $1
			)
		`
		err := database.QueryRow(ctx, checkQuery, transactionID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking transaction %d: %v", transactionID, err)
			continue
		}

		if !exists {
			log.Printf("Transaction %d not found in bill_registrations", transactionID)
			continue
		}

		// Обновляем raw_data
		updateQuery := `
			UPDATE bill_registrations
			SET raw_data = $1
			WHERE transaction_id_unique = $2
		`

		// Экранируем специальные символы для SQL
		escapedRawData := strings.ReplaceAll(rawData, "'", "''")

		result, err := database.Exec(ctx, updateQuery, escapedRawData, transactionID)
		if err != nil {
			log.Printf("Error updating transaction %d: %v", transactionID, err)
			continue
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("✓ Обновлена транзакция %d: raw_data восстановлен (%d строк)\n", transactionID, rowsAffected)
		} else {
			fmt.Printf("✗ Транзакция %d не обновлена (возможно, уже имеет raw_data)\n", transactionID)
		}
	}

	fmt.Println("\nГотово!")
}
