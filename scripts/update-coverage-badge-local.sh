#!/usr/bin/env bash
set -eu

# Runs tests, generates coverage, extracts percentage, and updates README badge line locally.
# Usage: ./scripts/update-coverage-badge-local.sh

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "Running tests and generating coverage profile..."
go test -v ./... -covermode=count -coverprofile=coverage.out

# Extract coverage percentage from go tool cover -func output
cov_line=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}')
# cov_line looks like 83.3%
if [[ -z "$cov_line" ]]; then
  echo "Could not determine coverage percentage"
  exit 1
fi

# Remove trailing % and format for badge
cov_pct=${cov_line%%%}
# Round to one decimal if needed
cov_display="$cov_pct%"

# Build shields.io badge URL
badge_url="https://img.shields.io/badge/Coverage-${cov_display}-brightgreen"
if (( $(echo "$cov_pct < 70" | bc -l) )); then
  badge_url="https://img.shields.io/badge/Coverage-${cov_display}-yellow"
fi
if (( $(echo "$cov_pct < 30" | bc -l) )); then
  badge_url="https://img.shields.io/badge/Coverage-${cov_display}-red"
fi

# Update README: replace line that contains '![Coverage](' with new badge URL
if grep -q "!\[Coverage\]" README.md; then
  sed -i "s|!\[Coverage\](.*)|![Coverage](${badge_url})|" README.md
  echo "Updated README.md with badge: ${badge_url}"
else
  # Insert after the title line (first H1)
  awk -v b="![Coverage](${badge_url})" 'NR==1{print; print b; next}1' README.md > README.md.tmp && mv README.md.tmp README.md
  echo "Inserted badge into README.md: ${badge_url}"
fi

echo "Done. Coverage: ${cov_display}"