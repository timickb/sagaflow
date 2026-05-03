SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PAYLOAD_FILE="$SCRIPT_DIR/start_process_order_payload.json"

grpcurl -plaintext -d "{
  \"saga_name\": \"process-order\",
  \"saga_version\": 1,
  \"initial_context\": \"$(base64 -w 0 < "$PAYLOAD_FILE")\"
}" localhost:3731 sagaflow.v1.SagaflowService/Start