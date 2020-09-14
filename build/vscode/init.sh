#!/bin/bash

if [ "$PROJECT_ID" == "" ]; then
  echo "PROJECT_ID is not set"
  exit 1
fi

PROJECTS_DIR=/home/coder/project/
PROJECT_DIR="$PROJECTS_DIR/$PROJECT_ID"

JSON_PATH=/home/coder/.local/share/code-server/User/globalStorage/redhat.mta-vscode-extension/.mta/tooling/data/model.json
TMP_PATH=/tmp/model.json

jq '(.configurations[].name = "'"$PROJECT_ID"'" | .configurations[].options.input[] = "'"$PROJECT_DIR"'")' "$JSON_PATH" > "$TMP_PATH" && mv "$TMP_PATH" "$JSON_PATH"

/usr/bin/code-server "$PROJECTS_DIR"
