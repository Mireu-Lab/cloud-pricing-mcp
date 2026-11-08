package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/catalog"
	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/models"
	"github.com/Mireu-Lab/cloud-pricing-mcp/pkg/calculator"
)

// JSON-RPC 2.0 structures
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// Server represents the Stdio MCP server.
type Server struct {
	reader io.Reader
	writer io.Writer
}

func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{reader: r, writer: w}
}

func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.reader)
	// Allow large tokens (up to 10MB)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error: "+err.Error())
			continue
		}

		s.handleRequest(&req)
	}
	return scanner.Err()
}

func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	bytes, _ := json.Marshal(resp)
	fmt.Fprintf(s.writer, "%s\n", bytes)
}

func (s *Server) sendError(id interface{}, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	bytes, _ := json.Marshal(resp)
	fmt.Fprintf(s.writer, "%s\n", bytes)
}

func (s *Server) handleRequest(req *Request) {
	switch req.Method {
	case "initialize":
		s.sendResponse(req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "cloud-pricing-mcp",
				"version": "1.0.0",
			},
		})

	case "notifications/initialized":
		// No response needed for notification
		return

	case "tools/list":
		s.sendResponse(req.ID, map[string]interface{}{
			"tools": getToolDefinitions(),
		})

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			s.sendError(req.ID, -32602, "Invalid params")
			return
		}

		res, err := executeTool(callParams.Name, callParams.Arguments)
		if err != nil {
			s.sendResponse(req.ID, map[string]interface{}{
				"isError": true,
				"content": []map[string]interface{}{
					{"type": "text", "text": fmt.Sprintf("Error: %v", err)},
				},
			})
			return
		}

		jsonBytes, _ := json.MarshalIndent(res, "", "  ")
		s.sendResponse(req.ID, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(jsonBytes)},
			},
		})

	default:
		s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func getToolDefinitions() []Tool {
	return []Tool{
		{
			Name:        "compare_cloud_costs",
			Description: "Universal FinOps multi-cloud comparison tool (NCP, AWS, Azure, GCP) to find the most cost-effective provider for a given VM & storage spec.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"vcpu":            map[string]interface{}{"type": "number", "default": 2, "description": "Number of vCPUs"},
					"memory_gb":       map[string]interface{}{"type": "number", "default": 4, "description": "Memory in GB"},
					"storage_gb":      map[string]interface{}{"type": "number", "default": 50, "description": "Block storage size in GB"},
					"storage_type":    map[string]interface{}{"type": "string", "default": "ssd", "description": "ssd or hdd"},
					"hours_per_month": map[string]interface{}{"type": "number", "default": 730, "description": "Running hours per month (730 = 24/7)"},
					"outbound_gb":     map[string]interface{}{"type": "number", "default": 100, "description": "Outbound Internet egress bandwidth in GB"},
					"is_spot":         map[string]interface{}{"type": "boolean", "default": false, "description": "Whether to apply Spot/Preemptible discount"},
				},
			},
		},
		{
			Name:        "calculate_ncp",
			Description: "Calculate costs for NAVER CLOUD PLATFORM (Server VPC, Block Storage SSD/HDD, Outbound Internet Bandwidth) in KRW.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"server_type":  map[string]interface{}{"type": "string", "default": "s2-g2", "description": "s2-g2 (2vCPU,4GB), s4-g2 (4vCPU,8GB), c4-g2 (High CPU), m4-g2 (High Memory)"},
					"count":        map[string]interface{}{"type": "number", "default": 1, "description": "Number of server instances"},
					"hours":        map[string]interface{}{"type": "number", "default": 730, "description": "Hours per month"},
					"storage_gb":   map[string]interface{}{"type": "number", "default": 50, "description": "Block storage size in GB"},
					"storage_type": map[string]interface{}{"type": "string", "default": "ssd", "description": "ssd or hdd"},
					"outbound_gb":  map[string]interface{}{"type": "number", "default": 0, "description": "Outbound Internet egress in GB (1TB free allowance included)"},
				},
			},
		},
		{
			Name:        "calculate_cloudflare",
			Description: "Calculate costs for Cloudflare Workers (Edge Serverless) and Cloudflare R2 Object Storage (Zero Egress).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"workers_reqs_millions": map[string]interface{}{"type": "number", "default": 0, "description": "Monthly Workers invocations in millions (10M included in $5 base)"},
					"r2_storage_gb":          map[string]interface{}{"type": "number", "default": 0, "description": "R2 stored volume in GB (10GB free)"},
					"r2_class_a_millions":   map[string]interface{}{"type": "number", "default": 0, "description": "R2 Class A operations (PUT, LIST) in millions (1M free)"},
					"r2_class_b_millions":   map[string]interface{}{"type": "number", "default": 0, "description": "R2 Class B operations (GET) in millions (10M free)"},
				},
			},
		},
		{
			Name:        "calculate_aws",
			Description: "Calculate costs for Amazon Web Services (Amazon EC2, EBS gp3 SSD, and Internet Egress in Seoul ap-northeast-2).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"instance_type": map[string]interface{}{"type": "string", "default": "t4g.medium", "description": "t4g.small, t4g.medium (Graviton2), t3.medium (x86), m6i.xlarge"},
					"count":         map[string]interface{}{"type": "number", "default": 1, "description": "Number of instances"},
					"hours":         map[string]interface{}{"type": "number", "default": 730, "description": "Hours per month"},
					"ebs_gb":        map[string]interface{}{"type": "number", "default": 50, "description": "EBS gp3 SSD size in GB"},
					"outbound_gb":   map[string]interface{}{"type": "number", "default": 0, "description": "Outbound egress bandwidth in GB (100GB free)"},
					"is_spot":       map[string]interface{}{"type": "boolean", "default": false, "description": "Apply AWS Spot discount (~65%)"},
				},
			},
		},
		{
			Name:        "calculate_azure",
			Description: "Calculate costs for Microsoft Azure (Virtual Machines, Managed Disks, and Bandwidth in Korea Central).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"vm_size":     map[string]interface{}{"type": "string", "default": "Standard_B2s", "description": "Standard_B2s (2vCPU,4GB), Standard_D4s_v5 (4vCPU,16GB)"},
					"count":       map[string]interface{}{"type": "number", "default": 1, "description": "Number of VMs"},
					"hours":       map[string]interface{}{"type": "number", "default": 730, "description": "Hours per month"},
					"disk_gb":     map[string]interface{}{"type": "number", "default": 50, "description": "Managed SSD size in GB"},
					"outbound_gb": map[string]interface{}{"type": "number", "default": 0, "description": "Outbound egress bandwidth in GB (100GB free)"},
					"is_spot":     map[string]interface{}{"type": "boolean", "default": false, "description": "Apply Azure Spot discount (~80%)"},
				},
			},
		},
		{
			Name:        "calculate_gcp",
			Description: "Calculate costs for Google Cloud Platform (Compute Engine, Persistent Disk, Standard Egress in Seoul asia-northeast3).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"machine_type": map[string]interface{}{"type": "string", "default": "e2-standard-2", "description": "e2-micro, e2-standard-2 (2vCPU,8GB), e2-standard-4"},
					"count":        map[string]interface{}{"type": "number", "default": 1, "description": "Number of instances"},
					"hours":        map[string]interface{}{"type": "number", "default": 730, "description": "Hours per month"},
					"disk_gb":      map[string]interface{}{"type": "number", "default": 50, "description": "pd-balanced SSD size in GB"},
					"outbound_gb":  map[string]interface{}{"type": "number", "default": 0, "description": "Outbound egress bandwidth in GB"},
					"is_spot":      map[string]interface{}{"type": "boolean", "default": false, "description": "Apply GCP Spot discount (~70%)"},
				},
			},
		},
		{
			Name:        "list_catalog_products",
			Description: "List all supported cloud pricing catalog products across all providers or filter by provider.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"provider": map[string]interface{}{"type": "string", "description": "ncp, cloudflare, aws, azure, gcp (optional)"},
				},
			},
		},
	}
}

