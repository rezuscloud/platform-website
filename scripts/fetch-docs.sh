#!/usr/bin/env bash
# fetch-docs.sh - Fetch documentation from configured repositories into docs/external/
#
# Uses `gh api` for authenticated downloads (works for private repos).
# Falls back to unauthenticated git clone for public repos.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="$SCRIPT_DIR/../docs/external"

# Ensure clean state
rm -rf "$TARGET_DIR"
mkdir -p "$TARGET_DIR"

# List of repos to fetch docs from (name:docs_path:branch)
# Source-of-truth: product docs live in rezuscloud/rezuscloud.
REPOS=(
    "platform-website:docs:master"
    "rezuscloud:docs:main"
)

# Download docs via GitHub tarball API (authenticated via gh CLI)
# Usage: fetch_via_gh org repo branch docs_path dest
fetch_via_gh() {
    local org="$1" repo="$2" branch="$3" docs_path="$4" dest="$5"

    local tmpdir
    tmpdir=$(mktemp -d)
    local tarball="$tmpdir/repo.tar.gz"

    # Download tarball via gh api (authenticated)
    if ! gh api "repos/${org}/${repo}/tarball/${branch}" > "$tarball" 2>/dev/null; then
        rm -rf "$tmpdir"
        return 1
    fi

    # Extract only the docs/ directory
    mkdir -p "$tmpdir/extract"
    tar -xzf "$tarball" -C "$tmpdir/extract" 2>/dev/null || {
        rm -rf "$tmpdir"
        return 1
    }

    # Find the extracted directory (gh tarballs have a prefix like rezuscloud-repo-hash/)
    local extracted
    extracted=$(find "$tmpdir/extract" -maxdepth 1 -type d | tail -1)

    if [ -d "${extracted}/${docs_path}" ]; then
        cp -r "${extracted}/${docs_path}/." "$dest/"
        rm -rf "$tmpdir"
        return 0
    fi

    rm -rf "$tmpdir"
    return 1
}

for entry in "${REPOS[@]}"; do
    IFS=':' read -r repo docs_path branch <<< "$entry"
    echo "Fetching docs from rezuscloud/$repo ($docs_path/)..."

    dest="$TARGET_DIR/$repo"
    mkdir -p "$dest"

    # Try authenticated download via gh CLI first
    if command -v gh &>/dev/null && gh auth status &>/dev/null 2>&1; then
        if fetch_via_gh "rezuscloud" "$repo" "$branch" "$docs_path" "$dest"; then
            count=$(find "$dest" -name '*.md' | wc -l)
            echo "  ✓ Fetched $count markdown files via gh"
            continue
        fi
        echo "  gh download incomplete, trying git clone..."
    fi

    # Fallback: git clone (works for public repos without auth)
    tmpdir=$(mktemp -d)
    if git clone --depth 1 --branch "$branch" "https://github.com/rezuscloud/$repo.git" "$tmpdir/repo" 2>/dev/null; then
        if [ -d "$tmpdir/repo/$docs_path" ]; then
            cp -r "$tmpdir/repo/$docs_path/." "$dest/"
            count=$(find "$dest" -name '*.md' | wc -l)
            echo "  ✓ Fetched $count markdown files"
        else
            echo "  ⚠ No $docs_path/ directory found"
        fi
    else
        echo "  ⚠ Could not fetch $repo docs, skipping"
    fi
    rm -rf "$tmpdir"
done

# Ensure at least one file exists
if [ -z "$(find "$TARGET_DIR" -type f)" ]; then
    echo "# Documentation" > "$TARGET_DIR/README.md"
fi

echo "Done. Docs available in $TARGET_DIR/"
