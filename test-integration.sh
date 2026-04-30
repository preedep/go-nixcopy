#!/usr/bin/env bash
# test-integration.sh — spin up local service containers and run integration tests.
#
# Usage:
#   ./test-integration.sh            # run all integration tests
#   ./test-integration.sh local      # local storage only (no Docker)
#   ./test-integration.sh s3         # S3/MinIO tests only
#   ./test-integration.sh sftp       # SFTP tests only
#   ./test-integration.sh -v         # pass -v to go test (verbose)
#   ./test-integration.sh --no-clean # keep containers running after tests

set -euo pipefail

# ── colours ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()    { echo -e "${CYAN}[INFO]${NC}  $*"; }
success() { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }
header()  { echo -e "\n${BOLD}═══ $* ═══${NC}"; }

# ── defaults ─────────────────────────────────────────────────────────────────
RUN_LOCAL=false
RUN_S3=false
RUN_SFTP=false
RUN_FTPS=false
RUN_ALL=true
VERBOSE=""
CLEAN=true

# ── arg parsing ──────────────────────────────────────────────────────────────
for arg in "$@"; do
  case "$arg" in
    local)      RUN_LOCAL=true;  RUN_ALL=false ;;
    s3)         RUN_S3=true;     RUN_ALL=false ;;
    sftp)       RUN_SFTP=true;   RUN_ALL=false ;;
    ftps)       RUN_FTPS=true;   RUN_ALL=false ;;
    -v|--verbose) VERBOSE="-v" ;;
    --no-clean) CLEAN=false ;;
    -h|--help)
      echo "Usage: $0 [local|s3|sftp|ftps] [-v] [--no-clean]"
      exit 0 ;;
    *) error "Unknown argument: $arg"; exit 1 ;;
  esac
done

if $RUN_ALL; then
  RUN_LOCAL=true
  RUN_S3=true
  RUN_SFTP=true
  RUN_FTPS=true
fi

# ── constants ─────────────────────────────────────────────────────────────────
MINIO_CONTAINER="nixcopy-minio-test"
SFTP_CONTAINER="nixcopy-sftp-test"
FTPS_CONTAINER="nixcopy-ftps-test"
MINIO_PORT=9000
SFTP_PORT=2222
FTPS_PORT=21
FTPS_PASSIVE_MIN=30000
FTPS_PASSIVE_MAX=30009
MINIO_USER="minioadmin"
MINIO_PASS="minioadmin"
MINIO_BUCKET="test-bucket"
SFTP_USER="testuser"
SFTP_PASS="testpass"
FTPS_USER="testuser"
FTPS_PASS="testpass"

# ── helpers ───────────────────────────────────────────────────────────────────
require_docker() {
  if ! command -v docker &>/dev/null; then
    error "Docker is not installed or not in PATH."
    exit 1
  fi
  if ! docker info &>/dev/null; then
    error "Docker daemon is not running."
    exit 1
  fi
}

port_free() {
  ! lsof -iTCP:"$1" -sTCP:LISTEN &>/dev/null 2>&1
}

container_running() {
  docker ps --format '{{.Names}}' | grep -q "^$1$"
}

wait_healthy() {
  local name=$1 timeout=${2:-30} interval=2
  info "Waiting for $name to be healthy..."
  for ((i=0; i<timeout; i+=interval)); do
    if docker inspect --format='{{.State.Health.Status}}' "$name" 2>/dev/null | grep -q "healthy"; then
      success "$name is healthy"
      return 0
    fi
    sleep $interval
  done
  error "$name did not become healthy within ${timeout}s"
  docker logs "$name" 2>&1 | tail -20
  return 1
}

# ── cleanup ───────────────────────────────────────────────────────────────────
cleanup() {
  if ! $CLEAN; then
    warn "Skipping cleanup (--no-clean). Containers still running:"
    docker ps --filter "name=nixcopy-" --format "  {{.Names}} → {{.Ports}}"
    return
  fi

  header "Cleanup"
  for name in "$MINIO_CONTAINER" "$SFTP_CONTAINER" "$FTPS_CONTAINER"; do
    if docker ps -a --format '{{.Names}}' | grep -q "^${name}$"; then
      info "Stopping and removing $name"
      docker rm -f "$name" &>/dev/null || true
    fi
  done
  success "Containers removed"
}

