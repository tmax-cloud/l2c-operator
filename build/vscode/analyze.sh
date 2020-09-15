#!/bin/bash

if [ "$PROJECT_ID" == "" ]; then
  echo "PROJECT_ID is not set"
  exit 1
fi

MTA_DIR=/home/coder/mta
MTA_CLI="$MTA_DIR/bin/mta-cli"

PROJECT_PATH="/home/coder/project/$PROJECT_ID/"

OUTPUT_DIR=/home/coder/.local/share/code-server/User/globalStorage/redhat.mta-vscode-extension/.mta/tooling/data
JSON_PATH="$OUTPUT_DIR/model.json"
TMP_PATH=/tmp/model.json

echo "Analyzing..."

"$MTA_CLI" --toolingMode --source weblogic --target jeus:7 --sourceMode --ignorePattern '\.class$' --windupHome "$MTA_DIR"  --input "$PROJECT_PATH" --output "$OUTPUT_DIR/-38dkf89vj-wtx81drip"
jq '(.configurations[].summary.executedTimestamp = "'"$(date +"%D @ %R")"'")' "$JSON_PATH" > "$TMP_PATH" && mv "$TMP_PATH" "$JSON_PATH"
