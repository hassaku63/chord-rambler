#!/usr/bin/env bash
# Checks the shape of the HCI index (.claude/skills/hci/index.md).
# Part of `make check`. It looks only at the file, never at the date or the
# network, so the same commit always gives the same result (ADR-0018).
#
# What it enforces:
#   1. Every entry (a "###" heading outside "外したもの") has all eight fields,
#      none of them empty.
#   2. 出典 holds a URL; 確認日 is "YYYY-MM-DD / 間隔 N 日（what was opened）"
#      or "未確認（reason）".
#   3. The index holds at most 40 entries.
#
# Needs bash and awk. The awk patterns avoid interval expressions, which older
# mawk does not support.
set -uo pipefail

cd "$(dirname "$0")/.." || exit 1

index=.claude/skills/hci/index.md
budget=40

if [ ! -f "$index" ]; then
  echo "hci-check: $index not found" >&2
  exit 1
fi

awk -v budget="$budget" '
  function report(msg) { printf "hci-check: %s\n", msg > "/dev/stderr"; bad++ }
  function close_entry(   i, n, names) {
    if (entry == "") return
    n = split("出典|何を言っているか|確からしさ|当たるべき場面|判断の問い|適切・不適切な例|chord-rambler との接点|確認日", names, "|")
    for (i = 1; i <= n; i++) if (!(names[i] in seen)) report(entry ": missing field -> " names[i])
    delete seen
    entry = ""
  }
  /^## / { close_entry(); skipped = ($0 ~ /^## 外したもの/); next }
  /^### / {
    close_entry()
    if (!skipped) { entry = substr($0, 5); count++ }
    next
  }
  entry != "" && /^\| / {
    n = split($0, cell, "|")
    if (n != 4) { report(entry ": a row must have two cells (no | inside a cell) -> " $0); next }
    name = cell[2]; value = cell[3]
    gsub(/^ +| +$/, "", name); gsub(/^ +| +$/, "", value)
    if (name == "欄" || name ~ /^-+$/) next
    seen[name] = 1
    if (value == "") report(entry ": empty field -> " name)
    if (name == "出典" && value !~ /https?:\/\//) report(entry ": 出典 has no URL")
    if (name == "確認日" && value !~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] \/ 間隔 [0-9]+ 日（.+）$/ && value !~ /^未確認（.+）$/)
      report(entry ": 確認日 must be YYYY-MM-DD / 間隔 N 日（開いたもの） or 未確認（理由） -> " value)
  }
  END {
    close_entry()
    if (count == 0) report("no entries")
    if (count > budget) report(count " entries exceed the budget of " budget "; remove some before adding")
    if (bad > 0) { print "hci-check: FAILED" > "/dev/stderr"; exit 1 }
    printf "hci-check: ok (%d entries)\n", count
  }
' "$index"
