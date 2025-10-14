#!/usr/bin/env bash
set -eu

# Runs tests, generates coverage, extracts percentage, and updates README badge line locally.
# Usage: ./scripts/update-coverage-badge-local.sh

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "Running tests and generating coverage profile..."
go test -v ./... -covermode=count -coverprofile=coverage.out

# If README uses Codecov badge, do not modify it locally
if grep -q "codecov.io/gh/christk1/kstack/graph/badge.svg" README.md; then
  echo "README uses Codecov dynamic badge. Skipping badge URL update."
  exit 0
fi

# Extract coverage percentage from go tool cover -func output
cov_line=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}')
# cov_line looks like 83.3%
if [[ -z "$cov_line" ]]; then
  echo "Could not determine coverage percentage"
  exit 1
fi

# Remove trailing % and compute displays
cov_pct=${cov_line%%%}
# Unencoded (for logs only) and encoded (for shields.io URL)
cov_display="$cov_pct%"
cov_display_enc="${cov_pct}%25"

# Build shields.io badge URL
badge_url="https://img.shields.io/badge/Coverage-${cov_display_enc}-brightgreen"
if (( $(echo "$cov_pct < 70" | bc -l) )); then
  badge_url="https://img.shields.io/badge/Coverage-${cov_display_enc}-yellow"
fi
if (( $(echo "$cov_pct < 30" | bc -l) )); then
  badge_url="https://img.shields.io/badge/Coverage-${cov_display_enc}-red"
fi

# Update README badge
# Supports two formats:
# 1) Markdown: ![Coverage](...)
# 2) HTML badges block: <img alt="Coverage" src="..." /> inside a centered <p>

if grep -qF '![Coverage]' README.md; then
  # Replace Markdown badge using sed with safe delimiter
  sed -i -E "s|!\\[Coverage\\]\\([^)]*\\)|![Coverage](${badge_url})|g" README.md
  echo "Updated README.md Markdown badge: ${badge_url}"
elif grep -qF 'alt="Coverage"' README.md; then
  # Replace HTML img src where alt="Coverage" using sed with capture groups
  sed -i -E 's|(img[^>]*alt="Coverage"[^>]*src=")[^"]*(")|\1'"${badge_url}"'\2|g' README.md
  echo "Updated README.md HTML badge: ${badge_url}"
else
  # Insert a Markdown badge under the title as fallback
  awk -v b="![Coverage](${badge_url})" 'NR==1{print; print b; next}1' README.md > README.md.tmp && mv README.md.tmp README.md
  echo "Inserted new Markdown badge into README.md: ${badge_url}"
fi

echo "Done. Coverage: ${cov_display}"