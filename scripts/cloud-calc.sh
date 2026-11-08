#!/usr/bin/env bash
# ==============================================================================
# cloud-calc.sh: Multi-Cloud Pricing Calculator Universal CLI
# Adheres strictly to Limmireu Global Rule (set -euo pipefail & clear logging)
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="${SCRIPT_DIR}/../bin/cloud-pricing-mcp"

# Fallback to system local bin if not in project bin
if [[ ! -x "${BIN_PATH}" ]]; then
    if [[ -x "${HOME}/.local/bin/cloud-pricing-mcp" ]]; then
        BIN_PATH="${HOME}/.local/bin/cloud-pricing-mcp"
    else
        echo "[ERROR] cloud-pricing-mcp binary not found! Please build it first." >&2
        exit 1
    fi
fi

log_info() {
    echo -e "\033[1;34m[INFO]\033[0m $*"
}

log_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $*"
}

log_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $*" >&2
}

print_usage() {
    cat << 'EOF'
================================================================================
  Multi-Cloud Pricing Calculator CLI (Shell-First Wrapper)
  Providers: NCP (Naver Cloud), Cloudflare, AWS, Azure, GCP
================================================================================
Usage:
  ./cloud-calc.sh [command] [options]

Commands:
  compare       Cross-cloud cost comparison (NCP vs AWS vs Azure vs GCP)
  ncp           Calculate Naver Cloud Platform cost
  cloudflare    Calculate Cloudflare Workers & R2 cost
  aws           Calculate AWS EC2, EBS & Egress cost
  azure         Calculate Azure VM, Disk & Egress cost
  gcp           Calculate Google Cloud Compute Engine & Storage cost
  catalog       View full pricing catalog or filter by provider
  quick-quote   Interactive wizard for instant cost recommendation
  help          Show this help message

Quick Examples:
  # 1. Compare 4 clouds for 2 vCPU, 4GB RAM, 50GB SSD, 100GB Outbound
  ./cloud-calc.sh compare --vcpu 2 --ram 4 --storage 50 --egress 100

  # 2. NCP Server Standard s2-g2 with 100GB SSD
  ./cloud-calc.sh ncp --server s2-g2 --storage 100 --type ssd

  # 3. Cloudflare 25M Workers requests + 1TB R2 Storage
  ./cloud-calc.sh cloudflare --workers-reqs 25 --r2-storage 1000

  # 4. AWS EC2 Graviton2 Spot instance
  ./cloud-calc.sh aws --instance t4g.medium --ebs 100 --spot

  # 5. Show only NCP catalog products
  ./cloud-calc.sh catalog --provider ncp
================================================================================
EOF
}

cmd="${1:-help}"
shift || true

