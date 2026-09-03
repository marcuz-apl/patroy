#!/bin/sh
set -e

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo ".")"
BASE_VERSION="$(cat "$REPO_ROOT/VERSION" 2>/dev/null || echo "0.1.0")"
BASE_VERSION="$(printf "%s" "$BASE_VERSION" | tr -d " \r\n")"
if [ -z "$BASE_VERSION" ]; then
    BASE_VERSION="0.1.0"
fi

TODAY="$(date -u +%y%m%d)"
SEQ="123456789abcdefghijklmnopqrstuvwxyz"

LAST_BUILD="$(git log -1 --format="%(trailers:key=Alfazen-Build,valueonly)" 2>/dev/null | tr -d " \r\n")"
COUNTER="1"

case "$LAST_BUILD" in
    *"+${TODAY}"?)
        LAST_C="$(printf "%s" "$LAST_BUILD" | sed -E "s/.*\\+${TODAY}(.)/\\1/")"
        REST="${SEQ#*$LAST_C}"
        if [ -n "$REST" ] && [ "$REST" != "$SEQ" ]; then
            COUNTER="$(printf "%.1s" "$REST")"
        else
            COUNTER="z"
        fi
        ;;
esac

echo "${BASE_VERSION}+${TODAY}${COUNTER}"
