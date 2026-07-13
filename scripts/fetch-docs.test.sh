#!/usr/bin/env bash
# fetch-docs.test.sh — Unit tests for lib-versions.sh (pure version selection).
#
# Run: bash scripts/fetch-docs.test.sh
# Exit code 0 = all pass, 1 = at least one failure.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib-versions.sh"

PASS=0
FAIL=0

assert_eq() {
    local desc="$1" expected="$2" actual="$3"
    if [ "$expected" = "$actual" ]; then
        PASS=$((PASS + 1))
        printf "  ✓ %s\n" "$desc"
    else
        FAIL=$((FAIL + 1))
        printf "  ✗ %s\n" "$desc"
        printf "    expected: %s\n" "$expected"
        printf "    actual:   %s\n" "$actual"
    fi
}

echo "=== extract_minor_key ==="

assert_eq "v0.0.1-106 → 0.0" \
    "0.0" "$(extract_minor_key "v0.0.1-106")"

assert_eq "v0.1.0 → 0.1" \
    "0.1" "$(extract_minor_key "v0.1.0")"

assert_eq "v1.2.3 → 1.2" \
    "1.2" "$(extract_minor_key "v1.2.3")"

assert_eq "v0.0.1-9 → 0.0" \
    "0.0" "$(extract_minor_key "v0.0.1-9")"

assert_eq "v10.20.30 → 10.20" \
    "10.20" "$(extract_minor_key "v10.20.30")"

# Unparseable tags return non-zero
if extract_minor_key "not-a-version" 2>/dev/null; then
    assert_eq "not-a-version should fail" "fail" "succeeded"
else
    PASS=$((PASS + 1))
    printf "  ✓ not-a-version correctly rejected\n"
fi

echo ""
echo "=== select_versions (single minor line) ==="

# All tags are minor 0.0 — should deduplicate to just the latest
RESULT=$(printf 'v0.0.1-99\nv0.0.1-100\nv0.0.1-106\nv0.0.1-105\n' | select_versions)
assert_eq "single minor: only latest kept" \
    "v0.0.1-106" "$RESULT"

echo ""
echo "=== select_versions (multiple minor lines) ==="

# Two minor lines: 0.0 and 0.1. Each line deduplicates to its latest.
# v0.1.0 and v0.1.1 are both minor 0.1 → only v0.1.1 (latest) kept.
RESULT=$(printf 'v0.0.1-106\nv0.1.0\nv0.0.1-99\nv0.1.1\n' | select_versions)
EXPECTED="v0.1.1
v0.0.1-106"
assert_eq "multiple minors: one per line, newest-first" \
    "$EXPECTED" "$RESULT"

echo ""
echo "=== select_versions (numeric prerelease sort) ==="

# Ensure 0.0.1-100 > 0.0.1-9 (not string sort)
RESULT=$(printf 'v0.0.1-9\nv0.0.1-100\nv0.0.1-50\n' | select_versions)
assert_eq "numeric prerelease: 100 > 9" \
    "v0.0.1-100" "$RESULT"

echo ""
echo "=== select_versions (three minor lines) ==="

RESULT=$(printf 'v0.0.5\nv1.0.0\nv0.1.3\nv0.0.1-106\nv1.1.0\nv0.1.0\n' | select_versions)
EXPECTED="v1.1.0
v1.0.0
v0.1.3
v0.0.5"
assert_eq "three minor lines, correct dedup + sort" \
    "$EXPECTED" "$RESULT"

echo ""
echo "=== write_manifest ==="

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

write_manifest "$TMPDIR/manifest.json" \
    "v0.0.1-106" "0.0" "2026-07-09T11:18:15Z" \
    -- next

MANIFEST=$(cat "$TMPDIR/manifest.json")

# Verify latest field
LATEST=$(grep '"latest"' "$TMPDIR/manifest.json" | cut -d'"' -f4)
assert_eq "manifest latest field" "v0.0.1-106" "$LATEST"

# Verify next field (the top-level "next" key, not the version entry)
NEXT=$(grep '  "next":' "$TMPDIR/manifest.json" | cut -d'"' -f4)
assert_eq "manifest next field" "next" "$NEXT"

# Verify version entry exists
if grep -q '"version": "v0.0.1-106"' "$TMPDIR/manifest.json"; then
    PASS=$((PASS + 1))
    printf "  ✓ manifest contains version entry\n"
else
    FAIL=$((FAIL + 1))
    printf "  ✗ manifest missing version entry\n"
fi

# Verify next entry exists
if grep -q '"isNext": true' "$TMPDIR/manifest.json"; then
    PASS=$((PASS + 1))
    printf "  ✓ manifest contains next entry\n"
else
    FAIL=$((FAIL + 1))
    printf "  ✗ manifest missing next entry\n"
fi

# Verify valid JSON (if jq available)
if command -v jq &>/dev/null; then
    if jq empty "$TMPDIR/manifest.json" 2>/dev/null; then
        PASS=$((PASS + 1))
        printf "  ✓ manifest is valid JSON\n"
    else
        FAIL=$((FAIL + 1))
        printf "  ✗ manifest is invalid JSON\n"
        cat "$TMPDIR/manifest.json"
    fi
fi

echo ""
echo "=== write_manifest (multiple versions) ==="

write_manifest "$TMPDIR/manifest2.json" \
    "v1.1.0" "1.1" "2026-08-01T10:00:00Z" \
    "v1.0.0" "1.0" "2026-07-15T10:00:00Z" \
    "v0.1.3" "0.1" "2026-07-10T10:00:00Z" \
    -- next

LATEST2=$(grep '"latest"' "$TMPDIR/manifest2.json" | cut -d'"' -f4)
assert_eq "multi-version latest" "v1.1.0" "$LATEST2"

VERSION_COUNT=$(grep -c '"version":' "$TMPDIR/manifest2.json")
assert_eq "multi-version entry count (3 releases + 1 next = 4)" "4" "$VERSION_COUNT"

echo ""
echo "================================"
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
