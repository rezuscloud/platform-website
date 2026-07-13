#!/usr/bin/env bash
# lib-versions.sh — Pure version-selection logic for multi-version doc fetching.
#
# Sourced by fetch-docs.sh (network orchestration) and fetch-docs.test.sh (unit
# tests). Contains NO network calls — only semver parsing and deduplication.
#
# Retention policy: deduplicate releases by minor version (major.minor), keep
# the latest release of each line. This is the Kubernetes/Helm docs model.
# For the current rezuscloud 0.0.1-NN scheme, minor "0.0" deduplicates to the
# highest -NN. When 0.1.0 ships, both "0.0" (latest) and "0.1" (latest) are kept.

# select_versions reads release tags (one per line) on stdin and outputs the
# retained tags (one per line), sorted newest-first.
#
# Each retained tag is the latest release of its minor version line.
select_versions() {
    declare -A seen_minor

    # Sort all tags newest-first (version sort handles 0.0.1-106 > 0.0.1-9
    # correctly, unlike naive string comparison).
    local sorted
    sorted=$(sort -rV)

    while IFS= read -r tag; do
        [ -z "$tag" ] && continue
        local minor_key
        minor_key=$(extract_minor_key "$tag") || continue

        if [ -z "${seen_minor[$minor_key]:-}" ]; then
            seen_minor[$minor_key]=1
            echo "$tag"
        fi
    done <<< "$sorted"
}

# extract_minor_key parses a semver-ish tag and outputs "major.minor".
# Returns non-zero if the tag cannot be parsed.
#
# Examples:
#   v0.0.1-106  → 0.0
#   v0.1.0      → 0.1
#   v1.2.3      → 1.2
extract_minor_key() {
    local tag="$1"
    local version="${tag#v}"          # strip leading v
    local base="${version%%-*}"       # strip prerelease (-NN or -rc1)
    # base is now "major.minor.patch"
    if [[ ! "$base" =~ ^[0-9]+\.[0-9]+\. ]]; then
        return 1
    fi
    # Extract major.minor (first two dot-separated fields)
    local IFS='.'
    read -r major minor _ <<< "$base"
    echo "${major}.${minor}"
}

# manifest_entry emits a single JSON object for a version entry.
# Usage: manifest_entry <version> <minor> <date>
manifest_entry() {
    local version="$1" minor="$2" date="$3"
    printf '    {"version": "%s", "minor": "%s", "date": "%s"}' \
        "$version" "$minor" "$date"
}

# manifest_entry_next emits the JSON object for the unreleased "next" entry.
manifest_entry_next() {
    printf '    {"version": "next", "label": "Unreleased (main)", "isNext": true}'
}

# write_manifest produces docs/versions/manifest.json from the selected versions.
# Usage: write_manifest <output_path> <latest_tag> <minor_of_latest> <date_of_latest>
#                        [tag2 minor2 date2] ... [-- next]
#
# The "-- next" sentinel appends the unreleased-next entry at the end.
write_manifest() {
    local output="$1"
    shift

    local entries=()
    local has_next=false
    local next_marker_seen=false

    while [ $# -gt 0 ]; do
        if [ "$1" = "--" ]; then
            next_marker_seen=true
            shift
            if [ "$1" = "next" ]; then
                has_next=true
            fi
            shift
            continue
        fi
        # version minor date
        entries+=("$(manifest_entry "$1" "$2" "$3")")
        shift 3
    done

    {
        echo "{"

        # Determine latest from first entry (entries are sorted newest-first)
        local latest="" latest_minor="" latest_date=""
        if [ ${#entries[@]} -gt 0 ]; then
            local first="${entries[0]}"
            latest=$(echo "$first" | grep -o '"version": "[^"]*"' | head -1 | cut -d'"' -f4)
        fi

        printf '  "latest": "%s",\n' "$latest"
        printf '  "next": "next",\n'
        printf '  "versions": [\n'

        local first=true
        for entry in "${entries[@]}"; do
            if [ "$first" = true ]; then
                first=false
            else
                printf ',\n'
            fi
            printf '%s' "$entry"
        done

        if [ "$has_next" = true ]; then
            if [ "$first" = false ]; then
                printf ',\n'
            fi
            printf '%s' "$(manifest_entry_next)"
        fi

        printf '\n  ]\n'
        echo "}"
    } > "$output"
}
