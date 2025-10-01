#!/bin/bash
set -e

echo "Running custom setup MASTER script..."
echo "wal_level = logical" >> "$PGDATA/postgresql.conf"
echo "END of running custom setup MASTER script..."