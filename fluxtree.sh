#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
NC='\033[0m'

# Get kustomization data: name|readyStatus|message|dep1,dep2,...
DATA=$(kubectl -n flux-system get kustomization -o json | jq -r '
  .items[] |
  (.metadata.name) + "|" +
  ((.status.conditions // [] | map(select(.type == "Ready")) | first // {status: "Unknown"}).status) + "|" +
  ((.status.conditions // [] | map(select(.type == "Ready")) | first // {message: ""}).message) + "|" +
  ([(.spec.dependsOn // [] | .[].name)] | join(","))
')

declare -A STATUS
declare -A MESSAGE
declare -A DEPS
declare -a NAMES

while IFS='|' read -r name ready message deps; do
  [[ -z "$name" ]] && continue
  NAMES+=("$name")
  STATUS[$name]="$ready"
  MESSAGE[$name]="$message"
  DEPS[$name]="$deps"
done <<< "$DATA"

# Find all nodes that list the given parent as a dependency
children_of() {
  local parent=$1
  for name in "${NAMES[@]}"; do
    local deps="${DEPS[$name]}"
    [[ -z "$deps" ]] && continue
    IFS=',' read -ra dep_array <<< "$deps"
    for dep in "${dep_array[@]}"; do
      if [[ "$dep" == "$parent" ]]; then
        echo "$name"
        break
      fi
    done
  done
}

print_node() {
  local name=$1
  local status="${STATUS[$name]}"
  local message="${MESSAGE[$name]}"

  local colour icon
  if [[ "$status" == "True" ]]; then
    colour="$GREEN"
    icon="●"
  elif [[ "$message" == *"dependency"* ]]; then
    colour="$YELLOW"
    icon="◌"
  else
    colour="$RED"
    icon="✗"
  fi

  echo -en "${colour}${icon} ${name}${NC}"
}

print_tree() {
  local name=$1
  local prefix=$2
  local connector=$3

  echo -en "${prefix}${connector}"
  print_node "$name"
  echo

  local extension
  if [[ -z "$connector" ]]; then
    extension=""
  elif [[ "$connector" == "└── " ]]; then
    extension="    "
  else
    extension="│   "
  fi
  local child_prefix="${prefix}${extension}"

  local children
  mapfile -t children < <(children_of "$name")
  local count=${#children[@]}
  local i=0

  for child in "${children[@]}"; do
    i=$((i + 1))
    if [[ $i -eq $count ]]; then
      print_tree "$child" "$child_prefix" "└── "
    else
      print_tree "$child" "$child_prefix" "├── "
    fi
  done
}

# Roots are nodes with no dependencies
for name in "${NAMES[@]}"; do
  if [[ -z "${DEPS[$name]}" ]]; then
    print_tree "$name" "" ""
  fi
done
