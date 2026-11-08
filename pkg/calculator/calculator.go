package calculator

import (
	"fmt"
	"math"
	"strings"

	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/catalog"
	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/models"
)

// CalculateNCP calculates estimated cost for Naver Cloud Platform.
func CalculateNCP(serverType string, count int, hours float64, storageGB float64, storageType string, outboundGB float64) (*models.CostCalculationResult, error) {
	if count <= 0 {
		count = 1
	}
	if hours <= 0 {
		hours = 730
	}
	if storageType == "" {
		storageType = "ssd"
	}

	exchangeRate := catalog.GetExchangeRate()
	result := &models.CostCalculationResult{
		Provider:     models.ProviderNCP,
		Service:      "Server & Storage",
		Region:       "KR (Korea)",
		Period:       fmt.Sprintf("%.0f hours (~1 month)", hours),
		ExchangeRate: exchangeRate,
		Breakdown:    make(map[string]models.CostDetail),
	}

	// 1. Compute
	serverRate := 92.0 // default s2-g2 (2vCPU, 4GB)
	serverName := "Standard s2-g2 (2vCPU, 4GB)"
	lowerType := strings.ToLower(serverType)
	switch {
	case strings.Contains(lowerType, "s4") || strings.Contains(lowerType, "4vcpu-8gb"):
		serverRate = 184.0
		serverName = "Standard s4-g2 (4vCPU, 8GB)"
	case strings.Contains(lowerType, "c4") || strings.Contains(lowerType, "high-cpu"):
		serverRate = 152.0
		serverName = "High CPU c4-g2 (4vCPU, 4GB)"
	case strings.Contains(lowerType, "m4") || strings.Contains(lowerType, "high-mem"):
		serverRate = 248.0
		serverName = "High Memory m4-g2 (4vCPU, 16GB)"
	}

	computeKRW := serverRate * hours * float64(count)
	result.Breakdown["compute"] = models.CostDetail{
		Units:       hours * float64(count),
		UnitName:    fmt.Sprintf("%s x %d (Server-Hours)", serverName, count),
		UnitPrice:   serverRate,
		Currency:    models.CurrencyKRW,
		SubtotalKRW: computeKRW,
		SubtotalUSD: computeKRW / exchangeRate,
	}

	// 2. Storage
	storageRate := 150.0 // SSD 150 KRW/GB-month
	if strings.ToLower(storageType) == "hdd" {
		storageRate = 80.0
	}
	storageKRW := storageRate * storageGB * float64(count) * (hours / 730.0)
	result.Breakdown["storage"] = models.CostDetail{
		Units:       storageGB * float64(count),
		UnitName:    fmt.Sprintf("Block Storage (%s, GB-Month)", strings.ToUpper(storageType)),
		UnitPrice:   storageRate,
		Currency:    models.CurrencyKRW,
		SubtotalKRW: storageKRW,
		SubtotalUSD: storageKRW / exchangeRate,
	}

	// 3. Network Outbound
	netKRW := 0.0
	billableGB := math.Max(0, outboundGB-1024.0) // 1TB free allowance for servers
	if billableGB > 0 {
		netKRW = billableGB * 100.0
	}
	result.Breakdown["network"] = models.CostDetail{
		Units:       outboundGB,
		UnitName:    "Outbound Internet Traffic (GB, 1TB free allowance)",
		UnitPrice:   100.0,
		Currency:    models.CurrencyKRW,
		SubtotalKRW: netKRW,
		SubtotalUSD: netKRW / exchangeRate,
	}

	totalKRW := computeKRW + storageKRW + netKRW
	result.TotalCostKRW = math.Round(totalKRW)
	result.TotalCostUSD = math.Round((totalKRW/exchangeRate)*100) / 100
	result.CostOptimizationTip = "💡 Tip: 1년 약정 시 10%, 2년 약정 시 20% 추가 할인을 적용할 수 있으며, 개발 환경은 미사용 시 서버 정지로 컴퓨팅 비용을 100% 절감할 수 있습니다."

	return result, nil
}

