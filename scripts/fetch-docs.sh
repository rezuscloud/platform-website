#!/usr/bin/env bash
# fetch-docs.sh - Fetch documentation from the project wikis into docs/external/
#
# User-facing documentation is authored in the project GitHub wikis
# (rezuscloud.wiki, platform-website.wiki), organized by the Diátaxis framework.
# ADRs and repo-level context stay in the repositories and are never fetched
# here, so internal architecture content cannot reach the public docs site.
#
# Wikis are public git repos, so this needs no authentication.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="$SCRIPT_DIR/../docs/external"

# Ensure clean state
rm -rf "$TARGET_DIR"
mkdir -p "$TARGET_DIR"

# Wikis to fetch (repo base name → its <name>.wiki.git)
WIKIS=(
    "rezuscloud"
    "platform-website"
)

for name in "${WIKIS[@]}"; do
    echo "Fetching wiki: $name"
    dest="$TARGET_DIR/$name"
    mkdir -p "$dest"

    tmpdir=$(mktemp -d)
    url="https://github.com/rezuscloud/${name}.wiki.git"

    if git clone --depth 1 "$url" "$tmpdir/wiki" 2>/dev/null; then
        # Copy wiki contents (excluding the .git dir)
        cp -r "$tmpdir/wiki/." "$dest/"
        rm -rf "$dest/.git"
        count=$(find "$dest" -name '*.md' | wc -l)
        echo "  ✓ Fetched $count markdown files"
    else
        echo "  ⚠ Could not fetch $name wiki ($url), skipping"
    fi
    rm -rf "$tmpdir"
done

# Ensure at least one file exists (so the store never indexes an empty tree)
if [ -z "$(find "$TARGET_DIR" -type f)" ]; then
    echo "# Documentation" > "$TARGET_DIR/README.md"
fi

echo "Done. Docs available in $TARGET_DIR/"
