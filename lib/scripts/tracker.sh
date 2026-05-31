#!/usr/bin/env bash
# tracker.sh - Sends bizevents to the offon-challenge-tracker Cloud Run service

TRACKER_URL="https://offon-challenge-tracker-291758365471.europe-west1.run.app"
EVENT_TYPE="offon-challenges"
SESSION_ID_FILE="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/.offon-session-id"

# -----------------------------------------------------------------------------
# Set tracking context and ensure a session ID exists
# Usage: set_tracking_context "00-lex-imperfecta" "beginner"
# -----------------------------------------------------------------------------
set_tracking_context() {
  local adventure=$1
  local level=$2

  export ADVENTURE="$adventure"
  export LEVEL="$level"

  if [[ ! -f "$SESSION_ID_FILE" ]]; then
    uuidgen > "$SESSION_ID_FILE"
  fi
  export OFFON_SESSION_ID
  OFFON_SESSION_ID=$(cat "$SESSION_ID_FILE")
}

# -----------------------------------------------------------------------------
# Send an event to the tracker (silent, never fails the caller)
# Usage: send_event "event.action" '{"extra": "fields"}'
# -----------------------------------------------------------------------------
send_event() {
  local action=$1
  local extra_fields=${2:-"{}"}

  local payload
  payload=$(jq -n \
    --arg event_type "$EVENT_TYPE" \
    --arg action "$action" \
    --arg adventure "${ADVENTURE:-unknown}" \
    --arg level "${LEVEL:-unknown}" \
    --arg session_id "${OFFON_SESSION_ID:-unknown}" \
    --arg github_user "${GITHUB_USER:-}" \
    --arg github_repo "${GITHUB_REPOSITORY:-}" \
    --argjson extra "$extra_fields" \
    '{
      "type": $event_type,
      "action": $action,
      "adventure": $adventure,
      "level": $level,
      "session.id": $session_id,
      "github.user": $github_user,
      "github.repo": $github_repo
    } + $extra'
  )

  curl -sS -X POST "$TRACKER_URL" \
    -H "Content-Type: application/json" \
    -d "$payload" \
    >/dev/null 2>&1 || true
}

track_container_created() {
  send_event "container.created"
}

track_container_initialized() {
  send_event "container.initialized"
}

track_verification_completed() {
  local status=$1
  local failed_checks=${2:-"[]"}

  send_event "verification.completed" "$(jq -n \
    --arg status "$status" \
    --argjson failed_checks "$failed_checks" \
    '{status: $status, failed_checks: $failed_checks}'
  )"
}
