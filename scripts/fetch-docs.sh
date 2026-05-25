#!/usr/bin/env bash
# fetch-docs.sh - Clone docs/ from configured repositories into docs/external/
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="$SCRIPT_DIR/../docs/external"

# Ensure clean state
rm -rf "$TARGET_DIR"
mkdir -p "$TARGET_DIR"

# List of repos to fetch docs from (name:docs_path:branch)
REPOS=(
    "platform-website:docs:master"
    "rezusctl:docs:main"
)

for entry in "${REPOS[@]}"; do
    IFS=':' read -r repo docs_path branch <<< "$entry"
    echo "Fetching docs from rezuscloud/$repo ($docs_path/)..."

    dest="$TARGET_DIR/$repo"
    mkdir -p "$dest"

    # Use git archive to fetch just the docs directory without cloning the whole repo
    if git archive --remote="https://github.com/rezuscloud/$repo.git" "$branch" "$docs_path" 2>/dev/null | tar -x -C "$dest" --strip-components=1 2>/dev/null; then
        count=$(find "$dest" -name '*.md' | wc -l)
        echo "  ✓ Fetched $count markdown files"
    else
        # Fallback: shallow clone and copy
        echo "  git archive not available, using shallow clone..."
        tmpdir=$(mktemp -d)
        git clone --depth 1 --sparse --branch "$branch" "https://github.com/rezuscloud/$repo.git" "$tmpdir/repo" 2>/dev/null || {
            echo "  ⚠ Could not clone $repo, skipping"
            rm -rf "$tmpdir"
            continue
        }
        (cd "$tmpdir/repo" && git sparse-checkout set "$docs_path" 2>/dev/null)
        if [ -d "$tmpdir/repo/$docs_path" ]; then
            cp -r "$tmpdir/repo/$docs_path/." "$dest/"
            count=$(find "$dest" -name '*.md' | wc -l)
            echo "  ✓ Fetched $count markdown files"
        else
            echo "  ⚠ No $docs_path/ directory found"
        fi
        rm -rf "$tmpdir"
    fi
done

# Ensure at least one file exists for go:embed to work
if [ -z "$(find "$TARGET_DIR" -type f)" ]; then
    echo "# Documentation" > "$TARGET_DIR/README.md"
fi

echo "Done. Docs available in $TARGET_DIR/"