case "${cmd}" in
    compare)
        VCPU=2
        RAM=4
        STORAGE=50
        STORAGE_TYPE="ssd"
        HOURS=730
        EGRESS=100
        SPOT="false"

        while [[ $# -gt 0 ]]; do
            case "$1" in
                --vcpu) VCPU="$2"; shift 2 ;;
                --ram) RAM="$2"; shift 2 ;;
                --storage) STORAGE="$2"; shift 2 ;;
                --storage-type|--type) STORAGE_TYPE="$2"; shift 2 ;;
                --hours) HOURS="$2"; shift 2 ;;
                --egress) EGRESS="$2"; shift 2 ;;
                --spot) SPOT="true"; shift 1 ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
        done

        log_info "Calculating cross-cloud comparison for: ${VCPU} vCPU, ${RAM}GB RAM, ${STORAGE}GB ${STORAGE_TYPE^^}, ${EGRESS}GB Egress (Spot: ${SPOT})..."
        if [[ "${SPOT}" == "true" ]]; then
            "${BIN_PATH}" compare -vcpu "${VCPU}" -ram "${RAM}" -storage "${STORAGE}" -storage-type "${STORAGE_TYPE}" -hours "${HOURS}" -egress "${EGRESS}" -spot
        else
            "${BIN_PATH}" compare -vcpu "${VCPU}" -ram "${RAM}" -storage "${STORAGE}" -storage-type "${STORAGE_TYPE}" -hours "${HOURS}" -egress "${EGRESS}"
        fi
        ;;

    ncp)
        SERVER="s2-g2"
        COUNT=1
        HOURS=730
        STORAGE=50
        STORAGE_TYPE="ssd"
        EGRESS=0

        while [[ $# -gt 0 ]]; do
            case "$1" in
                --server) SERVER="$2"; shift 2 ;;
                --count) COUNT="$2"; shift 2 ;;
                --hours) HOURS="$2"; shift 2 ;;
                --storage) STORAGE="$2"; shift 2 ;;
                --type|--storage-type) STORAGE_TYPE="$2"; shift 2 ;;
                --egress) EGRESS="$2"; shift 2 ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
        done

        log_info "Calculating NCP Cost for ${SERVER} (Count: ${COUNT}, Storage: ${STORAGE}GB ${STORAGE_TYPE})..."
        "${BIN_PATH}" ncp -server "${SERVER}" -count "${COUNT}" -hours "${HOURS}" -storage "${STORAGE}" -storage-type "${STORAGE_TYPE}" -egress "${EGRESS}"
        ;;

    cloudflare|cf)
        WORKERS_M=0
        R2_GB=0
        OPS_A_M=0
        OPS_B_M=0

        while [[ $# -gt 0 ]]; do
            case "$1" in
                --workers-reqs|--workers-m) WORKERS_M="$2"; shift 2 ;;
                --r2-storage|--r2-gb) R2_GB="$2"; shift 2 ;;
                --ops-a) OPS_A_M="$2"; shift 2 ;;
                --ops-b) OPS_B_M="$2"; shift 2 ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
        done

        log_info "Calculating Cloudflare Cost (Workers: ${WORKERS_M}M, R2: ${R2_GB}GB)..."
        "${BIN_PATH}" cloudflare -workers-reqs-m "${WORKERS_M}" -r2-storage-gb "${R2_GB}" -r2-ops-a-m "${OPS_A_M}" -r2-ops-b-m "${OPS_B_M}"
        ;;

    aws)
        INSTANCE="t4g.medium"
        COUNT=1
        HOURS=730
        EBS=50
        EGRESS=0
        SPOT="false"

        while [[ $# -gt 0 ]]; do
            case "$1" in
                --instance) INSTANCE="$2"; shift 2 ;;
                --count) COUNT="$2"; shift 2 ;;
                --hours) HOURS="$2"; shift 2 ;;
                --ebs|--storage) EBS="$2"; shift 2 ;;
                --egress) EGRESS="$2"; shift 2 ;;
                --spot) SPOT="true"; shift 1 ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
        done

        log_info "Calculating AWS Cost for ${INSTANCE} (Spot: ${SPOT})..."
        if [[ "${SPOT}" == "true" ]]; then
            "${BIN_PATH}" aws -instance "${INSTANCE}" -count "${COUNT}" -hours "${HOURS}" -ebs "${EBS}" -egress "${EGRESS}" -spot
        else
            "${BIN_PATH}" aws -instance "${INSTANCE}" -count "${COUNT}" -hours "${HOURS}" -ebs "${EBS}" -egress "${EGRESS}"
        fi
        ;;

    azure)
        VM="Standard_B2s"
        COUNT=1
        HOURS=730
        DISK=50
        EGRESS=0
        SPOT="false"

        while [[ $# -gt 0 ]]; do
            case "$1" in
                --vm) VM="$2"; shift 2 ;;
                --count) COUNT="$2"; shift 2 ;;
                --hours) HOURS="$2"; shift 2 ;;
                --disk|--storage) DISK="$2"; shift 2 ;;
                --egress) EGRESS="$2"; shift 2 ;;
                --spot) SPOT="true"; shift 1 ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
        done

        log_info "Calculating Azure Cost for ${VM} (Spot: ${SPOT})..."
        if [[ "${SPOT}" == "true" ]]; then
            "${BIN_PATH}" azure -vm "${VM}" -count "${COUNT}" -hours "${HOURS}" -disk "${DISK}" -egress "${EGRESS}" -spot
        else
            "${BIN_PATH}" azure -vm "${VM}" -count "${COUNT}" -hours "${HOURS}" -disk "${DISK}" -egress "${EGRESS}"
        fi
        ;;

    gcp)
        MACHINE="e2-standard-2"
        COUNT=1
        HOURS=730
        DISK=50
        EGRESS=0
        SPOT="false"

        while [[ $# -gt 0 ]]; do
            case "$1" in
                --machine) MACHINE="$2"; shift 2 ;;
                --count) COUNT="$2"; shift 2 ;;
                --hours) HOURS="$2"; shift 2 ;;
                --disk|--storage) DISK="$2"; shift 2 ;;
                --egress) EGRESS="$2"; shift 2 ;;
                --spot) SPOT="true"; shift 1 ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
        done

        log_info "Calculating GCP Cost for ${MACHINE} (Spot: ${SPOT})..."
        if [[ "${SPOT}" == "true" ]]; then
            "${BIN_PATH}" gcp -machine "${MACHINE}" -count "${COUNT}" -hours "${HOURS}" -disk "${DISK}" -egress "${EGRESS}" -spot
        else
            "${BIN_PATH}" gcp -machine "${MACHINE}" -count "${COUNT}" -hours "${HOURS}" -disk "${DISK}" -egress "${EGRESS}"
        fi
        ;;

    catalog|pricing)
        PROVIDER=""
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --provider) PROVIDER="$2"; shift 2 ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
        done

        if [[ -n "${PROVIDER}" ]]; then
            "${BIN_PATH}" catalog -provider "${PROVIDER}"
        else
            "${BIN_PATH}" catalog
        fi
        ;;

    quick-quote)
        echo "=========================================================="
        echo "  ⚡ Multi-Cloud Quick Quote Wizard"
        echo "=========================================================="
        read -r -p "Enter CPU count (default 2): " q_cpu
        q_cpu="${q_cpu:-2}"
        read -r -p "Enter RAM in GB (default 4): " q_ram
        q_ram="${q_ram:-4}"
        read -r -p "Enter SSD Storage in GB (default 50): " q_storage
        q_storage="${q_storage:-50}"
        read -r -p "Enter Monthly Outbound Egress in GB (default 100): " q_egress
        q_egress="${q_egress:-100}"
        read -r -p "Use Spot/Preemptible discount? (y/N): " q_spot
        
        spot_flag=""
        if [[ "${q_spot}" =~ ^[Yy]$ ]]; then
            spot_flag="-spot"
        fi

        log_info "Running multi-cloud calculation..."
        "${BIN_PATH}" compare -vcpu "${q_cpu}" -ram "${q_ram}" -storage "${q_storage}" -egress "${q_egress}" ${spot_flag}
        ;;

    help|--help|-h)
        print_usage
        ;;

    *)
        log_error "Unknown command '${cmd}'"
        print_usage
        exit 1
        ;;
esac
