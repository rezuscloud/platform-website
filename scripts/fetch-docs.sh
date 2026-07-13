#!/usr/bin/env bash
# fetch-docs.sh - Fetch rezuscloud documentation for multiple release versions.
#
# Produces docs/versions/ with:
#   - One directory per retained release tag (deduplicated by minor version)
#   - A docs/versions/next/ directory for unreleased main-branch docs
#   - A docs/versions/manifest.json describing the available versions
#
# Retention policy (lib-versions.sh select_versions): for each minor version
# line (major.minor), keep only the latest release. This is the Kubernetes/Helm
# docs model. See lib-versions.sh for details.
#
# Backward compatibility: also copies the latest version's docs to
# docs/external/rezuscloud/ so the current (pre-versioning) Store keeps working
# until the versioned Store (#110) lands.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib-versions.sh"

ORG="rezuscloud"
REPO="rezuscloud"
DOCS_PATH="docs"
NEXT_BRANCH="main"

DOCS_ROOT="$SCRIPT_DIR/../docs"
TARGET_DIR="$DOCS_ROOT/versions"
EXTERNAL_DIR="$DOCS_ROOT/external"

# Clean state — idempotent re-runs
rm -rf "$TARGET_DIR"
rm -rf "$EXTERNAL_DIR"
mkdir -p "$TARGET_DIR"

# ---------------------------------------------------------------------------
# fetch_version_docs downloads a tarball of <org>/<repo> at <ref>, extracts
# the docs/ directory, and copies it to <dest>. Returns 0 on success.
# Uses gh api (authenticated) first, falls back to git clone (public repos).
# ---------------------------------------------------------------------------
fetch_version_docs() {
    local ref="$1" dest="$2"

    if command -v gh &>/dev/null && gh auth status &>/dev/null 2>&1; then
        _fetch_via_gh_tarball "$ref" "$dest" && return 0
    fi

    _fetch_via_git_clone "$ref" "$dest"
}

_fetch_via_gh_tarball() {
    local ref="$1" dest="$2"
    local tmpdir tarball extracted

    tmpdir=$(mktemp -d)
    tarball="$tmpdir/repo.tar.gz"

    if ! gh api "repos/${ORG}/${REPO}/tarball/${ref}" > "$tarball" 2>/dev/null; then
        rm -rf "$tmpdir"
        return 1
    fi

    mkdir -p "$tmpdir/extract"
    tar -xzf "$tarball" -C "$tmpdir/extract" 2>/dev/null || {
        rm -rf "$tmpdir"
        return 1
    }

    # gh tarballs have a prefix like rezuscloud-repo-<hash>/
    extracted=$(find "$tmpdir/extract" -maxdepth 1 -type d | tail -1)

    if [ -d "${extracted}/${DOCS_PATH}" ]; then
        mkdir -p "$dest"
        cp -r "${extracted}/${DOCS_PATH}/." "$dest/"
        rm -rf "$tmpdir"
        return 0
    fi

    rm -rf "$tmpdir"
    return 1
}

_fetch_via_git_clone() {
    local ref="$1" dest="$2"
    local tmpdir

    tmpdir=$(mktemp -d)
    if git clone --depth 1 --branch "$ref" "https://github.com/${ORG}/${REPO}.git" "$tmpdir/repo" 2>/dev/null; then
        if [ -d "$tmpdir/repo/$DOCS_PATH" ]; then
            mkdir -p "$dest"
            cp -r "$tmpdir/repo/$DOCS_PATH/." "$dest/"
            rm -rf "$tmpdir"
            return 0
        fi
    fi

    rm -rf "$tmpdir"
    return 1
}

# ---------------------------------------------------------------------------
# 1. Query all release tags + creation dates.
# ---------------------------------------------------------------------------
if ! command -v jq &>/dev/null; then
    echo "ERROR: jq is required for multi-version doc fetching" >&2
    exit 1
fi

