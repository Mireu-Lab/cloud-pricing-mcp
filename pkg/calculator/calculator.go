package calculator

import (
	"fmt"
	"math"
	"strings"

	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/catalog"
	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/models"
)

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
	case strings.Contains(lower, "g4dn") || strings.Contains(lower, "gpu"):
		hourlyRate = 0.710
		instName = "g4dn.xlarge (4vCPU, 16GB, 1x T4)"
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
	result.CostOptimizationTip = "💡 Tip: AWS Graviton(ARM) 인스턴스는 동일 성능 대비 최대 20% 저렴하며, 배치성 작업은 Spot 사용 시 최대 65~70% 비용을 절감할 수 있습니다."

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
	case strings.Contains(lower, "nc4as") || strings.Contains(lower, "gpu"):
		hourlyRate = 0.628
		vmName = "Standard_NC4as_T4_v3 (4vCPU, 28GB, 1x T4)"
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
	result.CostOptimizationTip = "💡 Tip: Azure Spot 인스턴스는 최대 80-90% 저렴하며 AKS 마스터 노드는 기본 무료입니다."

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
	case strings.Contains(lower, "g2-standard-4") || strings.Contains(lower, "gpu"):
		hourlyRate = 0.758
		machName = "g2-standard-4 (4vCPU, 16GB, 1x L4)"
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
	result.CostOptimizationTip = "💡 Tip: GCS Archive 스토리지는 지연시간 0초(밀리초 즉시 인출)를 지원하며, Compute Engine CUD(1~3년) 시 최대 55% 할인됩니다."

	return result, nil
}

// CompareBig3 executes cross-cloud comparison between AWS, Azure, and GCP.
func CompareBig3(req models.CompareRequest) (*models.CompareResponse, error) {
	if req.HoursPerMonth <= 0 {
		req.HoursPerMonth = 730
	}
	if req.StorageGB <= 0 {
		req.StorageGB = 50
	}

	// 1. AWS
	awsRes, _ := CalculateAWS("t4g.medium", 1, req.HoursPerMonth, req.StorageGB, req.OutboundGB, req.IsSpot)
	awsQuote := models.ProviderQuote{
		Provider:     models.ProviderAWS,
		InstanceName: "AWS EC2 t4g.medium (2vCPU, 4GB ARM Graviton2)",
		ComputeCost:  awsRes.Breakdown["compute"].SubtotalKRW,
		StorageCost:  awsRes.Breakdown["storage"].SubtotalKRW,
		NetworkCost:  awsRes.Breakdown["network"].SubtotalKRW,
		TotalCostKRW: awsRes.TotalCostKRW,
		TotalCostUSD: awsRes.TotalCostUSD,
		Notes:        "Seoul region Graviton2 ARM instance (highest energy/cost efficiency)",
	}

	// 2. Azure
	azureRes, _ := CalculateAzure("Standard_B2s", 1, req.HoursPerMonth, req.StorageGB, req.OutboundGB, req.IsSpot)
	azureQuote := models.ProviderQuote{
		Provider:     models.ProviderAzure,
		InstanceName: "Azure Standard_B2s (2vCPU, 4GB Burstable)",
		ComputeCost:  azureRes.Breakdown["compute"].SubtotalKRW,
		StorageCost:  azureRes.Breakdown["storage"].SubtotalKRW,
		NetworkCost:  azureRes.Breakdown["network"].SubtotalKRW,
		TotalCostKRW: azureRes.TotalCostKRW,
		TotalCostUSD: azureRes.TotalCostUSD,
		Notes:        "Korea Central region burstable VM",
	}

	// 3. GCP
	gcpRes, _ := CalculateGCP("e2-standard-2", 1, req.HoursPerMonth, req.StorageGB, req.OutboundGB, req.IsSpot)
	gcpQuote := models.ProviderQuote{
		Provider:     models.ProviderGCP,
		InstanceName: "GCP e2-standard-2 (2vCPU, 8GB)",
		ComputeCost:  gcpRes.Breakdown["compute"].SubtotalKRW,
		StorageCost:  gcpRes.Breakdown["storage"].SubtotalKRW,
		NetworkCost:  gcpRes.Breakdown["network"].SubtotalKRW,
		TotalCostKRW: gcpRes.TotalCostKRW,
		TotalCostUSD: gcpRes.TotalCostUSD,
		Notes:        "Seoul region balanced general VM with 8GB RAM",
	}

	quotes := []models.ProviderQuote{awsQuote, azureQuote, gcpQuote}

	// Find best option
	best := quotes[0]
	for _, q := range quotes[1:] {
		if q.TotalCostKRW < best.TotalCostKRW {
			best = q
		}
	}

	advice := fmt.Sprintf("🏆 Lowest cost option is [%s] (%s, ~%.0f KRW / $%.2f monthly). Commitments (1-3 yrs) offer additional 20-50%% savings.",
		strings.ToUpper(string(best.Provider)), best.InstanceName, best.TotalCostKRW, best.TotalCostUSD)

	return &models.CompareResponse{
		SpecRequested: req,
		Quotes:        quotes,
		BestOption:    best,
		SavingsAdvice: advice,
	}, nil
}
