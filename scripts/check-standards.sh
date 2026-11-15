#!/usr/bin/env bash
# Checks the shape of the standards index (.claude/skills/standards/index.md).
# Part of `make check`. It looks only at the file, never at the date or the
# network, so the same commit always gives the same result (ADR-0015).
#
# What it enforces:
#   1. Every entry (a "###" heading) has all seven fields, none of them empty.
#   2. 一次情報 holds a URL; 確認 is "YYYY-MM-DD / 間隔 N 日".
#
# Needs bash and awk. The awk patterns avoid interval expressions, which older
# mawk does not support.
set -uo pipefail

cd "$(dirname "$0")/.." || exit 1

index=.claude/skills/standards/index.md

if [ ! -f "$index" ]; then
  echo "standards-check: $index not found" >&2
  exit 1
fi

awk '
  function report(msg) { printf "standards-check: %s\n", msg > "/dev/stderr"; bad++ }
  function close_entry(   i, n, names) {
    if (entry == "") return
    n = split("発行元|何を決めているか|一次情報|状態|当たるべき場面|chord-rambler との接点|確認", names, "|")
    for (i = 1; i <= n; i++) if (!(names[i] in seen)) report(entry ": missing field -> " names[i])
    delete seen
    entry = ""
  }
  /^## / { close_entry(); next }
  /^### / { close_entry(); entry = substr($0, 5); count++; next }
  entry != "" && /^\| / {
    n = split($0, cell, "|")
    name = cell[2]; value = cell[3]
    gsub(/^ +| +$/, "", name); gsub(/^ +| +$/, "", value)
    if (name == "欄" || name ~ /^-+$/) next
    seen[name] = 1
    if (value == "") report(entry ": empty field -> " name)
    if (name == "一次情報" && value !~ /https?:\/\//) report(entry ": 一次情報 has no URL")
    if (name == "確認" && value !~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] \/ 間隔 [0-9]+ 日$/)
      report(entry ": 確認 must be YYYY-MM-DD / 間隔 N 日 -> " value)
  }
  END {
    close_entry()
    if (count == 0) report("no entries")
    if (bad > 0) { print "standards-check: FAILED" > "/dev/stderr"; exit 1 }
    printf "standards-check: ok (%d entries)\n", count
  }
' "$index"