// CalculateCloudflare calculates estimated cost for Cloudflare.
func CalculateCloudflare(workersReqsMillions float64, r2StorageGB float64, r2ClassAMillions float64, r2ClassBMillions float64) (*models.CostCalculationResult, error) {
	exchangeRate := catalog.GetExchangeRate()
	result := &models.CostCalculationResult{
		Provider:     models.ProviderCloudflare,
		Service:      "Workers & R2 Storage",
		Region:       "Global Edge (Anycast 300+ cities)",
		Period:       "Monthly Usage",
		ExchangeRate: exchangeRate,
		Breakdown:    make(map[string]models.CostDetail),
	}

	// 1. Workers
	workersUSD := 0.0
	if workersReqsMillions > 0 {
		workersUSD = 5.0 // Base paid plan (includes 10M requests)
		if workersReqsMillions > 10.0 {
			workersUSD += (workersReqsMillions - 10.0) * 0.30
		}
	}
	result.Breakdown["workers"] = models.CostDetail{
		Units:       workersReqsMillions,
		UnitName:    "Workers Requests (Millions, Base 10M incl.)",
		UnitPrice:   0.30,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: workersUSD,
		SubtotalKRW: workersUSD * exchangeRate,
	}

	// 2. R2 Storage (10GB free)
	billableR2GB := math.Max(0, r2StorageGB-10.0)
	r2StorageUSD := billableR2GB * 0.015
	result.Breakdown["r2_storage"] = models.CostDetail{
		Units:       r2StorageGB,
		UnitName:    "R2 Standard Storage (GB-Month, 10GB free)",
		UnitPrice:   0.015,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: r2StorageUSD,
		SubtotalKRW: r2StorageUSD * exchangeRate,
	}

	// 3. R2 Operations (1M Class A free, 10M Class B free)
	billableClassA := math.Max(0, r2ClassAMillions-1.0)
	opsAUSD := billableClassA * 4.50
	billableClassB := math.Max(0, r2ClassBMillions-10.0)
	opsBUSD := billableClassB * 0.36
	opsUSD := opsAUSD + opsBUSD
	result.Breakdown["r2_operations"] = models.CostDetail{
		Units:       r2ClassAMillions + r2ClassBMillions,
		UnitName:    "R2 Class A/B Operations (Millions, with free tiers)",
		UnitPrice:   4.50,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: opsUSD,
		SubtotalKRW: opsUSD * exchangeRate,
	}

	// 4. Egress (Always free)
	result.Breakdown["egress"] = models.CostDetail{
		Units:       0,
		UnitName:    "Unlimited Internet Egress (Zero bandwidth egress fee)",
		UnitPrice:   0.0,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: 0.0,
		SubtotalKRW: 0.0,
	}

	totalUSD := workersUSD + r2StorageUSD + opsUSD
	result.TotalCostUSD = math.Round(totalUSD*100) / 100
	result.TotalCostKRW = math.Round(totalUSD * exchangeRate)
	result.CostOptimizationTip = "💡 Tip: Cloudflare R2는 AWS S3 대비 아웃바운드(Egress) 요금이 $0(완전 무료)이므로 비디오/데이터셋/대용량 트래픽 호스팅 시 최대 90% 이상 비용을 절약합니다."

	return result, nil
}

