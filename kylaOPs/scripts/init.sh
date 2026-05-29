#!/usr/bin/env bash
# Initialise the kylaOPs environment: copy every *.env.example to *.env if
# the target does not already exist. Idempotent — safe to run repeatedly.
set -euo pipefail

OPS_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENVS_DIR="$OPS_DIR/envs"

green=$'\033[0;32m'; yellow=$'\033[0;33m'; gray=$'\033[0;90m'; reset=$'\033[0m'

if [ ! -d "$ENVS_DIR" ]; then
    echo "${yellow}envs/ directory missing at $ENVS_DIR${reset}" >&2
    exit 1
fi

created=0
skipped=0

shopt -s nullglob dotglob
for example in "$ENVS_DIR"/*.env.example; do
    target="${example%.example}"
    if [ -e "$target" ]; then
        echo "${gray}skip${reset}    $(basename "$target") ${gray}(already exists)${reset}"
        skipped=$((skipped + 1))
    else
        cp "$example" "$target"
        echo "${green}create${reset}  $(basename "$target")"
        created=$((created + 1))
    fi
done

echo
echo "${green}✓${reset} init complete — ${green}${created}${reset} created, ${gray}${skipped}${reset} skipped"
echo "${gray}edit the .env files under kylaOPs/envs/ and run 'make up'${reset}"
