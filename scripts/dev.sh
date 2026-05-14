#!/usr/bin/env bash
# macOS / Linux용. Windows는 npm run dev:window (scripts/dev.windows.ps1).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Windows Git Bash: 부모 셸 PATH가 갱신되기 전이면 go 가 안 보일 수 있음(기본 설치 경로 보강)
if ! command -v go >/dev/null 2>&1; then
  for d in "/c/Program Files/Go/bin" "/c/Program Files (x86)/Go/bin"; do
    if [[ -x "${d}/go.exe" ]]; then
      export PATH="${d}:${PATH}"
      break
    fi
  done
fi

PORT="${PORT:-}"
if [[ -z "$PORT" && -f .env ]]; then
  PORT="$(
    grep -E '^[[:space:]]*PORT[[:space:]]*=' .env | head -1 | cut -d= -f2- | sed "s/^['\"]//; s/['\"]\$//" | tr -d '\r'
  )"
fi
PORT="${PORT:-8080}"

# 이전에 떠 있는 동일 포트 프로세스 정리 (에어 자식 외 잔류 방지)
PIDS="$(lsof -nP -iTCP:"${PORT}" -sTCP:LISTEN -t 2>/dev/null || true)"
if [[ -n "$PIDS" ]]; then
  # shellcheck disable=SC2086
  kill $PIDS 2>/dev/null || true
  sleep 0.3
  PIDS2="$(lsof -nP -iTCP:"${PORT}" -sTCP:LISTEN -t 2>/dev/null || true)"
  if [[ -n "$PIDS2" ]]; then
    # shellcheck disable=SC2086
    kill -9 $PIDS2 2>/dev/null || true
  fi
fi

AIR="${AIR:-}"
if [[ -n "$AIR" && -x "$AIR" ]]; then
  exec "$AIR"
fi
if command -v air >/dev/null 2>&1; then
  exec air
fi

BIN_DIR="${ROOT}/bin"
mkdir -p "$BIN_DIR"
AIRBIN="${BIN_DIR}/air"
if [[ ! -x "$AIRBIN" ]]; then
  echo "dev: Air를 ${AIRBIN} 에 설치합니다(최초 1회)." >&2
  GOBIN="$BIN_DIR" go install github.com/air-verse/air@latest
fi
exec "$AIRBIN"
