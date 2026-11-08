package models

// Provider represents supported cloud providers (Big 3: AWS, Azure, GCP).
type Provider string

const (
	ProviderAWS   Provider = "aws"
	ProviderAzure Provider = "azure"
	ProviderGCP   Provider = "gcp"
)

// Currency represents currency code.
type Currency string

const (
	CurrencyKRW Currency = "KRW"
	CurrencyUSD Currency = "USD"
)

// PriceItem represents a normalized cloud product pricing item.
type PriceItem struct {
	ID                  string   `json:"id"`
	Provider            Provider `json:"provider"`
	Category            string   `json:"category"`
	MainProduct         string   `json:"main_product"`
	SubService          string   `json:"sub_service"`
	ServiceType         string   `json:"service_type"`
	Description         string   `json:"description"`
	PricingUnit         string   `json:"pricing_unit"`
	PricingModel        string   `json:"pricing_model"`
	Currency            Currency `json:"currency"`
	BaseRate            float64  `json:"base_rate"`
	FreeTierAllowance   string   `json:"free_tier_allowance,omitempty"`
	CostOptimizationTip string   `json:"cost_optimization_tip,omitempty"`
}

// CostCalculationResult represents the calculation output.
type CostCalculationResult struct {
	Provider            Provider              `json:"provider"`
	Service             string                `json:"service"`
	Region              string                `json:"region"`
	Period              string                `json:"period"`
	TotalCostUSD        float64               `json:"total_cost_usd"`
	TotalCostKRW        float64               `json:"total_cost_krw"`
	ExchangeRate        float64               `json:"exchange_rate_usd_krw"`
	Breakdown           map[string]CostDetail `json:"breakdown"`
	CostOptimizationTip string                `json:"cost_optimization_tip,omitempty"`
}

// CostDetail breakdown item.
type CostDetail struct {
	Units       float64  `json:"units"`
	UnitName    string   `json:"unit_name"`
	UnitPrice   float64  `json:"unit_price"`
	Currency    Currency `json:"currency"`
	SubtotalUSD float64  `json:"subtotal_usd"`
	SubtotalKRW float64  `json:"subtotal_krw"`
}

// CompareRequest represents cross-cloud comparison request.
type CompareRequest struct {
	VCPU             int     `json:"vcpu"`
	MemoryGB         float64 `json:"memory_gb"`
	StorageGB        float64 `json:"storage_gb"`
	StorageType      string  `json:"storage_type"` // ssd, hdd
	HoursPerMonth    float64 `json:"hours_per_month"`
	OutboundGB       float64 `json:"outbound_gb"`
	IsSpot           bool    `json:"is_spot"`
	CommitmentPeriod string  `json:"commitment_period"` // none, 1-year, 3-year
}

// ProviderQuote comparison result for a single provider.
type ProviderQuote struct {
	Provider     Provider `json:"provider"`
	InstanceName string   `json:"instance_name"`
	ComputeCost  float64  `json:"compute_cost_krw"`
	StorageCost  float64  `json:"storage_cost_krw"`
	NetworkCost  float64  `json:"network_cost_krw"`
	TotalCostKRW float64  `json:"total_cost_krw"`
	TotalCostUSD float64  `json:"total_cost_usd"`
	Notes        string   `json:"notes,omitempty"`
}

// CompareResponse represents ranking and comparison across all clouds.
type CompareResponse struct {
	SpecRequested CompareRequest  `json:"spec_requested"`
	Quotes        []ProviderQuote `json:"quotes"`
	BestOption    ProviderQuote   `json:"best_option"`
	SavingsAdvice string          `json:"savings_advice"`
}
