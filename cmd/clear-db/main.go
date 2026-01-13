package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/user/go-frontol-loader/pkg/config"
	"github.com/user/go-frontol-loader/pkg/db"
	"github.com/user/go-frontol-loader/pkg/models"
)

func main() {
	// Загружаем конфигурацию БД
	cfg, err := config.LoadDBConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Подключаемся к БД
	database, err := db.NewPool(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Проверяем флаг подтверждения
	if len(os.Args) > 1 && os.Args[1] != "--confirm" {
		fmt.Println("⚠️  WARNING: This will delete all transaction data from the database!")
		fmt.Println()
		fmt.Println("To confirm, run:")
		fmt.Printf("  %s --confirm\n", os.Args[0])
		database.Close()
		os.Exit(1)
	}

	fmt.Println("🗑️  Starting database cleanup...")
	fmt.Println("   This will delete all data from transaction tables.")
	fmt.Println("   Reference tables will be preserved.")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Список всех tx_* таблиц транзакций для очистки
	tables := make([]string, 0, len(models.TxSchemas))
	for table := range models.TxSchemas {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	// Отключаем проверку внешних ключей для ускорения
	if _, err := database.Exec(ctx, "SET session_replication_role = 'replica'"); err != nil {
		log.Printf("Warning: failed to disable foreign key checks: %v", err)
	}

	// Очищаем каждую таблицу
	clearedCount := 0
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := database.Exec(ctx, query); err != nil {
			log.Printf("Error truncating %s: %v", table, err)
			continue
		}
		clearedCount++
		fmt.Printf("  ✓ Cleared: %s\n", table)
	}

	// Включаем обратно проверку внешних ключей
	if _, err := database.Exec(ctx, "SET session_replication_role = 'origin'"); err != nil {
		log.Printf("Warning: failed to enable foreign key checks: %v", err)
	}

	fmt.Println()
	fmt.Printf("✅ Successfully cleared %d/%d tables\n", clearedCount, len(tables))

	// Выводим статистику
	fmt.Println()
	fmt.Println("📊 Database statistics:")

	stats := []struct {
		name  string
		query string
	}{
		{"Item registrations", "SELECT COUNT(*) FROM tx_item_registration_1_11"},
		{"Fiscal payments", "SELECT COUNT(*) FROM tx_fiscal_payment_40"},
		{"Document opens", "SELECT COUNT(*) FROM tx_document_open_42"},
	}

	for _, stat := range stats {
		var count int64
		if err := database.QueryRow(ctx, stat.query).Scan(&count); err != nil {
			log.Printf("Error getting count for %s: %v", stat.name, err)
			continue
		}
		fmt.Printf("  %s: %d rows\n", stat.name, count)
	}

	fmt.Println()
	fmt.Println("✅ Database cleanup completed!")
}