// CalculateAWS calculates estimated cost for Amazon Web Services.
func CalculateAWS(instanceType string, count int, hours float64, ebsGB float64, outboundGB float64, isSpot bool) (*models.CostCalculationResult, error) {
	if count <= 0 {
		count = 1
	}
	if hours <= 0 {
		hours = 730
	}
	exchangeRate := catalog.GetExchangeRate()
	result := &models.CostCalculationResult{
		Provider:     models.ProviderAWS,
		Service:      "Amazon EC2, EBS & Data Transfer",
		Region:       "ap-northeast-2 (Seoul)",
		Period:       fmt.Sprintf("%.0f hours (~1 month)", hours),
		ExchangeRate: exchangeRate,
		Breakdown:    make(map[string]models.CostDetail),
	}

	// 1. EC2 Compute
	hourlyRate := 0.0416 // default t4g.medium (2vCPU, 4GB Graviton2)
	instName := "t4g.medium (2vCPU, 4GB Graviton2)"
	lower := strings.ToLower(instanceType)
	switch {
	case strings.Contains(lower, "t4g.small") || strings.Contains(lower, "2vcpu-2gb"):
		hourlyRate = 0.0208
		instName = "t4g.small (2vCPU, 2GB)"
	case strings.Contains(lower, "t3.medium"):
		hourlyRate = 0.052
		instName = "t3.medium (2vCPU, 4GB x86)"
	case strings.Contains(lower, "m6i.xlarge") || strings.Contains(lower, "4vcpu-16gb"):
		hourlyRate = 0.2304
		instName = "m6i.xlarge (4vCPU, 16GB)"
	}

	if isSpot {
		hourlyRate *= 0.35 // ~65% spot discount
		instName += " [Spot ~65% DC]"
	}

	computeUSD := hourlyRate * hours * float64(count)
	result.Breakdown["compute"] = models.CostDetail{
		Units:       hours * float64(count),
		UnitName:    fmt.Sprintf("%s x %d (Instance-Hours)", instName, count),
		UnitPrice:   hourlyRate,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: computeUSD,
		SubtotalKRW: computeUSD * exchangeRate,
	}

	// 2. EBS gp3
	ebsUSD := ebsGB * 0.096 * float64(count) * (hours / 730.0)
	result.Breakdown["storage"] = models.CostDetail{
		Units:       ebsGB * float64(count),
		UnitName:    "EBS General Purpose SSD (gp3, GB-Month)",
		UnitPrice:   0.096,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: ebsUSD,
		SubtotalKRW: ebsUSD * exchangeRate,
	}

	// 3. Egress
	billableEgress := math.Max(0, outboundGB-100.0) // 100GB free tier
	netUSD := billableEgress * 0.114
	result.Breakdown["network"] = models.CostDetail{
		Units:       outboundGB,
		UnitName:    "Internet Egress (Seoul, 100GB Free Tier)",
		UnitPrice:   0.114,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: netUSD,
		SubtotalKRW: netUSD * exchangeRate,
	}

	totalUSD := computeUSD + ebsUSD + netUSD
	result.TotalCostUSD = math.Round(totalUSD*100) / 100
	result.TotalCostKRW = math.Round(totalUSD * exchangeRate)
	result.CostOptimizationTip = "💡 Tip: x86 대비 AWS Graviton(ARM) 인스턴스는 동일 성능 기준 최대 20% 저렴하며, 배치성 작업은 Spot 사용 시 최대 70% 비용을 절감할 수 있습니다."

	return result, nil
}

// CalculateAzure calculates estimated cost for Microsoft Azure.
func CalculateAzure(vmSize string, count int, hours float64, diskGB float64, outboundGB float64, isSpot bool) (*models.CostCalculationResult, error) {
	if count <= 0 {
		count = 1
	}
	if hours <= 0 {
		hours = 730
	}
	exchangeRate := catalog.GetExchangeRate()
	result := &models.CostCalculationResult{
		Provider:     models.ProviderAzure,
		Service:      "Azure Virtual Machines, Managed Disks & Bandwidth",
		Region:       "koreacentral (Seoul)",
		Period:       fmt.Sprintf("%.0f hours (~1 month)", hours),
		ExchangeRate: exchangeRate,
		Breakdown:    make(map[string]models.CostDetail),
	}

	// 1. VM Compute
	hourlyRate := 0.0496 // Standard_B2s
	vmName := "Standard_B2s (2vCPU, 4GB Burstable)"
	lower := strings.ToLower(vmSize)
	switch {
	case strings.Contains(lower, "d4s") || strings.Contains(lower, "4vcpu-16gb"):
		hourlyRate = 0.216
		vmName = "Standard_D4s_v5 (4vCPU, 16GB)"
	}

	if isSpot {
		hourlyRate *= 0.20 // ~80% spot discount
		vmName += " [Spot ~80% DC]"
	}

	computeUSD := hourlyRate * hours * float64(count)
	result.Breakdown["compute"] = models.CostDetail{
		Units:       hours * float64(count),
		UnitName:    fmt.Sprintf("%s x %d (VM-Hours)", vmName, count),
		UnitPrice:   hourlyRate,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: computeUSD,
		SubtotalKRW: computeUSD * exchangeRate,
	}

	// 2. Managed Disk (Standard/Premium SSD approx $0.10/GB)
	diskUSD := diskGB * 0.10 * float64(count) * (hours / 730.0)
	result.Breakdown["storage"] = models.CostDetail{
		Units:       diskGB * float64(count),
		UnitName:    "Azure Managed SSD (GB-Month)",
		UnitPrice:   0.10,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: diskUSD,
		SubtotalKRW: diskUSD * exchangeRate,
	}

	// 3. Egress
	billableEgress := math.Max(0, outboundGB-100.0)
	netUSD := billableEgress * 0.110
	result.Breakdown["network"] = models.CostDetail{
		Units:       outboundGB,
		UnitName:    "Internet Egress (Korea Central, 100GB Free Tier)",
		UnitPrice:   0.110,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: netUSD,
		SubtotalKRW: netUSD * exchangeRate,
	}

	totalUSD := computeUSD + diskUSD + netUSD
	result.TotalCostUSD = math.Round(totalUSD*100) / 100
	result.TotalCostKRW = math.Round(totalUSD * exchangeRate)
	result.CostOptimizationTip = "💡 Tip: Azure Hybrid Benefit(AHB) 보유 시 Windows/SQL 라이선스 비용을 최대 40% 절감할 수 있으며 AKS 마스터 노드는 기본 무료입니다."

	return result, nil
}

