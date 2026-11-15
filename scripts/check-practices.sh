#!/usr/bin/env bash
# Checks the shape of the practice index (.claude/skills/practices/index.md).
# Part of `make check`. It looks only at the file, never at the date or the
# network, so the same commit always gives the same result (ADR-0007).
#
# What it enforces:
#   1. Every entry (a "###" heading outside "外したもの") has all nine fields.
#   2. 出どころ holds a URL; ソース, 対象 and 確認 follow their fixed forms.
#   3. The index holds at most 40 entries.
#   4. Every 対象 matches the version in use: an npm package's major in
#      frontend/package-lock.json, go's major.minor and node's major in
#      .tool-versions. A name that is none of these fails; it is never skipped.
#   5. Every direct dependency of frontend/package.json, and go and node, is the
#      対象 of an entry or a row of 対象外の依存 (and not both).
#
# It reads only committed files (never node_modules), so the same commit always
# gives the same result. Needs bash, awk and node (node reads the JSON files;
# the frontend needs it anyway). The awk patterns avoid interval expressions,
# which older mawk does not support.
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
  /^## / {
    close_entry()
    skipped = ($0 ~ /^## 外したもの/ || $0 ~ /^## 対象外の依存/)
    exempt = ($0 ~ /^## 対象外の依存/)
    next
  }
  exempt && /^\| `/ {
    n = split($0, cell, "|")
    dep = cell[2]; kind = cell[3]; reason = cell[4]
    gsub(/[` ]/, "", dep); gsub(/^ +| +$/, "", kind); gsub(/^ +| +$/, "", reason)
    if (kind != "要らない" && kind != "まだない") report("対象外の依存 " dep ": 種類 must be 要らない or まだない -> " kind)
    if (reason == "") report("対象外の依存 " dep ": no reason")
    printf "EXEMPT\t%s\n", dep
    next
  }
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
    if (name == "ソース" && value != "なし" && value !~ /^`[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+:[^@`]+@[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+`(, `[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+:[^@`]+@[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+`)*$/)
      report(entry ": ソース must be `owner/repo:path@commit` (comma separated) or なし -> " value)
    if (name == "対象" && value != "なし" && value !~ /^`@?[^`@]+@[0-9][0-9.]*`(, *`@?[^`@]+@[0-9][0-9.]*`)*$/)
      report(entry ": 対象 must be `name@version` (comma separated) or なし -> " value)
    if (name == "対象" && value != "なし") {
      n = split(value, targets, ",")
      for (i = 1; i <= n; i++) { t = targets[i]; gsub(/[` ]/, "", t); printf "TARGET\t%s\t%s\n", entry, t }
    }
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

# --- 4 and 5. versions in use and coverage of the dependencies ---------------
if ! command -v node >/dev/null 2>&1; then
  echo "practices-check: node is needed to read frontend/package.json and package-lock.json" >&2
  exit 1
fi

tool() { awk -v t="$1" '$1 == t { print $2 }' .tool-versions; }
go_version=$(tool golang | cut -d. -f1-2)
node_version=$(tool nodejs | cut -d. -f1)

# "name<TAB>major" for every direct dependency, from the lockfile's versions.
npm_versions=$(node -e '
  const pkg = require("./frontend/package.json");
  const lock = require("./frontend/package-lock.json");
  for (const name of Object.keys({ ...pkg.dependencies, ...pkg.devDependencies }).sort()) {
    const entry = lock.packages["node_modules/" + name];
    console.log(name + "\t" + (entry ? entry.version.split(".")[0] : "missing"));
  }
') || { echo "practices-check: could not read frontend/package.json or package-lock.json" >&2; exit 1; }

known="$(printf 'go\t%s\nnode\t%s\n%s\n' "$go_version" "$node_version" "$npm_versions")"
targets=$(printf '%s\n' "$problems" | awk -F '\t' '$1 == "TARGET" { print $2 "\t" $3 }')
exempts=$(printf '%s\n' "$problems" | awk -F '\t' '$1 == "EXEMPT" { print $2 }')
problems=$(printf '%s\n' "$problems" | grep -v -e '^TARGET' -e '^EXEMPT' | sed '/^$/d')

version_of() { printf '%s\n' "$known" | awk -F '\t' -v n="$1" '$1 == n { print $2 }'; }
add() { problems="${problems:+$problems$'\n'}practices-check: $*"; status=1; }

while IFS=$'\t' read -r entry target; do
  [ -z "$target" ] && continue
  name=${target%@*}; want=${target##*@}
  have=$(version_of "$name")
  if [ -z "$have" ]; then
    add "$entry: 対象 $name is not a dependency of frontend/package.json, nor go or node"
  elif [ "$name" = go ] && ! printf '%s' "$want" | grep -Eq '^[0-9]+\.[0-9]+$'; then
    add "$entry: 対象 go must be major.minor (go@$go_version) -> $target"
  elif [ "$name" != go ] && ! printf '%s' "$want" | grep -Eq '^[0-9]+$'; then
    add "$entry: 対象 $name must be a major version ($name@$have) -> $target"
  elif [ "$want" != "$have" ]; then
    add "$entry: 対象 $target, but $name@$have is in use; open the source again and update the entry (docs/dev/knowledge-review.md)"
  fi
done <<<"$targets"

while IFS= read -r dep; do
  [ -z "$dep" ] && continue
  [ -n "$(version_of "$dep")" ] || add "対象外の依存 $dep is not a dependency of frontend/package.json, nor go or node"
done <<<"$exempts"

covered=$(printf '%s\n' "$targets" | awk -F '\t' 'NF { sub(/@[^@]*$/, "", $2); print $2 }' | sort -u)
while IFS=$'\t' read -r name _; do
  [ -z "$name" ] && continue
  in_entry=$(printf '%s\n' "$covered" | grep -Fxc -- "$name")
  in_exempt=$(printf '%s\n' "$exempts" | grep -Fxc -- "$name")
  if [ "$in_entry" -eq 0 ] && [ "$in_exempt" -eq 0 ]; then
    add "$name has no entry and is not in 対象外の依存; add one or the other"
  elif [ "$in_entry" -gt 0 ] && [ "$in_exempt" -gt 0 ]; then
    add "$name has an entry and is also in 対象外の依存; remove the row"
  fi
done <<<"$known"

if [ "$status" -ne 0 ]; then
  printf '%s\n' "$problems" >&2
  echo "practices-check: FAILED" >&2
  exit 1
fi
echo "practices-check: ok"
