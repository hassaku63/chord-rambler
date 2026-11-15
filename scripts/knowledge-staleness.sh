#!/usr/bin/env bash
# Lists the knowledge this repository keeps that may have gone stale (ADR-0007):
#   1. Entries of the practice index and the standards index past their review
#      date (確認 date + 間隔).
#   2. Source URLs that no longer open, or that now answer from another place
#      (both indexes, docs/architecture/theory-conventions.md).
#   3. Practice index entries whose recorded Git source file has changed since
#      the recorded commit.
#
# It depends on the date and the network, so it is NOT part of `make check`:
# the monthly workflow runs it and opens an issue from the report.
#
# Usage: bash scripts/knowledge-staleness.sh [--offline]
#   --offline   only the date check (no network)
# Environment:
#   TODAY=YYYY-MM-DD   override today's date (for testing)
#   GH_TOKEN           optional; raises the GitHub API rate limit
# Prints a Markdown report to stdout. Exit status: 0 nothing found, 3 findings.
set -uo pipefail

cd "$(dirname "$0")/.." || exit 1

index=.claude/skills/practices/index.md
standards=.claude/skills/standards/index.md
url_files=("$index" "$standards" docs/architecture/theory-conventions.md)
offline=0
[ "${1:-}" = "--offline" ] && offline=1

# Seconds since the epoch for a YYYY-MM-DD date, with GNU or BSD date.
epoch() { date -u -d "$1" +%s 2>/dev/null || date -u -j -f %Y-%m-%d "$1" +%s; }
today=${TODAY:-$(date -u +%Y-%m-%d)}
now=$(epoch "$today")

findings=$(mktemp)
trap 'rm -f "$findings"' EXIT
section() { printf '\n## %s\n\n' "$1" >> "$findings"; }

# Entries of an index (the practice index by default) as
# "name<TAB>field<TAB>value" lines.
entries() {
  awk '
    /^## / { entry = ""; skipped = ($0 ~ /^## 外したもの/); next }
    /^### / { entry = skipped ? "" : substr($0, 5); next }
    entry != "" && /^\| / {
      n = split($0, cell, "|")
      name = cell[2]; value = cell[3]
      gsub(/^ +| +$/, "", name); gsub(/^ +| +$/, "", value)
      if (name == "欄" || name ~ /^-+$/) next
      printf "%s\t%s\t%s\n", entry, name, value
    }
  ' "${1:-$index}"
}

# --- 1. review dates ---------------------------------------------------------
# Prints "- [label]name（…）" for each entry of the index in $1 past its date.
overdue_in() { # file label
  entries "$1" | awk -F '\t' '$2 == "確認" { print $1 "\t" $3 }' | while IFS=$'\t' read -r name value; do
    checked=${value%% *}
    days=$(printf '%s' "$value" | sed -E 's/.*間隔 ([0-9]+) 日.*/\1/')
    due=$(( $(epoch "$checked") + days * 86400 ))
    if [ "$now" -gt "$due" ]; then
      printf -- '- %s%s（確認日 %s、間隔 %s 日）\n' "$2" "$name" "$checked" "$days"
    fi
  done
}
overdue=$(overdue_in "$index" ""; overdue_in "$standards" "標準の索引: ")
if [ -n "$overdue" ]; then
  section "見直しの期限を過ぎた項目"
  printf '%s\n' "$overdue" >> "$findings"
fi

if [ "$offline" -eq 0 ]; then
  # --- 2. source URLs --------------------------------------------------------
  # A site that refuses automated requests (401, 403, 429) is reported apart:
  # it says nothing about whether the page still exists.
  broken=""; moved=""; refused=""
  while IFS= read -r url; do
    out=$(curl -sS -L -o /dev/null -m 30 -A 'Mozilla/5.0 (knowledge-staleness check)' \
      -w '%{http_code} %{url_effective}' "$url" 2>/dev/null) || out="000 $url"
    code=${out%% *}; final=${out#* }
    case "$code" in
      200)
        # Ignore a trailing slash or http->https; report a different host or path.
        a=$(printf '%s' "$url" | sed -E 's#^https?://##; s#/$##')
        b=$(printf '%s' "$final" | sed -E 's#^https?://##; s#/$##')
        [ "$a" = "$b" ] || moved+="- $url → $final"$'\n' ;;
      401|403|429) refused+="- $url（$code）"$'\n' ;;
      *) broken+="- $url（$code）"$'\n' ;;
    esac
  done < <(grep -ohE 'https?://[^ )|`、。「」]+' "${url_files[@]}" | sed -E 's/[.,]$//' | sort -u)
  if [ -n "$broken" ]; then section "開けない出どころ"; printf '%s' "$broken" >> "$findings"; fi
  if [ -n "$moved" ]; then section "別の場所へ移った出どころ"; printf '%s' "$moved" >> "$findings"; fi

  # --- 3. recorded Git sources -----------------------------------------------
  auth=()
  [ -n "${GH_TOKEN:-}" ] && auth=(-H "Authorization: Bearer $GH_TOKEN")
  changed=""; unknown=""
  while IFS=$'\t' read -r name src; do
    repo=${src%%:*}; rest=${src#*:}; path=${rest%@*}; sha=${rest##*@}
    latest=$(curl -sS -m 30 "${auth[@]}" -H 'Accept: application/vnd.github+json' \
      "https://api.github.com/repos/$repo/commits?path=$path&per_page=1" 2>/dev/null \
      | grep -m1 -oE '"sha": *"[0-9a-f]{40}"' | grep -oE '[0-9a-f]{40}')
    if [ -z "$latest" ]; then
      unknown+="- $name: \`$src\`"$'\n'
      continue
    fi
    case "$latest" in "$sha"*) continue ;; esac
    # The recorded commit is the repository head on the day of the check, so the
    # newest commit touching the file is normally its ancestor. The file changed
    # only if that commit is not contained in the recorded one.
    status=$(curl -sS -m 30 "${auth[@]}" -H 'Accept: application/vnd.github+json' \
      "https://api.github.com/repos/$repo/compare/$latest...$sha" 2>/dev/null \
      | grep -m1 -oE '"status": *"[a-z]+"' | grep -oE '(ahead|behind|identical|diverged)')
    case "$status" in
      ahead|identical) ;;
      "") unknown+="- $name: \`$src\`"$'\n' ;;
      *) changed+="- $name: \`$repo:$path\` https://github.com/$repo/compare/$sha...$latest"$'\n' ;;
    esac
  done < <(entries | awk -F '\t' '$2 == "ソース" && $3 != "なし" { n = split($3, s, ", "); for (i = 1; i <= n; i++) { gsub(/`/, "", s[i]); print $1 "\t" s[i] } }')
  if [ -n "$changed" ]; then section "記録したコミットのあとに変わったソース"; printf '%s' "$changed" >> "$findings"; fi
  if [ -n "$unknown" ]; then section "確かめられなかったソース（ファイルが移ったか、API が答えなかった）"; printf '%s' "$unknown" >> "$findings"; fi
  if [ -n "$refused" ]; then
    printf '\n## 自動の確認を拒んだサイト（参考。ページの有無は分からない）\n\n%s' "$refused" > "$findings.refused"
  fi
fi

if [ -s "$findings" ]; then
  printf '# 知識の鮮度の報告（%s）\n' "$today"
  printf '\n確かめ直す手順は %s。\n' "\`docs/dev/knowledge-review.md\`"
  cat "$findings"
  [ -f "$findings.refused" ] && cat "$findings.refused" && rm -f "$findings.refused"
  exit 3
fi
rm -f "$findings.refused"
echo "knowledge-staleness: nothing to review ($today)"