// CalculateGCP calculates estimated cost for Google Cloud Platform.
func CalculateGCP(machineType string, count int, hours float64, diskGB float64, outboundGB float64, isSpot bool) (*models.CostCalculationResult, error) {
	if count <= 0 {
		count = 1
	}
	if hours <= 0 {
		hours = 730
	}
	exchangeRate := catalog.GetExchangeRate()
	result := &models.CostCalculationResult{
		Provider:     models.ProviderGCP,
		Service:      "Google Compute Engine & Cloud Storage",
		Region:       "asia-northeast3 (Seoul)",
		Period:       fmt.Sprintf("%.0f hours (~1 month)", hours),
		ExchangeRate: exchangeRate,
		Breakdown:    make(map[string]models.CostDetail),
	}

	// 1. Compute Engine
	hourlyRate := 0.0804 // e2-standard-2 (2vCPU, 8GB)
	machName := "e2-standard-2 (2vCPU, 8GB)"
	lower := strings.ToLower(machineType)
	switch {
	case strings.Contains(lower, "e2-micro"):
		hourlyRate = 0.0084
		machName = "e2-micro (2vCPU shared, 1GB)"
	case strings.Contains(lower, "e2-standard-4") || strings.Contains(lower, "4vcpu-16gb"):
		hourlyRate = 0.1608
		machName = "e2-standard-4 (4vCPU, 16GB)"
	}

	if isSpot {
		hourlyRate *= 0.30 // ~70% spot discount
		machName += " [Spot ~70% DC]"
	}

	computeUSD := hourlyRate * hours * float64(count)
	result.Breakdown["compute"] = models.CostDetail{
		Units:       hours * float64(count),
		UnitName:    fmt.Sprintf("%s x %d (Instance-Hours)", machName, count),
		UnitPrice:   hourlyRate,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: computeUSD,
		SubtotalKRW: computeUSD * exchangeRate,
	}

	// 2. Persistent Disk
	diskUSD := diskGB * 0.12 * float64(count) * (hours / 730.0)
	result.Breakdown["storage"] = models.CostDetail{
		Units:       diskGB * float64(count),
		UnitName:    "Balanced Persistent Disk (pd-balanced, GB-Month)",
		UnitPrice:   0.12,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: diskUSD,
		SubtotalKRW: diskUSD * exchangeRate,
	}

	// 3. Egress
	netUSD := outboundGB * 0.085
	result.Breakdown["network"] = models.CostDetail{
		Units:       outboundGB,
		UnitName:    "Internet Egress (Standard Tier, Seoul)",
		UnitPrice:   0.085,
		Currency:    models.CurrencyUSD,
		SubtotalUSD: netUSD,
		SubtotalKRW: netUSD * exchangeRate,
	}

	totalUSD := computeUSD + diskUSD + netUSD
	result.TotalCostUSD = math.Round(totalUSD*100) / 100
	result.TotalCostKRW = math.Round(totalUSD * exchangeRate)
	result.CostOptimizationTip = "💡 Tip: Compute Engine 장기 약정(CUD 1년/3년) 시 37~55% 영구 할인되며, Cloud Run 등 서버리스 아키텍처 도입 시 유휴 비용을 0원으로 축소할 수 있습니다."

	return result, nil
}

