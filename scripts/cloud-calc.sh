#!/usr/bin/env bash
# ==============================================================================
# cloud-calc.sh: Big 3 Cloud Pricing Calculator Universal CLI (AWS, Azure, GCP)
# Adheres strictly to Limmireu Global Rule (set -euo pipefail & clear logging)
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="${SCRIPT_DIR}/../bin/cloud-pricing-mcp"

# Fallback to system local bin if not in project bin
if [[ ! -x "${BIN_PATH}" ]]; then
    if [[ -x "${HOME}/.gemini/antigravity-cli/bin/cloud-pricing-mcp" ]]; then
        BIN_PATH="${HOME}/.gemini/antigravity-cli/bin/cloud-pricing-mcp"
    elif [[ -x "${HOME}/.local/bin/cloud-pricing-mcp" ]]; then
        BIN_PATH="${HOME}/.local/bin/cloud-pricing-mcp"
    else
        echo "[ERROR] cloud-pricing-mcp binary not found! Please build it first." >&2
        exit 1
    fi
fi

log_info() {
    echo -e "\033[1;34m[INFO]\033[0m $*"
}

log_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $*" >&2
}

print_usage() {
    cat << 'EOF'
================================================================================
  Big 3 Cloud Pricing Calculator CLI (AWS, Azure, GCP)
================================================================================
Usage:
  ./cloud-calc.sh [command] [options]

Commands:
  compare       Cross-cloud cost comparison (AWS vs Azure vs GCP)
  aws           Calculate AWS EC2, EBS & Egress cost
  azure         Calculate Azure VM, Disk & Egress cost
  gcp           Calculate Google Cloud Compute Engine & Storage cost
  catalog       View full pricing catalog or filter by provider (aws, azure, gcp)
  quick-quote   Interactive wizard for instant Big 3 cost recommendation
  help          Show this help message

Quick Examples:
  # 1. Compare Big 3 for 2 vCPU, 4GB RAM, 50GB SSD, 100GB Outbound
  ./cloud-calc.sh compare --vcpu 2 --ram 4 --storage 50 --egress 100

  # 2. AWS EC2 Graviton2 Spot instance
  ./cloud-calc.sh aws --instance t4g.medium --ebs 100 --spot

  # 3. Azure Standard_B2s VM
  ./cloud-calc.sh azure --vm Standard_B2s --disk 100

  # 4. GCP e2-standard-2 instance
  ./cloud-calc.sh gcp --machine e2-standard-2 --disk 100

  # 5. Show GCP catalog products
  ./cloud-calc.sh catalog --provider gcp
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

        log_info "Calculating Big 3 comparison: ${VCPU} vCPU, ${RAM}GB RAM, ${STORAGE}GB ${STORAGE_TYPE^^}, ${EGRESS}GB Egress (Spot: ${SPOT})..."
        if [[ "${SPOT}" == "true" ]]; then
            "${BIN_PATH}" compare -vcpu "${VCPU}" -ram "${RAM}" -storage "${STORAGE}" -storage-type "${STORAGE_TYPE}" -hours "${HOURS}" -egress "${EGRESS}" -spot
        else
            "${BIN_PATH}" compare -vcpu "${VCPU}" -ram "${RAM}" -storage "${STORAGE}" -storage-type "${STORAGE_TYPE}" -hours "${HOURS}" -egress "${EGRESS}"
        fi
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
        echo "  ⚡ Big 3 Cloud Quick Quote Wizard (AWS vs Azure vs GCP)"
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

        log_info "Running Big 3 calculation..."
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