trap cleanup EXIT

# ── MinIO ─────────────────────────────────────────────────────────────────────
start_minio() {
  header "Starting MinIO (S3-compatible)"

  if container_running "$MINIO_CONTAINER"; then
    info "MinIO container already running — reusing it"
    return 0
  fi

  if ! port_free $MINIO_PORT; then
    error "Port $MINIO_PORT is already in use. Stop the process or use --no-clean to reuse a running container."
    exit 1
  fi

  docker run -d \
    --name "$MINIO_CONTAINER" \
    -p "${MINIO_PORT}:9000" \
    -e MINIO_ROOT_USER="$MINIO_USER" \
    -e MINIO_ROOT_PASSWORD="$MINIO_PASS" \
    --health-cmd "curl -sf http://localhost:9000/minio/health/live" \
    --health-interval 2s \
    --health-retries 15 \
    minio/minio server /data \
    > /dev/null

  wait_healthy "$MINIO_CONTAINER"

  info "Creating bucket: $MINIO_BUCKET"
  docker run --rm \
    --network host \
    -e MC_HOST_local="http://${MINIO_USER}:${MINIO_PASS}@localhost:${MINIO_PORT}" \
    minio/mc mb --ignore-existing "local/${MINIO_BUCKET}" \
    > /dev/null

  success "MinIO ready at http://localhost:${MINIO_PORT}"
}

# ── SFTP ──────────────────────────────────────────────────────────────────────
start_sftp() {
  header "Starting SFTP server (atmoz/sftp)"

  if container_running "$SFTP_CONTAINER"; then
    info "SFTP container already running — reusing it"
    return 0
  fi

  if ! port_free $SFTP_PORT; then
    error "Port $SFTP_PORT is already in use."
    exit 1
  fi

  # atmoz/sftp format:  user:pass:::upload_dir
  docker run -d \
    --name "$SFTP_CONTAINER" \
    -p "${SFTP_PORT}:22" \
    atmoz/sftp \
    "${SFTP_USER}:${SFTP_PASS}:::upload" \
    > /dev/null

  # atmoz/sftp has no built-in healthcheck; poll SSH port
  info "Waiting for SFTP to accept connections..."
  for i in $(seq 1 20); do
    if docker exec "$SFTP_CONTAINER" test -d /home/"$SFTP_USER"/upload 2>/dev/null; then
      success "SFTP ready at localhost:${SFTP_PORT} (user: $SFTP_USER)"
      return 0
    fi
    sleep 1
  done

  error "SFTP container did not become ready in time"
  docker logs "$SFTP_CONTAINER" 2>&1 | tail -20
  return 1
}

# ── FTPS ──────────────────────────────────────────────────────────────────────
start_ftps() {
  header "Starting FTPS server (stilliard/pure-ftpd, explicit TLS)"

  if container_running "$FTPS_CONTAINER"; then
    info "FTPS container already running — reusing it"
    return 0
  fi

  if ! port_free $FTPS_PORT; then
    error "Port $FTPS_PORT is already in use."
    exit 1
  fi

  docker run -d \
    --name "$FTPS_CONTAINER" \
    -p "${FTPS_PORT}:21" \
    -p "${FTPS_PASSIVE_MIN}-${FTPS_PASSIVE_MAX}:${FTPS_PASSIVE_MIN}-${FTPS_PASSIVE_MAX}" \
    -e FTP_USER_NAME="$FTPS_USER" \
    -e FTP_USER_PASS="$FTPS_PASS" \
    -e FTP_USER_HOME=/home/testuser \
    -e PUBLICHOST=127.0.0.1 \
    -e PASSIVE_MIN_PORT="$FTPS_PASSIVE_MIN" \
    -e PASSIVE_MAX_PORT="$FTPS_PASSIVE_MAX" \
    -e ADDED_FLAGS="--tls=1" \
    stilliard/pure-ftpd \
    > /dev/null

  info "Waiting for FTPS to be ready..."
  for i in $(seq 1 30); do
    if docker exec "$FTPS_CONTAINER" ps -e -o comm 2>/dev/null | grep -q pure-ftpd; then
      success "FTPS ready at localhost:${FTPS_PORT} (user: $FTPS_USER, TLS: explicit, skip_verify: true)"
      return 0
    fi
    sleep 1
  done

  error "FTPS container did not become ready in time"
  docker logs "$FTPS_CONTAINER" 2>&1 | tail -20
  return 1
}