// CompareAllClouds executes unified cross-cloud FinOps comparison.
func CompareAllClouds(req models.CompareRequest) (*models.CompareResponse, error) {
	if req.HoursPerMonth <= 0 {
		req.HoursPerMonth = 730
	}
	if req.StorageGB <= 0 {
		req.StorageGB = 50
	}

	// 1. NCP
	ncpRes, _ := CalculateNCP("s2-g2", 1, req.HoursPerMonth, req.StorageGB, req.StorageType, req.OutboundGB)
	ncpQuote := models.ProviderQuote{
		Provider:     models.ProviderNCP,
		InstanceName: "NCP Standard s2-g2 (2vCPU, 4GB)",
		ComputeCost:  ncpRes.Breakdown["compute"].SubtotalKRW,
		StorageCost:  ncpRes.Breakdown["storage"].SubtotalKRW,
		NetworkCost:  ncpRes.Breakdown["network"].SubtotalKRW,
		TotalCostKRW: ncpRes.TotalCostKRW,
		TotalCostUSD: ncpRes.TotalCostUSD,
		Notes:        "KR 리전 기본 원화 과금, 1TB 아웃바운드 무료 기본 제공",
	}

	// 2. AWS
	awsRes, _ := CalculateAWS("t4g.medium", 1, req.HoursPerMonth, req.StorageGB, req.OutboundGB, req.IsSpot)
	awsQuote := models.ProviderQuote{
		Provider:     models.ProviderAWS,
		InstanceName: "AWS EC2 t4g.medium (2vCPU, 4GB ARM Graviton2)",
		ComputeCost:  awsRes.Breakdown["compute"].SubtotalKRW,
		StorageCost:  awsRes.Breakdown["storage"].SubtotalKRW,
		NetworkCost:  awsRes.Breakdown["network"].SubtotalKRW,
		TotalCostKRW: awsRes.TotalCostKRW,
		TotalCostUSD: awsRes.TotalCostUSD,
		Notes:        "Seoul 리전 Graviton2 인스턴스 (최고 전력/비용 효율)",
	}

	// 3. Azure
	azureRes, _ := CalculateAzure("Standard_B2s", 1, req.HoursPerMonth, req.StorageGB, req.OutboundGB, req.IsSpot)
	azureQuote := models.ProviderQuote{
		Provider:     models.ProviderAzure,
		InstanceName: "Azure Standard_B2s (2vCPU, 4GB Burstable)",
		ComputeCost:  azureRes.Breakdown["compute"].SubtotalKRW,
		StorageCost:  azureRes.Breakdown["storage"].SubtotalKRW,
		NetworkCost:  azureRes.Breakdown["network"].SubtotalKRW,
		TotalCostKRW: azureRes.TotalCostKRW,
		TotalCostUSD: azureRes.TotalCostUSD,
		Notes:        "Korea Central 리전 Burstable VM",
	}

	// 4. GCP
	gcpRes, _ := CalculateGCP("e2-standard-2", 1, req.HoursPerMonth, req.StorageGB, req.OutboundGB, req.IsSpot)
	gcpQuote := models.ProviderQuote{
		Provider:     models.ProviderGCP,
		InstanceName: "GCP e2-standard-2 (2vCPU, 8GB)",
		ComputeCost:  gcpRes.Breakdown["compute"].SubtotalKRW,
		StorageCost:  gcpRes.Breakdown["storage"].SubtotalKRW,
		NetworkCost:  gcpRes.Breakdown["network"].SubtotalKRW,
		TotalCostKRW: gcpRes.TotalCostKRW,
		TotalCostUSD: gcpRes.TotalCostUSD,
		Notes:        "Seoul 리전 균형 잡힌 범용 VM (8GB RAM 탑재)",
	}

	quotes := []models.ProviderQuote{ncpQuote, awsQuote, azureQuote, gcpQuote}

	// Find best option (lowest TotalCostKRW)
	best := quotes[0]
	for _, q := range quotes[1:] {
		if q.TotalCostKRW < best.TotalCostKRW {
			best = q
		}
	}

	advice := fmt.Sprintf("🏆 가장 비용 효율적인 옵션은 [%s] (%s, 월 약 %.0f원 / $%.2f) 입니다. 장기 안정 운영 시 약정(1-3년)을 추가 결합하면 추가 20~50%% 절감이 가능합니다.",
		strings.ToUpper(string(best.Provider)), best.InstanceName, best.TotalCostKRW, best.TotalCostUSD)

	return &models.CompareResponse{
		SpecRequested: req,
		Quotes:        quotes,
		BestOption:    best,
		SavingsAdvice: advice,
	}, nil
}
