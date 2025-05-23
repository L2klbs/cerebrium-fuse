#!/bin/bash
# test/populate.sh — Reads all files under /mnt/all-projects to populate up SSD cache

set -e

echo "📂 Populate SSD cache by reading all files under /mnt/all-projects..."

find /mnt/all-projects -type f -name "*.go" | while read -r file; do
    echo "🔍 Reading: $file"
    cat "$file" > /dev/null
done

echo "✅ Done warming up cache."