func executeTool(name string, args map[string]interface{}) (interface{}, error) {
	getFloat := func(k string, def float64) float64 {
		if v, ok := args[k]; ok {
			switch n := v.(type) {
			case float64:
				return n
			case int:
				return float64(n)
			}
		}
		return def
	}
	getString := func(k string, def string) string {
		if v, ok := args[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
		return def
	}
	getBool := func(k string, def bool) bool {
		if v, ok := args[k]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return def
	}

	switch name {
	case "compare_cloud_costs":
		req := models.CompareRequest{
			VCPU:          int(getFloat("vcpu", 2)),
			MemoryGB:      getFloat("memory_gb", 4),
			StorageGB:     getFloat("storage_gb", 50),
			StorageType:   getString("storage_type", "ssd"),
			HoursPerMonth: getFloat("hours_per_month", 730),
			OutboundGB:    getFloat("outbound_gb", 100),
			IsSpot:        getBool("is_spot", false),
		}
		return calculator.CompareAllClouds(req)

	case "calculate_ncp":
		return calculator.CalculateNCP(
			getString("server_type", "s2-g2"),
			int(getFloat("count", 1)),
			getFloat("hours", 730),
			getFloat("storage_gb", 50),
			getString("storage_type", "ssd"),
			getFloat("outbound_gb", 0),
		)

	case "calculate_cloudflare":
		return calculator.CalculateCloudflare(
			getFloat("workers_reqs_millions", 0),
			getFloat("r2_storage_gb", 0),
			getFloat("r2_class_a_millions", 0),
			getFloat("r2_class_b_millions", 0),
		)

	case "calculate_aws":
		return calculator.CalculateAWS(
			getString("instance_type", "t4g.medium"),
			int(getFloat("count", 1)),
			getFloat("hours", 730),
			getFloat("ebs_gb", 50),
			getFloat("outbound_gb", 0),
			getBool("is_spot", false),
		)

	case "calculate_azure":
		return calculator.CalculateAzure(
			getString("vm_size", "Standard_B2s"),
			int(getFloat("count", 1)),
			getFloat("hours", 730),
			getFloat("disk_gb", 50),
			getFloat("outbound_gb", 0),
			getBool("is_spot", false),
		)

	case "calculate_gcp":
		return calculator.CalculateGCP(
			getString("machine_type", "e2-standard-2"),
			int(getFloat("count", 1)),
			getFloat("hours", 730),
			getFloat("disk_gb", 50),
			getFloat("outbound_gb", 0),
			getBool("is_spot", false),
		)

	case "list_catalog_products":
		prov := getString("provider", "")
		if prov != "" {
			return catalog.GetProductsByProvider(models.Provider(prov)), nil
		}
		return catalog.GetAllProducts(), nil

	default:
		return nil, fmt.Errorf("unknown tool name: %s", name)
	}
}
