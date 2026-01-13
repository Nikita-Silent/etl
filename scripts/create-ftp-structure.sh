#!/bin/sh
# Script to create FTP folder structure based on KASSA_STRUCTURE
# This script is run by the ftp-structure-init container

set -e

# Use absolute paths in FTP user home directory
FTP_USER_NAME="${FTP_USER:-frontol}"
FTP_USER_HOME="${FTP_USER_HOME:-/home/ftp/${FTP_USER_NAME}}"
FTP_REQUEST_BASE="${FTP_REQUEST_DIR:-/request}"
FTP_RESPONSE_BASE="${FTP_RESPONSE_DIR:-/response}"
KASSA_STRUCTURE="${KASSA_STRUCTURE:-P13:P13;N22:N22_Inter,N22_FURN;SH54:SH54;S6:S6;L98:L98;L32:L32;S39:S39;O49:O49;L28:L28}"

# FTP user and group
FTP_USER="${FTP_USER:-frontol}"
FTP_GROUP="${FTP_GROUP:-${FTP_USER}}"
# Use UID/GID 1000:1000 to match FTP server user
FTP_UID="${FTP_UID:-1000}"
FTP_GID="${FTP_GID:-1000}"

# Convert relative paths to absolute
if [ "${FTP_REQUEST_BASE#/}" != "$FTP_REQUEST_BASE" ]; then
    FTP_REQUEST_DIR="$FTP_USER_HOME$FTP_REQUEST_BASE"
else
    FTP_REQUEST_DIR="$FTP_USER_HOME/$FTP_REQUEST_BASE"
fi

if [ "${FTP_RESPONSE_BASE#/}" != "$FTP_RESPONSE_BASE" ]; then
    FTP_RESPONSE_DIR="$FTP_USER_HOME$FTP_RESPONSE_BASE"
else
    FTP_RESPONSE_DIR="$FTP_USER_HOME/$FTP_RESPONSE_BASE"
fi

echo "=========================================="
echo "FTP Structure Init Container"
echo "=========================================="
echo "FTP_USER_HOME: $FTP_USER_HOME"
echo "FTP_REQUEST_DIR: $FTP_REQUEST_DIR"
echo "FTP_RESPONSE_DIR: $FTP_RESPONSE_DIR"
echo "KASSA_STRUCTURE: $KASSA_STRUCTURE"
echo ""

# Create base directories (idempotent operation)
echo "Checking base directories..."
if [ ! -d "$FTP_REQUEST_DIR" ]; then
    mkdir -p "$FTP_REQUEST_DIR"
    chown "$FTP_UID:$FTP_GID" "$FTP_REQUEST_DIR"
    chmod 775 "$FTP_REQUEST_DIR"
    echo "✓ Created request directory: $FTP_REQUEST_DIR (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
else
    chown "$FTP_UID:$FTP_GID" "$FTP_REQUEST_DIR"
    chmod 775 "$FTP_REQUEST_DIR"
    echo "✓ Request directory exists: $FTP_REQUEST_DIR (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
fi

if [ ! -d "$FTP_RESPONSE_DIR" ]; then
    mkdir -p "$FTP_RESPONSE_DIR"
    chown "$FTP_UID:$FTP_GID" "$FTP_RESPONSE_DIR"
    chmod 775 "$FTP_RESPONSE_DIR"
    echo "✓ Created response directory: $FTP_RESPONSE_DIR (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
else
    chown "$FTP_UID:$FTP_GID" "$FTP_RESPONSE_DIR"
    chmod 775 "$FTP_RESPONSE_DIR"
    echo "✓ Response directory exists: $FTP_RESPONSE_DIR (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
fi
echo ""

# Parse KASSA_STRUCTURE and create folders
# Format: "KASSA_CODE:FOLDER1,FOLDER2;KASSA_CODE2:FOLDER1"
IFS=';'
for kassa_group in $KASSA_STRUCTURE; do
    # Extract kassa code and folders
    kassa_code=$(echo "$kassa_group" | cut -d':' -f1 | tr -d ' ')
    folders=$(echo "$kassa_group" | cut -d':' -f2)
    
    if [ -z "$kassa_code" ] || [ -z "$folders" ]; then
        continue
    fi
    
    echo "Processing kassa: $kassa_code"
    
    # Create folders for each folder name
    IFS=','
    for folder_name in $folders; do
        folder_name=$(echo "$folder_name" | tr -d ' ')
        if [ -n "$folder_name" ]; then
            request_path="$FTP_REQUEST_DIR/$kassa_code/$folder_name"
            response_path="$FTP_RESPONSE_DIR/$kassa_code/$folder_name"
            
            # Create directories (mkdir -p is idempotent - won't fail if exists)
            if [ ! -d "$request_path" ]; then
                mkdir -p "$request_path"
                chown "$FTP_UID:$FTP_GID" "$request_path"
                chmod 775 "$request_path"
                echo "  Created: $request_path (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
            else
                chown "$FTP_UID:$FTP_GID" "$request_path"
                chmod 775 "$request_path"
                echo "  Exists:  $request_path (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
            fi
            
            if [ ! -d "$response_path" ]; then
                mkdir -p "$response_path"
                chown "$FTP_UID:$FTP_GID" "$response_path"
                chmod 775 "$response_path"
                echo "  Created: $response_path (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
            else
                chown "$FTP_UID:$FTP_GID" "$response_path"
                chmod 775 "$response_path"
                echo "  Exists:  $response_path (owner: $FTP_OWNER_USER:$FTP_OWNER_GROUP)"
            fi
        fi
    done
    IFS=';'
done

# Fix ownership and permissions recursively for all created directories
echo ""
echo "Setting correct ownership and permissions..."
# Set ownership first
chown -R "$FTP_UID:$FTP_GID" "$FTP_USER_HOME"
# Set permissions: 775 for directories (rwx for owner/group, rx for others)
# This allows owner and group to read, write, and execute (list/create files)
find "$FTP_USER_HOME" -type d -exec chmod 775 {} \;
# Set permissions for files (if any exist): 664 (rw for owner/group, r for others)
find "$FTP_USER_HOME" -type f -exec chmod 664 {} \;
echo "✓ Ownership and permissions set recursively"

echo ""
echo "FTP folder structure created successfully!"

