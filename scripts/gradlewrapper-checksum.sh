#!/usr/bin/env bash
set -euo pipefail

# Description:
#   Downloads the list of all Gradle versions from Gradle's official service,
#   and filters it to only the fields needed for wrapper checksum verification.
#
#
# Output:
#   - JSON file: pkg/builtin/builtinaction/gradle/gradlecommon/wrapper-checksums.json
#
# Notes:
#   - If `jq` is not installed, the full unfiltered JSON is saved.

# config
GRADLE_VERSIONS_URL="https://services.gradle.org/versions/all"

# init
tmpFile=$(mktemp)
trap 'rm -f "$tmpFile"' EXIT
scriptDir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
targetFile="$scriptDir/../pkg/builtin/builtinaction/gradle/gradlecommon/wrapper-checksums.json"

# download
echo "Updating Gradle wrapper checksums from $GRADLE_VERSIONS_URL to wrapper-checksums.json..."
curl -fsSL "$GRADLE_VERSIONS_URL" -o "$tmpFile"

# filter to relevant fields only, if jq is available
if command -v jq >/dev/null 2>&1; then
    jq '[ .[] | select(.nightly != true and .snapshot != true) | { version, checksum, wrapperChecksum } ]' "$tmpFile" > "$targetFile"
else
    cp "$tmpFile" "$targetFile"
fi

echo "Updated Gradle wrapper checksums."