# ── run tests ─────────────────────────────────────────────────────────────────
run_tests() {
  local tags="$1"
  local run_filter="$2"
  shift 2
  local env_vars=("$@")

  local cmd=(go test -tags=integration -race $VERBOSE)
  [[ -n "$run_filter" ]] && cmd+=(-run "$run_filter")
  cmd+=(./internal/infrastructure/storage/...)

  info "Running: ${cmd[*]}"
  env "${env_vars[@]}" "${cmd[@]}"
}

# ── main ──────────────────────────────────────────────────────────────────────
header "go-nixcopy Integration Tests"
echo -e "  Target:  ${BOLD}$(go env GOOS)/$(go env GOARCH)${NC}"
echo -e "  Suites:  local=$RUN_LOCAL  s3=$RUN_S3  sftp=$RUN_SFTP"

EXIT_CODE=0

# ─ local storage ─
if $RUN_LOCAL; then
  header "Suite: Local Storage"
  info "No containers needed — using t.TempDir()"
  if go test -tags=integration -race $VERBOSE -run "TestLocalStorage" \
       ./internal/infrastructure/storage/...; then
    success "Local storage tests passed"
  else
    error "Local storage tests FAILED"
    EXIT_CODE=1
  fi
fi

# ─ S3 / MinIO ─
if $RUN_S3; then
  require_docker
  start_minio

  header "Suite: S3 / MinIO"
  if run_tests "" "TestS3Storage" \
       "S3_ENDPOINT=http://localhost:${MINIO_PORT}" \
       "S3_ACCESS_KEY=${MINIO_USER}" \
       "S3_SECRET_KEY=${MINIO_PASS}" \
       "S3_BUCKET=${MINIO_BUCKET}" \
       "S3_REGION=us-east-1"; then
    success "S3 tests passed"
  else
    error "S3 tests FAILED"
    EXIT_CODE=1
  fi
fi

# ─ SFTP ─
if $RUN_SFTP; then
  require_docker
  start_sftp

  header "Suite: SFTP"
  if run_tests "" "TestSFTPStorage" \
       "SFTP_HOST=localhost" \
       "SFTP_PORT=${SFTP_PORT}" \
       "SFTP_USERNAME=${SFTP_USER}" \
       "SFTP_PASSWORD=${SFTP_PASS}"; then
    success "SFTP tests passed"
  else
    error "SFTP tests FAILED"
    EXIT_CODE=1
  fi
fi

# ─ FTPS ─
if $RUN_FTPS; then
  require_docker
  start_ftps

  header "Suite: FTPS"
  if run_tests "" "TestFTPSStorage" \
       "FTPS_HOST=localhost" \
       "FTPS_PORT=${FTPS_PORT}" \
       "FTPS_USERNAME=${FTPS_USER}" \
       "FTPS_PASSWORD=${FTPS_PASS}" \
       "FTPS_TLS_MODE=explicit" \
       "FTPS_SKIP_VERIFY=true"; then
    success "FTPS tests passed"
  else
    error "FTPS tests FAILED"
    EXIT_CODE=1
  fi
fi

# ─ summary ─
header "Summary"
if [[ $EXIT_CODE -eq 0 ]]; then
  success "All integration tests passed ✓"
else
  error "One or more test suites FAILED ✗"
fi

exit $EXIT_CODE
