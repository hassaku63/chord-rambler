#!/usr/bin/env bash
# Deterministic documentation checks. Part of `make check` and CI.
#
# What it enforces (see docs/dev/docs-hygiene.md for the why):
#   1. Every relative Markdown link resolves to an existing file.
#   2. Every repository path mentioned in backticks exists.
#   3. Every file under docs/ is listed in docs/README.md (ADRs via their index).
#   4. Agent instruction files stay within their line budget.
#   5. No merge conflict markers are left in any Markdown file.
#
# No dependencies beyond bash, git, grep, sed, find. Must stay portable (macOS/Linux).
set -uo pipefail

cd "$(dirname "$0")/.." || exit 1

problems=$(mktemp)
trap 'rm -f "$problems"' EXIT
problem() { printf 'docs-check: %s\n' "$*" | tee -a "$problems" >&2; }

# Markdown files we own (skip dependencies).
md_files() {
  find . -type f -name '*.md' \
    -not -path '*/node_modules/*' \
    -not -path './.git/*' | sort
}

# --- 1. relative links ------------------------------------------------------
while IFS= read -r file; do
  dir=$(dirname "$file")
  # [text](target) where target is not a URL and not a pure anchor.
  grep -oE '\]\([^)]+\)' "$file" 2>/dev/null | sed -E 's/^\]\(//; s/\)$//' | while IFS= read -r target; do
    case "$target" in
      http://*|https://*|mailto:*|'#'*) continue ;;
    esac
    path="${target%%#*}"
    [ -z "$path" ] && continue
    if [ ! -e "$dir/$path" ]; then
      problem "$file: broken link -> $target"
    fi
  done
done < <(md_files)

# --- 2. backticked repository paths ----------------------------------------
# A backticked word that starts with a top-level directory of this repository
# (`docs/...`, `scripts/...`) is a path and must exist. The directories are read
# from git, so a new top-level directory is covered without touching this script.
# Placeholders containing < > * … are skipped. Gitignored paths
# (.claude/settings.local.json) are absent from the checkout and allowed.
tops=$(git ls-files --cached --others --exclude-standard | grep / | cut -d/ -f1 | sort -u \
  | sed -E 's/[.]/\\./g' | paste -sd '|' -)
if [ -n "$tops" ]; then
  while IFS= read -r file; do
    grep -oE "\`($tops)/[^\` ]*\`" "$file" 2>/dev/null \
      | tr -d '`' | sort -u | while IFS= read -r path; do
      case "$path" in
        *'<'*|*'>'*|*'*'*|*'…'*) continue ;;
      esac
      if [ ! -e "$path" ] && ! git check-ignore -q -- "$path"; then
        problem "$file: path does not exist -> $path"
      fi
    done
  done < <(md_files)
fi

# --- 3. docs index covers every doc ------------------------------------------
index=docs/README.md
while IFS= read -r doc; do
  rel="${doc#./docs/}"
  case "$rel" in
    README.md) continue ;;
    architecture/adr/*) # ADRs are indexed by their own README
      adr_rel="${rel#architecture/adr/}"
      [ "$adr_rel" = "README.md" ] && continue
      grep -qF "($adr_rel)" docs/architecture/adr/README.md || problem "docs/architecture/adr/README.md: ADR not indexed -> $rel"
      continue ;;
  esac
  grep -qF "($rel)" "$index" || problem "$index: document not indexed -> docs/$rel"
done < <(find ./docs -type f -name '*.md' | sort)

# --- 4. instruction files stay short ---------------------------------------
budget=150
for f in AGENTS.md CLAUDE.md; do
  lines=$(wc -l < "$f" | tr -d ' ')
  if [ "$lines" -gt "$budget" ]; then
    problem "$f: $lines lines exceeds the $budget-line budget; move detail into docs/ and link it"
  fi
done

# --- 5. no leftover merge conflict markers ------------------------------------
while IFS= read -r file; do
  # A conflict always has both ends; a lone "=======" is a setext heading.
  if grep -qE '^<<<<<<< ' "$file" && grep -qE '^>>>>>>> ' "$file"; then
    problem "$file: merge conflict markers left in the file"
  fi
done < <(md_files)

if [ -s "$problems" ]; then
  echo "docs-check: FAILED ($(wc -l < "$problems" | tr -d ' ') problems)" >&2
  exit 1
fi
echo "docs-check: ok"
