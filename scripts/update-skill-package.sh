#!/bin/sh
set -eu

SKILL_DIR="skill/naver-place-rank-cli"
OUT="naver-place-rank-cli.skill"

if [ ! -d "$SKILL_DIR" ]; then
  echo "skill directory not found: $SKILL_DIR" >&2
  exit 1
fi

if ! command -v zip >/dev/null 2>&1; then
  echo "zip command not found" >&2
  exit 1
fi

(
  cd "$(dirname "$SKILL_DIR")"
  zip -r -q -X "../$OUT" "$(basename "$SKILL_DIR")" -x "*.DS_Store"
)
