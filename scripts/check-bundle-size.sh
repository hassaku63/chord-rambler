#!/usr/bin/env bash
# Budget for what the screen loads up front: the gzip size of frontend/dist/index.html
# plus every <script src> and <link rel="stylesheet"> it references, summed.
# Part of `make check`, after the build. Uses only bash, grep, sed, gzip and wc.
#
# The budget lives here and only here. BUNDLE_BUDGET_BYTES overrides it for one run
# (to see the check fail, or to try a value); do not set it in CI.
#
# Why this number: no entry in the practice index covers it, and web.dev (the
# performance budgets page) could not be opened when it was chosen. So it is the
# size at the time (about 78.7 KB gzip) with about 30% headroom. Lower it when
# the screen gets smaller; raise it only on purpose, in a PR that says why.
set -uo pipefail

cd "$(dirname "$0")/.." || exit 1

BUDGET_BYTES=${BUNDLE_BUDGET_BYTES:-102400} # 100 KiB
dist=${1:-frontend/dist}
html="$dist/index.html"

if [ ! -f "$html" ]; then
  echo "bundle-size: $html not found (run the build first: make frontend-build)" >&2
  exit 1
fi

gz() { gzip -9 -c -- "$1" | wc -c | tr -d ' '; }

# Paths in index.html are absolute from the site root ("/assets/x.js").
refs=$( {
  grep -oE '<script[^>]*[[:space:]]src="[^"]+"' "$html" | sed -E 's/.*[[:space:]]src="([^"]+)"/\1/'
  grep -oE '<link[^>]*rel="stylesheet"[^>]*>' "$html" | grep -oE 'href="[^"]+"' | sed -E 's/href="([^"]+)"/\1/'
} | sort -u)

total=$(gz "$html")
report=$(printf '  %8d  %s\n' "$total" "$html")
for ref in $refs; do
  case "$ref" in
    http://* | https://* | //*) continue ;; # not ours; it is not in dist
  esac
  file="$dist/${ref#/}"
  if [ ! -f "$file" ]; then
    echo "bundle-size: $html refers to $ref but $file does not exist" >&2
    exit 1
  fi
  size=$(gz "$file")
  total=$((total + size))
  report+=$'\n'$(printf '  %8d  %s' "$size" "$file")
done

if [ "$total" -gt "$BUDGET_BYTES" ]; then
  echo "bundle-size: FAILED: ${total} bytes gzip loaded up front, budget ${BUDGET_BYTES} bytes (over by $((total - BUDGET_BYTES)))" >&2
  echo "$report" | sort -rn >&2
  exit 1
fi
printf 'bundle-size: ok (%d of %d bytes gzip)\n' "$total" "$BUDGET_BYTES"
