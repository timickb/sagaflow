#!/bin/sh
set -eu

apk add --no-cache curl jq >/dev/null

echo "Waiting for Kafka Connect..."
until curl -fsS http://debezium_connect:8083/connectors >/dev/null; do
  sleep 2
done

echo "Mounted connectors dir:"
ls -la /connectors || true

found=0

for file in /connectors/*.json; do
  [ -f "$file" ] || continue

  found=1

  name="${file##*/}"
  name="${name%.json}"

  echo "Applying connector: $name from $file"

  curl -fsS -X PUT "http://debezium_connect:8083/connectors/$name/config" \
    -H "Content-Type: application/json" \
    --data-binary @"$file"

  echo
done

if [ "$found" = "0" ]; then
  echo "No connector JSON files found in /connectors"
  exit 1
fi

echo "All connector configs applied"