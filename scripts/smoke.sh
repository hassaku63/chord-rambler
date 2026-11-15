#!/usr/bin/env bash
# Smoke test of a running stack: the web container and the api behind it
# (ADR-0010). `make up` starts one; CI runs this against it.
#
# Usage: scripts/smoke.sh <base URL> <expected version>
#   e.g. scripts/smoke.sh http://localhost:8080 "$(git describe --tags --always --dirty)"
#
# Checks:
#   1. GET /api/v1/health answers 200 with the expected version (it waits for
#      the api to come up, a fixed number of tries).
#   2. GET / answers 200 with the screen's HTML.
#   3. A file that does not exist answers 404 (no fallback to the page).
#   4. An API path that does not exist answers 404 with a problem document
#      from the api, not nginx's own page.
set -uo pipefail

base="${1:?base URL, e.g. http://localhost:8080}"
want_version="${2:?expected version}"
tries="${SMOKE_TRIES:-30}"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
failures=0
fail() { printf 'smoke: FAIL %s\n' "$*" >&2; failures=$((failures + 1)); }
pass() { printf 'smoke: ok   %s\n' "$*"; }

# fetch <path>: writes the body to $tmp/body and prints "status content-type".
fetch() {
  curl -sS -m 10 -o "$tmp/body" -w '%{http_code} %{content_type}' "$base$1" 2>/dev/null || echo "000 -"
}

# --- 1. health, waiting for the api -----------------------------------------
for i in $(seq 1 "$tries"); do
  read -r status ctype <<<"$(fetch /api/v1/health)"
  [ "$status" = 200 ] && break
  [ "$i" -lt "$tries" ] && sleep 1
done
if [ "$status" != 200 ]; then
  fail "GET /api/v1/health answered $status after $tries tries"
elif ! grep -q "\"version\":\"$want_version\"" "$tmp/body"; then
  fail "GET /api/v1/health reports another version: $(cat "$tmp/body") (want $want_version)"
else
  pass "GET /api/v1/health is 200 with version $want_version"
fi

# --- 2. the screen -----------------------------------------------------------
read -r status ctype <<<"$(fetch /)"
if [ "$status" = 200 ] && [[ "$ctype" == text/html* ]] && grep -q 'id="root"' "$tmp/body"; then
  pass "GET / is the screen's HTML"
else
  fail "GET / answered $status ($ctype) without the screen's root element"
fi

# --- 3. a missing file ---------------------------------------------------------
read -r status ctype <<<"$(fetch /assets/no-such-file.js)"
if [ "$status" = 404 ]; then
  pass "a missing file is 404"
else
  fail "GET /assets/no-such-file.js answered $status, want 404"
fi

# --- 4. a missing API path -----------------------------------------------------
read -r status ctype <<<"$(fetch /api/v1/no-such-path)"
if [ "$status" = 404 ] && [[ "$ctype" == application/problem+json* ]]; then
  pass "a missing API path is the api's 404 problem document"
else
  fail "GET /api/v1/no-such-path answered $status ($ctype), want 404 application/problem+json"
fi

if [ "$failures" -gt 0 ]; then
  echo "smoke: FAILED ($failures)" >&2
  exit 1
fi
echo "smoke: ok"
