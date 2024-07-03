#!/bin/bash
# Set the path to your database
DB_PATH=$(pwd)/data/forum.db

# Export the environment variable for the current shell
export FORUM_DB_PATH="$DB_PATH"

# Confirm setting
echo "FORUM_DB_PATH set to $FORUM_DB_PATH"

exec go run ./cmd/server/main.go
