# Cloud Pricing MCP & CLI (Multi-Cloud FinOps Engine)

[![Go Version](https://img.shields.io/badge/go-1.24%2B-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![MCP Compatible](https://img.shields.io/badge/MCP-2024--11--05-orange.svg)](https://modelcontextprotocol.io)

고성능 **멀티 클라우드 비용 계산기 MCP(Model Context Protocol) 서버 및 CLI 도구**입니다.
NCP(네이버 클라우드 플랫폼), Cloudflare, AWS, Azure, Google Cloud(GCP)의 컴퓨팅, 스토리지, 네트워크 비용을 고속으로 산출하고 동일 사양에 대한 최저가 비교(FinOps)를 수행합니다.

---

## 🌟 주요 특징

1. **이원화된 계산 인터페이스 (통합 + 개별)**:
   - `compare_cloud_costs`: 동일 vCPU/RAM/스토리지/트래픽에 대해 **5대 클라우드 최저가 자동 랭킹 비교**
   - 프로바이더별 개별 툴: `calculate_ncp`, `calculate_cloudflare`, `calculate_aws`, `calculate_azure`, `calculate_gcp`
2. **LLM 친화적 MCP Server (Stdio)**:
   - Claude Desktop, Antigravity, Cursor, Zed 등 모든 MCP 호환 에이전트에 즉시 플러그인 가능
3. **독립 고속 연산 (Zero API Key 오프라인 지원)**:
   - 외부 API 호출 지연 없이 내장된 단가 카탈로그를 기반으로 밀리초(ms) 단위의 견적 산출
4. **한국 리전 최적화 및 환율 변환**:
   - 원화(KRW) 및 달러(USD) 동시 산출
   - 서울 리전(NCP KR, AWS `ap-northeast-2`, Azure `koreacentral`, GCP `asia-northeast3`) 기준

---

## 🛠 지원 프로바이더 및 서비스

| Provider | 지원 서비스 | 특징 |
| :--- | :--- | :--- |
| **NCP (Naver Cloud)** | Server VPC (s2, s4, c4, m4), SSD/HDD Block Storage, Outbound Traffic, Cloud DB, NKS | 1TB 무료 아웃바운드, 원화(KRW) 과금 |
| **Cloudflare** | Workers (Paid base + Extra), R2 Object Storage (Standard, Infrequent Access), Class A/B Ops | **Egress 100% 무료**, 10GB R2 Free Tier |
| **AWS** | Amazon EC2 (t4g Graviton2, t3, m6i), EBS gp3 SSD, Internet Egress | Spot 최대 65% 할인, 100GB Free Egress |
| **Azure** | Azure VMs (B-series, D-series v5), Managed SSD, Internet Egress | Spot 최대 80% 할인, AKS 마스터 무료 |
| **GCP** | Compute Engine (e2-micro, e2-standard), pd-balanced SSD, Egress | Spot 최대 70% 할인, 장기 약정 할인 |

---

## 📦 설치 및 빌드

```bash
git clone https://github.com/Mireu-Lab/cloud-pricing-mcp.git
cd cloud-pricing-mcp
go build -o bin/cloud-pricing-mcp ./cmd/cloud-pricing-mcp
```

---

## 💻 CLI 사용법

### 1. 멀티 클라우드 최저가 비교 (Universal Compare)
```bash
./bin/cloud-pricing-mcp compare -vcpu 2 -ram 4 -storage 50 -egress 100
```

### 2. 프로바이더별 개별 견적
```bash
# NCP (2vCPU 4GB Server + 100GB SSD)
./bin/cloud-pricing-mcp ncp -server s2-g2 -storage 100 -storage-type ssd

# Cloudflare (Workers 25M requests + 1TB R2 storage)
./bin/cloud-pricing-mcp cf -workers-reqs-m 25 -r2-storage-gb 1000

# AWS (EC2 Graviton2 Spot 인스턴스)
./bin/cloud-pricing-mcp aws -instance t4g.medium -ebs 100 -spot

# Azure (Standard_B2s VM)
./bin/cloud-pricing-mcp azure -vm Standard_B2s -disk 100

# GCP (e2-standard-2)
./bin/cloud-pricing-mcp gcp -machine e2-standard-2 -disk 100
```

---

## 🤖 MCP(Model Context Protocol) 연동 가이드

`mcp_config.json` (또는 Claude Desktop 설정)에 등록:

```json
{
  "mcpServers": {
    "cloud-pricing": {
      "command": "/path/to/bin/cloud-pricing-mcp",
      "args": ["server"]
    }
  }
}
```

---

## 👥 프로젝트 거버넌스
* **프로젝트 매니저 (Project Manager)**: limmireu1214@gmail.com
* **구현 책임자 (Implementation Owner)**: limmireu1214@gmail.com
* **저장소**: [https://github.com/Mireu-Lab/cloud-pricing-mcp](https://github.com/Mireu-Lab/cloud-pricing-mcp)

---

## 📄 License
MIT License
