#!/usr/bin/env bash

# =============================================================================
# update-image-tag.sh
# Updates kustomize image tags and commits the change for Flux to pick up.
# =============================================================================

set -euo pipefail

# ─── Configuration ────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd -P)"
KUSTOMIZE_BASE_PATH="${REPO_ROOT}/k8s/kustomize/apps/envs"
DEFAULT_ENV="dev"

REGISTRY="ghcr.io"
REPO_OWNER="sudheerkumarp0357"
REPO_NAME="streamforge"

# Service → image artifact mapping (mirrors CI pipeline artifacts)
declare -A SERVICE_IMAGE_MAP=(
  ["frontend"]="frontend"
  ["api"]="api"
  ["transcoder"]="transcoder"
)

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()    { echo -e "${CYAN}[INFO]${NC}  $*"; }
success() { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# ─── Usage ────────────────────────────────────────────────────────────────────
usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Updates image tags in kustomization.yml using 'kustomize edit set image'.
Commits and pushes the change so Flux can detect and apply it.

No kubectl or cluster access required — this is a local file operation only.

OPTIONS:
  -s, --service <name>     Service to update: frontend | api | transcoder
                           Repeat for multiple services.
  -t, --tag <tag>          Image tag (e.g. sha-abc1234).
                           One tag applies to all services, or pair with each -s.
  -e, --env <env>          Target environment (default: ${DEFAULT_ENV})
  -d, --dry-run            Show what would change without modifying files or pushing.
  -h, --help               Show this help.

EXAMPLES:
  # Update single service
  $(basename "$0") -s frontend -t sha-abc1234

  # Update multiple services with same tag
  $(basename "$0") -s frontend -s api -s transcoder -t sha-abc1234

  # Update multiple services with different tags
  $(basename "$0") -s frontend -t sha-aaa1111 -s api -t sha-bbb2222

  # Target prod environment
  $(basename "$0") -s api -t sha-abc1234 -e prod

  # Dry run — no files modified, no git push
  $(basename "$0") -s frontend -t sha-abc1234 --dry-run
EOF
  exit "${1:-0}"
}

# ─── Dependency check ─────────────────────────────────────────────────────────
check_dependencies() {
  local missing=()
  for cmd in kustomize git; do
    if ! command -v "${cmd}" &>/dev/null; then
      missing+=("${cmd}")
    fi
  done
  if [[ ${#missing[@]} -gt 0 ]]; then
    error "Missing required tools: ${missing[*]}"
    error "Install kustomize: https://kubectl.docs.kubernetes.io/installation/kustomize/"
    exit 1
  fi
}

# ─── Parse Arguments ──────────────────────────────────────────────────────────
SERVICES=()
TAGS=()
ENV="${DEFAULT_ENV}"
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    -s|--service)
      [[ -z "${2:-}" ]] && { error "Missing value for --service"; usage 1; }
      SERVICES+=("$2"); shift 2 ;;
    -t|--tag)
      [[ -z "${2:-}" ]] && { error "Missing value for --tag"; usage 1; }
      TAGS+=("$2"); shift 2 ;;
    -e|--env)
      [[ -z "${2:-}" ]] && { error "Missing value for --env"; usage 1; }
      ENV="$2"; shift 2 ;;
    -d|--dry-run)
      DRY_RUN=true; shift ;;
    -h|--help)
      usage 0 ;;
    *)
      error "Unknown argument: $1"; usage 1 ;;
  esac
done

# ─── Validate Inputs ──────────────────────────────────────────────────────────
[[ ${#SERVICES[@]} -eq 0 ]] && { error "At least one --service is required."; usage 1; }
[[ ${#TAGS[@]} -eq 0 ]]     && { error "At least one --tag is required."; usage 1; }

# Single tag → apply to all services
if [[ ${#TAGS[@]} -eq 1 && ${#SERVICES[@]} -gt 1 ]]; then
  single_tag="${TAGS[0]}"
  TAGS=()
  for _ in "${SERVICES[@]}"; do TAGS+=("${single_tag}"); done
fi

if [[ ${#SERVICES[@]} -ne ${#TAGS[@]} ]]; then
  error "Mismatch: ${#SERVICES[@]} service(s) but ${#TAGS[@]} tag(s)."
  error "Either provide one tag for all services, or one tag per service."
  usage 1
fi

for svc in "${SERVICES[@]}"; do
  if [[ -z "${SERVICE_IMAGE_MAP[$svc]+_}" ]]; then
    error "Unknown service: '${svc}'. Valid: ${!SERVICE_IMAGE_MAP[*]}"
    exit 1
  fi
done

# ─── Resolve kustomization.yml ────────────────────────────────────────────────
KUSTOMIZATION_DIR="${KUSTOMIZE_BASE_PATH}/${ENV}"
KUSTOMIZATION_FILE="${KUSTOMIZATION_DIR}/kustomization.yml"

if [[ ! -f "${KUSTOMIZATION_FILE}" ]]; then
  error "File not found: ${KUSTOMIZATION_FILE}"
  exit 1
fi

# ─── Main ─────────────────────────────────────────────────────────────────────
check_dependencies

info "Environment : ${ENV}"
info "Target file : ${KUSTOMIZATION_FILE}"
${DRY_RUN} && warn "DRY RUN — no files will be modified or pushed."
echo ""

UPDATED=0
COMMIT_MSG_PARTS=()

for i in "${!SERVICES[@]}"; do
  svc="${SERVICES[$i]}"
  tag="${TAGS[$i]}"
  artifact="${SERVICE_IMAGE_MAP[$svc]}"
  image="${REGISTRY}/${REPO_OWNER}/${REPO_NAME}/${artifact}"

  info "Service: ${svc}"
  info "  Image : ${image}"
  info "  Tag   : ${tag}"

  if ${DRY_RUN}; then
    info "  [DRY RUN] Would run: kustomize edit set image ${image}=${image}:${tag}"
  else
    # kustomize handles YAML parsing correctly — no sed/awk fragility
    # No cluster access needed — purely a local file edit
    (
      cd "${KUSTOMIZATION_DIR}"
      kustomize edit set image "${image}=${image}:${tag}"
    )
    success "  Updated."
  fi

  COMMIT_MSG_PARTS+=("${svc}=${tag}")
  UPDATED=$((UPDATED + 1))
  echo ""
done

# ─── Git commit and push ───────────────────────────────────────────────────────
if [[ ${UPDATED} -eq 0 ]]; then
  warn "No services were updated."
  exit 0
fi