echo "Querying releases from ${ORG}/${REPO}..."
ALL_RELEASES=$(gh release list --repo "${ORG}/${REPO}" --json tagName,createdAt --limit 500 2>/dev/null | \
    jq -r '.[] | "\(.createdAt)\t\(.tagName)"' || echo "")

if [ -z "$ALL_RELEASES" ]; then
    echo "  No releases found (or gh not authenticated). Fetching next only."
fi

# ---------------------------------------------------------------------------
# 2. Select versions: dedup by minor version, keep latest of each.
# ---------------------------------------------------------------------------
ALL_TAGS=$(echo "$ALL_RELEASES" | cut -f2)
SELECTED=""
if [ -n "$ALL_TAGS" ]; then
    SELECTED=$(echo "$ALL_TAGS" | select_versions)
fi

# ---------------------------------------------------------------------------
# 3. Fetch each selected version's docs.
# ---------------------------------------------------------------------------
MANIFEST_ARGS=()
LATEST_TAG=""
LATEST_DATE=""
LATEST_MINOR=""

while IFS= read -r tag; do
    [ -z "$tag" ] && continue

    # Look up the release date for this tag
    date=$(echo "$ALL_RELEASES" | grep -P "\t${tag}$" | cut -f1)
    minor=$(extract_minor_key "$tag") || minor=""

    echo "Fetching docs for ${tag} (minor ${minor})..."
    if fetch_version_docs "$tag" "$TARGET_DIR/$tag"; then
        count=$(find "$TARGET_DIR/$tag" -name '*.md' | wc -l)
        echo "  ✓ ${tag}: ${count} markdown files"
        MANIFEST_ARGS+=("$tag" "$minor" "$date")
        # First selected tag is the latest (select_versions sorts newest-first)
        if [ -z "$LATEST_TAG" ]; then
            LATEST_TAG="$tag"
            LATEST_DATE="$date"
            LATEST_MINOR="$minor"
        fi
    else
        echo "  ⚠ Could not fetch ${tag}, skipping"
    fi
done <<< "$SELECTED"

# ---------------------------------------------------------------------------
# 4. Fetch next (unreleased main branch).
# ---------------------------------------------------------------------------
echo "Fetching docs for next (${NEXT_BRANCH})..."
if fetch_version_docs "$NEXT_BRANCH" "$TARGET_DIR/next"; then
    count=$(find "$TARGET_DIR/next" -name '*.md' | wc -l)
    echo "  ✓ next: ${count} markdown files"
    MANIFEST_ARGS+=("--" "next")
else
    echo "  ⚠ Could not fetch next, versioned docs will not include unreleased"
fi

# ---------------------------------------------------------------------------
# 5. Write manifest.
# ---------------------------------------------------------------------------
if [ ${#MANIFEST_ARGS[@]} -gt 0 ]; then
    write_manifest "$TARGET_DIR/manifest.json" "${MANIFEST_ARGS[@]}"
    echo "  ✓ Manifest written: $(cat "$TARGET_DIR/manifest.json" | jq -r '.latest') latest"
else
    # Fallback: no versions at all — write an empty manifest
    echo '{"latest":"","next":"","versions":[]}' > "$TARGET_DIR/manifest.json"
    echo "  ⚠ No versions fetched, wrote empty manifest"
fi

# ---------------------------------------------------------------------------
# 6. Backward compatibility: copy latest version to docs/external/rezuscloud/.
#    The current (pre-versioning) Store strips external/<repo>/ and serves
#    these at their natural path. This keeps the live site working until the
#    versioned Store (#110) lands.
# ---------------------------------------------------------------------------
if [ -n "$LATEST_TAG" ]; then
    mkdir -p "$EXTERNAL_DIR/rezuscloud"
    cp -r "$TARGET_DIR/$LATEST_TAG/." "$EXTERNAL_DIR/rezuscloud/"
    echo "  ✓ Backward-compat: latest (${LATEST_TAG}) copied to docs/external/rezuscloud/"
fi

echo "Done. Versioned docs in ${TARGET_DIR}/"
