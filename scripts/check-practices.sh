#!/usr/bin/env bash
# Checks the shape of the practice index (.claude/skills/practices/index.md).
# Part of `make check`. It looks only at the file, never at the date or the
# network, so the same commit always gives the same result (ADR-0007).
#
# What it enforces:
#   1. Every entry (a "###" heading outside "外したもの") has all nine fields.
#   2. 出どころ holds a URL; ソース, 対象 and 確認 follow their fixed forms.
#   3. The index holds at most 40 entries.
#
# No dependencies beyond bash and awk. Must stay portable (macOS/Linux), so the
# patterns avoid interval expressions, which older mawk does not support.
set -uo pipefail

cd "$(dirname "$0")/.." || exit 1

index=.claude/skills/practices/index.md
budget=40

if [ ! -f "$index" ]; then
  echo "practices-check: $index not found" >&2
  exit 1
fi

problems=$(awk -v budget="$budget" '
  function report(msg) { printf "practices-check: %s\n", msg; bad++ }
  function close_entry(   i, n, names) {
    if (entry == "") return
    n = split("出どころ ソース 何を言っているか 確からしさ 当たるべき場面 判断の問い このプロジェクトの立場 対象 確認", names, " ")
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
    name = cell[2]; value = cell[3]
    gsub(/^ +| +$/, "", name); gsub(/^ +| +$/, "", value)
    if (name == "欄" || name ~ /^-+$/) next
    seen[name] = 1
    if (value == "") report(entry ": empty field -> " name)
    if (name == "出どころ" && value !~ /https?:\/\//) report(entry ": 出どころ has no URL")
    if (name == "ソース" && value != "なし" && value !~ /^`[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+:[^@`]+@[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+`$/)
      report(entry ": ソース must be `owner/repo:path@commit` or なし -> " value)
    if (name == "対象" && value != "なし" && value !~ /^`[^`@]+@[0-9]+`(, *`[^`@]+@[0-9]+`)*$/)
      report(entry ": 対象 must be `name@major` (comma separated) or なし -> " value)
    if (name == "確認" && value !~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] \/ 間隔 [0-9]+ 日$/)
      report(entry ": 確認 must be YYYY-MM-DD / 間隔 N 日 -> " value)
  }
  END {
    close_entry()
    if (count > budget) report(count " entries exceed the budget of " budget "; remove some before adding")
    exit bad > 0
  }
' "$index")
status=$?

if [ "$status" -ne 0 ]; then
  printf '%s\n' "$problems" >&2
  echo "practices-check: FAILED" >&2
  exit 1
fi
echo "practices-check: ok"
