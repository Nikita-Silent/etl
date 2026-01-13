#!/bin/bash
# Script to fix FTP passive mode PUBLICHOST configuration

set -e

ENV_FILE="${1:-.env}"

echo "=========================================="
echo "FTP Passive Mode Configuration Fix"
echo "=========================================="
echo ""

# Get host IP addresses
HOST_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
if [ -z "$HOST_IP" ]; then
    HOST_IP=$(ip addr show | grep "inet " | grep -v "127.0.0.1" | head -1 | awk '{print $2}' | cut -d'/' -f1)
fi

echo "Detected host IP: $HOST_IP"
echo ""

# Check if .env exists
if [ ! -f "$ENV_FILE" ]; then
    echo "Creating $ENV_FILE from env.example..."
    cp env.example "$ENV_FILE"
fi

# Check current PUBLICHOST
CURRENT_PUBLICHOST=$(grep "^PUBLICHOST=" "$ENV_FILE" 2>/dev/null | cut -d'=' -f2 | tr -d '"' | tr -d "'" || echo "")

if [ -n "$CURRENT_PUBLICHOST" ] && [ "$CURRENT_PUBLICHOST" != "localhost" ]; then
    echo "Current PUBLICHOST: $CURRENT_PUBLICHOST"
    echo ""
    read -p "Do you want to change it? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping current PUBLICHOST: $CURRENT_PUBLICHOST"
        exit 0
    fi
fi

echo "Please select PUBLICHOST value:"
echo "1) localhost (for local connections only)"
echo "2) $HOST_IP (detected host IP)"
echo "3) 192.168.65.1 (Docker Desktop/WSL2 gateway)"
echo "4) Enter custom IP/hostname"
echo ""
read -p "Choice [1-4]: " choice

case $choice in
    1)
        NEW_PUBLICHOST="localhost"
        ;;
    2)
        NEW_PUBLICHOST="$HOST_IP"
        ;;
    3)
        NEW_PUBLICHOST="192.168.65.1"
        ;;
    4)
        read -p "Enter IP/hostname: " NEW_PUBLICHOST
        ;;
    *)
        echo "Invalid choice. Using detected host IP: $HOST_IP"
        NEW_PUBLICHOST="$HOST_IP"
        ;;
esac

# Update .env file
if grep -q "^PUBLICHOST=" "$ENV_FILE"; then
    # Replace existing PUBLICHOST
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|^PUBLICHOST=.*|PUBLICHOST=$NEW_PUBLICHOST|" "$ENV_FILE"
    else
        # Linux
        sed -i "s|^PUBLICHOST=.*|PUBLICHOST=$NEW_PUBLICHOST|" "$ENV_FILE"
    fi
else
    # Add PUBLICHOST if not exists
    echo "PUBLICHOST=$NEW_PUBLICHOST" >> "$ENV_FILE"
fi

echo ""
echo "✓ Updated PUBLICHOST to: $NEW_PUBLICHOST"
echo ""
echo "Next steps:"
echo "1. Restart FTP server: docker compose restart ftp-server"
echo "2. Test connection from Frontol"
echo ""

