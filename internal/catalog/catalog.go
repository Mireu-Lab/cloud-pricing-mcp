package catalog

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/models"
)

// Default USD to KRW rate fallback.
const DefaultUSDToKRW = 1350.0

type CatalogData struct {
	UpdatedAt        string             `json:"updated_at"`
	Version          string             `json:"version"`
	CurrencyExchange map[string]float64 `json:"currency_exchange"`
	Products         []models.PriceItem `json:"products"`
}

var (
	once sync.Once
	data CatalogData
)

func initCatalog() {
	// Initialize Big 3 (AWS, Azure, GCP) pricing catalog
	data = CatalogData{
		UpdatedAt: "2026-11-08T00:00:00Z",
		Version:   "2.0.0-big3",
		CurrencyExchange: map[string]float64{
			"USD_TO_KRW": DefaultUSDToKRW,
		},
		Products: getDefaultCatalogProducts(),
	}
}

func GetExchangeRate() float64 {
	once.Do(initCatalog)
	if rate, ok := data.CurrencyExchange["USD_TO_KRW"]; ok && rate > 0 {
		return rate
	}
	return DefaultUSDToKRW
}

func GetAllProducts() []models.PriceItem {
	once.Do(initCatalog)
	return data.Products
}

func GetProductsByProvider(provider models.Provider) []models.PriceItem {
	once.Do(initCatalog)
	var filtered []models.PriceItem
	for _, p := range data.Products {
		if p.Provider == provider {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func FindProductByID(id string) (*models.PriceItem, error) {
	once.Do(initCatalog)
	for _, p := range data.Products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("product ID %s not found", id)
}

func getDefaultCatalogProducts() []models.PriceItem {
	return []models.PriceItem{
		// ==================== AWS (Amazon Web Services) ====================
		{
			ID:                  "aws-ec2-t4g-small",
			Provider:            models.ProviderAWS,
			Category:            "Compute",
			MainProduct:         "Amazon EC2",
			SubService:          "t4g.small (2vCPU, 2GB ARM Graviton2)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "Arm-based Graviton2 burstable general purpose instance (Seoul)",
			PricingUnit:         "USD/hour",
			PricingModel:        "On-Demand",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0208,
			CostOptimizationTip: "Spot saves up to 70%, Compute Savings Plans saves ~35%",
		},
		{
			ID:                  "aws-ec2-t4g-medium",
			Provider:            models.ProviderAWS,
			Category:            "Compute",
			MainProduct:         "Amazon EC2",
			SubService:          "t4g.medium (2vCPU, 4GB ARM Graviton2)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "Arm-based Graviton2 burstable general purpose instance (Seoul)",
			PricingUnit:         "USD/hour",
			PricingModel:        "On-Demand",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0416,
			CostOptimizationTip: "Switch to Graviton for 20% better price-performance than x86",
		},
		{
			ID:                  "aws-ec2-t3-medium",
			Provider:            models.ProviderAWS,
			Category:            "Compute",
			MainProduct:         "Amazon EC2",
			SubService:          "t3.medium (2vCPU, 4GB x86)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "General purpose x86 instance (Seoul)",
			PricingUnit:         "USD/hour",
			PricingModel:        "On-Demand",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.052,
			CostOptimizationTip: "Migrate to t4g.medium to save 20%",
		},
		{
			ID:                  "aws-ec2-m6i-xlarge",
			Provider:            models.ProviderAWS,
			Category:            "Compute",
			MainProduct:         "Amazon EC2",
			SubService:          "m6i.xlarge (4vCPU, 16GB)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "General purpose Intel Xeon scalable instance (Seoul)",
			PricingUnit:         "USD/hour",
			PricingModel:        "On-Demand",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.2304,
			CostOptimizationTip: "Spot instance discount ~65% ($0.08/hr)",
		},
		{
			ID:                  "aws-ec2-g4dn-xlarge",
			Provider:            models.ProviderAWS,
			Category:            "Compute",
			MainProduct:         "Amazon EC2",
			SubService:          "g4dn.xlarge (4vCPU, 16GB, 1x T4 16GB)",
			ServiceType:         "GPU Accelerated Compute",
			Description:         "NVIDIA T4 GPU accelerated instance (Seoul)",
			PricingUnit:         "USD/hour",
			PricingModel:        "On-Demand",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.710,
			CostOptimizationTip: "Spot discount saves ~70% ($0.21/hr)",
		},
		{
			ID:                  "aws-s3-standard",
			Provider:            models.ProviderAWS,
			Category:            "Storage",
			MainProduct:         "Amazon S3",
			SubService:          "S3 Standard",
			ServiceType:         "Object Storage",
			Description:         "High durability object storage for frequently accessed data (Seoul)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity Tiered",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.025,
			FreeTierAllowance:   "5 GB S3 standard free for 12 months",
			CostOptimizationTip: "Use S3 Intelligent-Tiering to auto-archive cold data",
		},
		{
			ID:                  "aws-s3-glacier-deep",
			Provider:            models.ProviderAWS,
			Category:            "Storage",
			MainProduct:         "Amazon S3",
			SubService:          "S3 Glacier Deep Archive",
			ServiceType:         "Archive Storage",
			Description:         "Lowest-cost archive storage (retrieval within 12 hours)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.002, // $0.002/GB
			CostOptimizationTip: "Ideal for long-term compliance retention",
		},
		{
			ID:                  "aws-ebs-gp3",
			Provider:            models.ProviderAWS,
			Category:            "Storage",
			MainProduct:         "Amazon EBS",
			SubService:          "General Purpose SSD (gp3)",
			ServiceType:         "Block Storage",
			Description:         "Next-generation general-purpose SSD volume (Seoul)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.096, // 3,000 IOPS & 125 MB/s included
			CostOptimizationTip: "20% cheaper than gp2 with baseline 3000 IOPS",
		},
		{
			ID:                  "aws-net-outbound",
			Provider:            models.ProviderAWS,
			Category:            "Networking",
			MainProduct:         "Data Transfer",
			SubService:          "Outbound Data Transfer (Internet)",
			ServiceType:         "Network Egress",
			Description:         "Egress traffic to internet from Seoul region",
			PricingUnit:         "USD/GB",
			PricingModel:        "Usage Tiered",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.114, // Seoul outbound to internet ~$0.114/GB
			FreeTierAllowance:   "100 GB per month free egress",
			CostOptimizationTip: "Use CloudFront CDN for cheaper egress rates",
		},
		{
			ID:                  "aws-eks-cluster",
			Provider:            models.ProviderAWS,
			Category:            "Containers",
			MainProduct:         "Amazon EKS",
			SubService:          "EKS Cluster Management",
			ServiceType:         "Managed Kubernetes",
			Description:         "Control plane cluster management fee",
			PricingUnit:         "USD/hour",
			PricingModel:        "Hourly Cluster Fee",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.10, // $73/month
		},

		// ==================== AZURE (Microsoft Azure) ====================
		{
			ID:                  "azure-vm-b2s",
			Provider:            models.ProviderAzure,
			Category:            "Compute",
			MainProduct:         "Virtual Machines",
			SubService:          "Standard_B2s (2vCPU, 4GB)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "Burstable general purpose B-series VM (Korea Central)",
			PricingUnit:         "USD/hour",
			PricingModel:        "Pay-as-you-go",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0496,
			CostOptimizationTip: "1-yr reserved saves ~40%, 3-yr reserved saves ~60%",
		},
		{
			ID:                  "azure-vm-d4s-v5",
			Provider:            models.ProviderAzure,
			Category:            "Compute",
			MainProduct:         "Virtual Machines",
			SubService:          "Standard_D4s_v5 (4vCPU, 16GB)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "General purpose D-series v5 VM (Korea Central)",
			PricingUnit:         "USD/hour",
			PricingModel:        "Pay-as-you-go",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.216,
			CostOptimizationTip: "Azure Spot instances save up to 80-90%",
		},
		{
			ID:                  "azure-vm-nc4as-t4-v3",
			Provider:            models.ProviderAzure,
			Category:            "Compute",
			MainProduct:         "Virtual Machines",
			SubService:          "Standard_NC4as_T4_v3 (4vCPU, 28GB, 1x T4 16GB)",
			ServiceType:         "GPU Accelerated VM",
			Description:         "NVIDIA Tesla T4 GPU VM for AI inference and graphics",
			PricingUnit:         "USD/hour",
			PricingModel:        "Pay-as-you-go",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.628,
		},
		{
			ID:                  "azure-blob-hot",
			Provider:            models.ProviderAzure,
			Category:            "Storage",
			MainProduct:         "Azure Blob Storage",
			SubService:          "Hot Tier (LRS)",
			ServiceType:         "Object Storage",
			Description:         "Frequently accessed unstructured data storage (Korea Central)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0208,
			CostOptimizationTip: "Use Cool ($0.01) or Archive ($0.0018) for inactive data",
		},
		{
			ID:                  "azure-blob-archive",
			Provider:            models.ProviderAzure,
			Category:            "Storage",
			MainProduct:         "Azure Blob Storage",
			SubService:          "Archive Tier (LRS)",
			ServiceType:         "Archive Storage",
			Description:         "Ultra low-cost storage for rarely accessed data (Korea Central)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0018, // $0.0018/GB
		},
		{
			ID:                  "azure-disk-premium-ssd",
			Provider:            models.ProviderAzure,
			Category:            "Storage",
			MainProduct:         "Azure Managed Disks",
			SubService:          "Premium SSD (P10 128GB)",
			ServiceType:         "Managed Block Storage",
			Description:         "High performance SSD disk for production VMs",
			PricingUnit:         "USD/month",
			PricingModel:        "Disk Unit",
			Currency:            models.CurrencyUSD,
			BaseRate:            21.76, // P10 128GB ~$21.76/month
			CostOptimizationTip: "Standard SSD is adequate for non-production workloads",
		},
		{
			ID:                  "azure-net-egress",
			Provider:            models.ProviderAzure,
			Category:            "Networking",
			MainProduct:         "Bandwidth",
			SubService:          "Internet Egress",
			ServiceType:         "Network Egress",
			Description:         "Outbound data transfer from Korea Central",
			PricingUnit:         "USD/GB",
			PricingModel:        "Tiered Usage",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.110,
			FreeTierAllowance:   "First 100 GB/month free",
		},
		{
			ID:                  "azure-aks-cluster",
			Provider:            models.ProviderAzure,
			Category:            "Containers",
			MainProduct:         "Azure Kubernetes Service (AKS)",
			SubService:          "Free Cluster Management",
			ServiceType:         "Managed Kubernetes",
			Description:         "Free tier cluster management (Standard SLA optional at $0.10/hr)",
			PricingUnit:         "USD/hour",
			PricingModel:        "Free Baseline",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.00,
			FreeTierAllowance:   "Cluster management is 100% free by default",
			CostOptimizationTip: "Pay only for VM worker nodes, unlike EKS/GKE master fees",
		},

		// ==================== GCP (Google Cloud Platform) ====================
		{
			ID:                  "gcp-ce-e2-micro",
			Provider:            models.ProviderGCP,
			Category:            "Compute",
			MainProduct:         "Compute Engine",
			SubService:          "e2-micro (2vCPU, 1GB shared)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "Shared-core machine type for low-cost workloads",
			PricingUnit:         "USD/hour",
			PricingModel:        "Pay-as-you-go",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0084,
			FreeTierAllowance:   "1 free e2-micro instance in US regions",
		},
		{
			ID:                  "gcp-ce-e2-standard-2",
			Provider:            models.ProviderGCP,
			Category:            "Compute",
			MainProduct:         "Compute Engine",
			SubService:          "e2-standard-2 (2vCPU, 8GB)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "General purpose balanced VM (asia-northeast3 Seoul)",
			PricingUnit:         "USD/hour",
			PricingModel:        "Pay-as-you-go",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0804,
			CostOptimizationTip: "Spot discount saves 60-91%",
		},
		{
			ID:                  "gcp-ce-e2-standard-4",
			Provider:            models.ProviderGCP,
			Category:            "Compute",
			MainProduct:         "Compute Engine",
			SubService:          "e2-standard-4 (4vCPU, 16GB)",
			ServiceType:         "IaaS (Virtual Machines)",
			Description:         "General purpose balanced VM (asia-northeast3 Seoul)",
			PricingUnit:         "USD/hour",
			PricingModel:        "Pay-as-you-go",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.1608,
			CostOptimizationTip: "CUD 1-year saves 37%, 3-year saves 55%",
		},
		{
			ID:                  "gcp-ce-g2-standard-4",
			Provider:            models.ProviderGCP,
			Category:            "Compute",
			MainProduct:         "Compute Engine",
			SubService:          "g2-standard-4 (4vCPU, 16GB, 1x NVIDIA L4 24GB)",
			ServiceType:         "GPU Accelerated Compute",
			Description:         "Next-generation NVIDIA L4 GPU instance for AI inference and graphics",
			PricingUnit:         "USD/hour",
			PricingModel:        "Pay-as-you-go",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.758,
			CostOptimizationTip: "Spot saves ~65% ($0.26/hr)",
		},
		{
			ID:                  "gcp-gcs-standard",
			Provider:            models.ProviderGCP,
			Category:            "Storage",
			MainProduct:         "Cloud Storage",
			SubService:          "Standard Class",
			ServiceType:         "Object Storage",
			Description:         "Unstructured object storage (asia-northeast3 Seoul)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.023,
			CostOptimizationTip: "Coldline ($0.007) or Archive ($0.0025) for infrequently accessed files",
		},
		{
			ID:                  "gcp-gcs-archive",
			Provider:            models.ProviderGCP,
			Category:            "Storage",
			MainProduct:         "Cloud Storage",
			SubService:          "Archive Class (Instant Retrieval)",
			ServiceType:         "Archive Storage",
			Description:         "Lowest-cost archive storage with zero wait time retrieval (Seoul)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0025, // $0.0025/GB
			CostOptimizationTip: "Instant millisecond access without retrieval wait times",
		},
		{
			ID:                  "gcp-pd-balanced",
			Provider:            models.ProviderGCP,
			Category:            "Storage",
			MainProduct:         "Persistent Disk",
			SubService:          "Balanced Persistent Disk (pd-balanced)",
			ServiceType:         "Block Storage",
			Description:         "Balanced performance and cost SSD persistent disk (Seoul)",
			PricingUnit:         "USD/GB-month",
			PricingModel:        "Capacity",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.12,
			CostOptimizationTip: "Standard PD (pd-standard) is $0.048/GB for low IOPS needs",
		},
		{
			ID:                  "gcp-cloud-run-cpu",
			Provider:            models.ProviderGCP,
			Category:            "Compute",
			MainProduct:         "Cloud Run",
			SubService:          "vCPU Allocation",
			ServiceType:         "Serverless Container Execution",
			Description:         "Cloud Run Tier 1 CPU allocation per second (Seoul)",
			PricingUnit:         "USD/vCPU-second",
			PricingModel:        "Per-second pay-per-use",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.0000288,
			FreeTierAllowance:   "180,000 vCPU-seconds free per month",
			CostOptimizationTip: "Scale to 0 when idle to eliminate all compute costs",
		},
		{
			ID:                  "gcp-net-egress",
			Provider:            models.ProviderGCP,
			Category:            "Networking",
			MainProduct:         "Network Egress",
			SubService:          "Internet Egress (Standard Tier)",
			ServiceType:         "Network Egress",
			Description:         "Internet egress from Seoul using Standard Tier",
			PricingUnit:         "USD/GB",
			PricingModel:        "Usage",
			Currency:            models.CurrencyUSD,
			BaseRate:            0.085,
			FreeTierAllowance:   "Inbound traffic (Ingress) is 100% free",
		},
	}
}
