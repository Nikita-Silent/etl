#!/bin/bash
# Простой скрипт для запуска ETL через docker-compose

set -e

DATE=${1:-$(date +%Y-%m-%d)}
MODE=${2:-webhook}

echo "🚀 Frontol ETL Runner"
echo "===================="
echo "Date: $DATE"
echo "Mode: $MODE"
echo ""

case $MODE in
  webhook)
    echo "📡 Triggering ETL via webhook..."
    curl -X POST http://localhost:$SERVER_PORT/api/load \
      -H 'Content-Type: application/json' \
      -d "{\"date\": \"$DATE\"}" \
      -w '\n'
    
    echo ""
    echo "✅ ETL triggered successfully!"
    echo "📋 Check logs: docker-compose logs -f webhook-server"
    ;;
    
  cli)
    echo "🔧 Running ETL via CLI..."
    echo ""
    
    echo "Step 1/3: Clearing FTP folders..."
    docker-compose run --rm clear-requests
    
    echo ""
    echo "Step 2/3: Waiting 60 seconds for Frontol response..."
    sleep 60
    
    echo ""
    echo "Step 3/3: Loading data for $DATE..."
    docker-compose run --rm loader ./frontol-loader $DATE
    
    echo ""
    echo "✅ ETL completed!"
    ;;
    
  *)
    echo "Usage: $0 [DATE] [MODE]"
    echo ""
    echo "Arguments:"
    echo "  DATE - Date in YYYY-MM-DD format (default: today)"
    echo "  MODE - Run mode: webhook or cli (default: webhook)"
    echo ""
    echo "Examples:"
    echo "  $0                          # Run ETL for today via webhook"
    echo "  $0 2024-12-18               # Run ETL for specific date via webhook"
    echo "  $0 2024-12-18 cli           # Run ETL via CLI"
    echo "  $0 \$(date +%Y-%m-%d) webhook  # Run ETL for today via webhook"
    exit 1
    ;;
esac
